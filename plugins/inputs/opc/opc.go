package opc

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/inputs"
	"encoding/json"
	"net"
	"time"
)

type OPC struct {
	Address  string            `json:"address"`
	Interval internal.Duration `json:"interval"`

	client    net.Conn
	connected bool
	quality   deviceAgent.Quality
	pointMap  map[string]points.PointDefine
	pointKeys []string

	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (t *OPC) SelfCheck() deviceAgent.Quality {
	return t.quality
}

func (t *OPC) Name() string {
	if t.NameOverride != "" {
		return t.NameOverride
	}
	return t.originName
}

func (t *OPC) Gather(acc deviceAgent.Accumulator) error {
	if !t.connected {
		if e := t.Start(); e != nil {
			return e
		}
	}
	b, _ := json.Marshal(t.pointKeys)
	_, e := t.client.Write(b)
	if e != nil {
		t.connected = false
		return e
	}

	done := make(chan error, 1)
	data := make(chan []byte)

	go func() {
		buf := make([]byte, 40960)
		n, e := t.client.Read(buf)
		if e != nil {
			done <- e
			return
		}
		data <- buf[:n]
	}()

	for {
		select {
		case <-done:
		case d := <-data:
			fields := make(map[string]interface{})
			json.Unmarshal(d, &fields)
			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
			return nil
		}
	}

	return nil
}

func (t *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t.pointKeys = make([]string, len(t.pointMap))
}

func (t *OPC) Start() error {
	l, e := net.DialTimeout("tcp", t.Address, time.Second*5)
	if e != nil {
		return e
	}
	t.client = l
	t.connected = true

	i := 0
	for _, v := range t.pointMap {
		t.pointKeys[i] = v.Address
		i++
	}
	return nil
}

func (t *OPC) Stop() {
	if t.connected {
		t.client.Close()
		t.connected = false
	}
}

func init() {
	inputs.Add("opc", func() deviceAgent.Input {
		return &OPC{
			originName: "opc",
			quality:    deviceAgent.QualityGood,
		}
	})
}
