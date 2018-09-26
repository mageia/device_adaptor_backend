package ftp

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/parsers"
	"encoding/csv"
	"github.com/axgle/mahonia"
	"github.com/jlaffaye/ftp"
	"log"
	"sync"
	"time"
)

type FTP struct {
	Address  string
	Username string
	Password string
	DevPath  string
	DataPath string

	connected bool
	done      chan struct{}
	client    *ftp.ServerConn
	pointMap  map[string]deviceAgent.PointDefine
	parser    parsers.Parser
}

func (*FTP) SampleConfig() string {
	return ""
}

func (*FTP) Description() string {
	return ""
}
func (f *FTP) SetParser(parser parsers.Parser) {
	f.parser = parser
}
func (f *FTP) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	f.pointMap = pointMap
}

func (*FTP) FlushPointMap(acc deviceAgent.Accumulator) error {
	return nil
}

func (f *FTP) gatherServer(client *ftp.ServerConn, acc deviceAgent.Accumulator) error {
	decoder := mahonia.NewDecoder("gbk")
	//encode := mahonia.NewEncoder("gbk")

	rDev, e := f.client.Retr(decoder.ConvertString(f.DevPath))
	if e != nil {
		log.Printf("Fetch dev file failed: %v", e)
		return e
	}
	devReader := csv.NewReader(decoder.NewReader(rDev))
	devReader.FieldsPerRecord = -1
	devReader.TrimLeadingSpace = true
	//devMap := make(map[string]interface{})
	return nil
}

func (f *FTP) Gather(acc deviceAgent.Accumulator) error {
	if !f.connected {
		if e := f.connect(); e != nil {
			return e
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(client *ftp.ServerConn) {
		defer wg.Done()
		acc.AddError(f.gatherServer(f.client, acc))
	}(f.client)
	wg.Wait()

	return nil
}

func (f *FTP) Start(acc deviceAgent.Accumulator) error {
	f.done = make(chan struct{})
	f.connected = false
	return f.connect()
}

func (f *FTP) connect() error {
	if c, e := ftp.DialTimeout(f.Address, time.Second*5); e != nil {
		log.Printf("Connect to %s failed: %v\n", f.Address, e)
		return e
	} else {
		e := c.Login(f.Username, f.Password)
		if e != nil {
			log.Printf("Login to %s failed: %v\n", f.Address, e)
			return e
		}
		f.client = c
		f.connected = true
	}
	return nil
}

func init() {
	inputs.Add("ftp", func() deviceAgent.Input {
		return &FTP{}
	})
}
