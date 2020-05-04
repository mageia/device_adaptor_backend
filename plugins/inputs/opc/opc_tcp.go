package opc

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"encoding/base64"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net"
	"time"
)

type Opc struct {
	Address      string            `json:"address"`
	Interval     internal.Duration `json:"interval"`
	Timeout      internal.Duration `json:"timeout"`
	EnableGzip   bool              `json:"enable_gzip"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`

	client             net.Conn
	connected          bool
	reader             *bufio.Reader
	writer             *bufio.Writer
	originName         string
	quality            device_adaptor.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	listening          bool
}

type QVT struct {
	Q string      `json:"q"`
	T int64       `json:"ts"`
	V interface{} `json:"v"`
}
type Response struct {
	DataSetType string
	Timestamp   int64
	Data        map[string]QVT
}

func (t *Opc) Start() error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		log.Error().Err(e).Msg("Connect.DialTimeout")
		return e
	}
	t.client = l
	t.connected = true

	t.reader = bufio.NewReader(l)
	t.writer = bufio.NewWriter(l)
	return nil
}

func (t *Opc) Stop() {
	if t.connected {
		t.connected = false
		t.client.Close()
	}
}

func (t *Opc) StartListen(ctx context.Context, acc device_adaptor.Accumulator) (bool, error) {
	if t.listening {
		return true, nil
	}
	t.listening = true
	for {
		select {
		case <-ctx.Done():
			t.listening = false
			return false, nil
		default:
			out := Response{}
			d, e := t.reader.ReadBytes('\n')
			if e != nil {
				log.Error().Err(e).Msg("ReadBytes")
				t.listening = false
				t.Stop()
				return false, e
			}

			if t.EnableGzip {
				r, e := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, bytes.NewReader(d)))
				if e != nil {
					log.Error().Err(e).Msg("gzip")
					continue
				}

				dd, e := ioutil.ReadAll(r)
				if e != nil {
					log.Error().Err(e).Msg("ReadAll")
					continue
				}
				if e := jsoniter.Unmarshal(dd, &out); e != nil {
					log.Error().Err(e).Msg("Unmarshal")
					continue
				}
			} else {
				if e := jsoniter.Unmarshal(d, &out); e != nil {
					log.Error().Err(e).Msg("Unmarshal")
					continue
				}
			}

			fields := make(map[string]interface{})
			for k, v := range out.Data {
				fields[k] = v.V
			}
			//log.Debug().Interface("out", out).Msg("End")
			acc.AddFields(t.Name(), fields, map[string]string{}, t.SelfCheck())
		}
	}
}

func (t *Opc) GetListening() bool {
	return t.listening
}

func (t *Opc) Name() string {
	if t.NameOverride != "" {
		return t.NameOverride
	}
	return t.originName
}

func (t *Opc) CheckGather(device_adaptor.Accumulator) error {
	if !t.connected {
		t.Start()
	}
	return nil
}

func (t *Opc) SelfCheck() device_adaptor.Quality {
	return t.quality
}

func (t *Opc) SetPointMap(map[string]points.PointDefine) {

}

func init() {
	inputs.Add("opc_tcp", func() device_adaptor.Input {
		return &Opc{
			originName: "opc_tcp",
			quality:    device_adaptor.QualityGood,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
