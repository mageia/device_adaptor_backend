package ftp

import (
	"device_adaptor"
	"device_adaptor/agent"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/parsers"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/jlaffaye/ftp"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"path"
	"runtime/debug"
	"sync"
	"time"
)

type FTP struct {
	Address     string         `json:"address"`
	PointPath   string         `json:"point_path"`
	PointDecode string         `json:"point_decode"`
	DataPath    string         `json:"data_path"`
	DataDecode  string         `json:"data_decode"`
	PointParser parsers.Parser `json:"point_parser"`
	DataParser  parsers.Parser `json:"data_parser"`

	originName        string
	connected         bool
	client            *ftp.ServerConn
	quality           device_agent.Quality
	basePath          string
	pointMap          map[string]points.PointDefine
	pointAddressToKey map[string]string
	FieldPrefix       string `json:"field_prefix"`
	FieldSuffix       string `json:"field_suffix"`
	NameOverride      string `json:"name_override"`
}

func (f *FTP) SelfCheck() device_agent.Quality {
	return f.quality
}

func (f *FTP) Name() string {
	if f.NameOverride != "" {
		return f.NameOverride
	}
	return f.originName
}
func (f *FTP) SetParser(parser map[string]parsers.Parser) {
}
func (f *FTP) SetPointMap(pointMap map[string]points.PointDefine) {
	f.pointMap = pointMap
	for k, v := range pointMap {
		f.pointAddressToKey[v.Address] = k
	}
}

func (*FTP) FlushPointMap(acc device_agent.Accumulator) error {
	return nil
}
func (f *FTP) gatherServer(client *ftp.ServerConn, acc device_agent.Accumulator) error {
	if f.DataPath == "" {
		return errors.New("empty data_path")
	}
	if f.DataDecode == "" {
		f.DataDecode = "utf-8"
	}
	fields := make(map[string]interface{})

	defer func(ftp *FTP) {
		if e := recover(); e != nil {
			debug.PrintStack()
			ftp.quality = device_agent.QualityDisconnect
			ftp.connected = false
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(ftp.Name(), fields, nil, ftp.SelfCheck())
	}(f)

	rData, e := f.client.Retr(mahonia.NewEncoder(f.DataDecode).ConvertString(path.Join(f.basePath, f.DataPath)))
	if e != nil {
		log.Error().Err(e).Str("code", f.DataDecode).Str("data_path", path.Join(f.basePath, f.DataPath)).Msg("Retrieve")
		return e
	}
	defer rData.Close()

	//TODO: csv parser
	dataReader := csv.NewReader(mahonia.NewDecoder(f.DataDecode).NewReader(rData))
	dataReader.FieldsPerRecord = -1
	dataReader.TrimLeadingSpace = true

	for {
		r, e := dataReader.Read()
		if e == io.EOF {
			break
		} else if e != nil {
			log.Error().Err(e).Msg("dataReader.Read")
			return e
		}
		if a, ok := f.pointAddressToKey[r[0]]; ok {
			fields[a] = r[1]
		}
	}

	return nil
}

func (f *FTP) Gather(acc device_agent.Accumulator) error {
	if !f.connected {
		if e := f.connect(); e != nil {
			return e
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(client *ftp.ServerConn) {
		defer wg.Done()
		if e := f.gatherServer(f.client, acc); e != nil {
			acc.AddError(e)
			f.Stop()
		}
	}(f.client)
	wg.Wait()

	return nil
}

func (f *FTP) Start() error {
	return f.connect()
}
func (f *FTP) Stop() {
	if f.connected {
		f.client.Quit()
		f.connected = false
	}
}

func (f *FTP) connect() error {
	_url, e := url.Parse(f.Address)
	if e != nil || _url.Scheme != "ftp" {
		return e
	}
	if _url.Port() == "" {
		_url.Host = fmt.Sprintf("%s:21", _url.Host)
	}
	c, e := ftp.DialTimeout(_url.Host, time.Second*5)
	if e != nil {
		log.Error().Err(e).Str("host", _url.Host).Msg("DialTimeout")
		return e
	}

	_password, _ := _url.User.Password()
	e = c.Login(_url.User.Username(), _password)
	if e != nil {
		log.Error().Err(e).Str("host", _url.Host).Msg("Login")
		return e
	}

	if f.PointDecode == "" {
		f.PointDecode = "utf-8"
	}

	//校验并保存Path
	if _url.Path != "" {
		f.basePath = _url.Path
		if e := c.ChangeDir(mahonia.NewEncoder(f.PointDecode).ConvertString(_url.Path)); e != nil {
			log.Error().Err(e).Str("path", _url.Path).Msg("ChangeDir")
			c.Logout()
			return e
		}
	}

	//解析并保存点表
	if f.PointPath != "" && len(f.pointMap) == 0 {
		rDev, e := c.Retr(mahonia.NewEncoder(f.PointDecode).ConvertString(path.Join(f.basePath, f.PointPath)))
		if e != nil {
			log.Error().Err(e).Str("pointPath", path.Join(f.basePath, f.PointPath)).Msg("Retrieve")
			return e
		}
		defer rDev.Close()

		//TODO: csv parser
		devReader := csv.NewReader(mahonia.NewDecoder(f.PointDecode).NewReader(rDev))
		devReader.FieldsPerRecord = -1
		devReader.TrimLeadingSpace = true

		for {
			r, e := devReader.Read()
			if e == io.EOF {
				break
			} else if e != nil {
				log.Error().Err(e).Msg("devReader.Read")
				return e
			}
			f.pointMap[r[0]] = points.PointDefine{Label: r[0], Name: r[1], Address: r[0]}
		}

		for k, v := range f.pointMap {
			f.pointAddressToKey[v.Address] = k
		}

		go func() {
			timeS := time.Now()
			begin := points.SqliteDB.Begin()
			if r := begin.Where("input_name = ?", f.Name()).Delete(points.PointDefine{}); r.Error != nil {
				r.Rollback()
				log.Error().Err(r.Error).Msg("Delete")
				return
			}

			for k, v := range f.pointMap {
				v.InputName = f.Name()
				if v.Name == "" {
					v.Name = k
				}
				v.PointKey = k
				if r := begin.Assign(v).FirstOrCreate(&v, "input_name = ? AND point_key = ?", f.Name(), v.PointKey); r.Error != nil {
					r.Rollback()
					log.Error().Err(r.Error).Msg("FirstOrCreate")
					return
				}
			}
			if r := begin.Commit(); r.Error != nil {
				r.Rollback()
				log.Error().Err(r.Error).Msg("UpdatePointMap")
				return
			}
			log.Debug().Str("TimeSince", time.Since(timeS).String()).Msg("ftp.UpdatePointMap")
			agent.Signal <- agent.PointDefineUpdateSignal{Input: f}
		}()
	}

	f.client = c
	f.connected = true
	return nil
}

func init() {
	inputs.Add("ftp", func() device_agent.Input {
		return &FTP{
			pointMap:          make(map[string]points.PointDefine, 0),
			pointAddressToKey: make(map[string]string),
			quality:           device_agent.QualityGood,
		}
	})
}
