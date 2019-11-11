package opc

import (
	"bufio"
	"context"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type OpcTcp struct {
	Address       string            `json:"address"`
	Interval      internal.Duration `json:"interval"`
	Timeout       internal.Duration `json:"timeout"`
	OPCServerName string            `json:"opc_server_name"`
	FieldPrefix   string            `json:"field_prefix"`
	FieldSuffix   string            `json:"field_suffix"`
	NameOverride  string            `json:"name_override"`

	client             net.Conn
	connected          bool
	reader             *bufio.Reader
	writer             *bufio.Writer
	originName         string
	quality            device_adaptor.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
}

func (t *OpcTcp) Name() string {
	if t.NameOverride != "" {
		return t.NameOverride
	}
	return t.originName
}

func (t *OpcTcp) CheckGather(device_adaptor.Accumulator) error {
	return nil
}

func (t *OpcTcp) SelfCheck() device_adaptor.Quality {
	return t.quality
}

func (t *OpcTcp) SetPointMap(map[string]points.PointDefine) {

}

func (t *OpcTcp) Connect() error {
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

func (t *OpcTcp) DisConnect() error {
	if t.connected {
		t.connected = false
		return t.client.Close()
	}
	return nil
}

func (t *OpcTcp) Listen(context.Context, device_adaptor.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("opc_tcp", func() device_adaptor.Input {
		return &OPC{
			originName: "opc_tcp",
			quality:    device_adaptor.QualityGood,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
