package opc

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/inputs"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type OPC struct {
	Address       string            `json:"address"`
	Interval      internal.Duration `json:"interval"`
	Timeout       internal.Duration `json:"timeout"`
	OPCServerName string            `json:"opc_server_name"`
	FieldPrefix   string            `json:"field_prefix"`
	FieldSuffix   string            `json:"field_suffix"`
	NameOverride  string            `json:"name_override"`
	originName    string
	client        net.Conn
	connected     bool
	exit          bool
	quality       deviceAgent.Quality
	pointMap      map[string]points.PointDefine
	pointKeys     []string
	receiveData   chan []byte
}

type opcServerResponse struct {
	Cmd     string      `json:"cmd"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
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

	defer func() {
		t.client.Close()
		t.connected = false
	}()

	if e := t.sendGetRealMsg(t.client); e != nil {
		acc.AddError(e)
		return e
	}

	buf := make([]byte, 4096)
	tmpResp := opcServerResponse{}
	n, e := t.client.Read(buf)
	if e != nil {
		log.Println(e)
		return e
	}
	e = json.Unmarshal(buf[:n], &tmpResp)
	if e != nil {
		return e
	}
	fields := make(map[string]interface{})
	if tmpResp.Cmd == "real_time_data" && tmpResp.Success {
		switch r := tmpResp.Result.(type) {
		case map[string]interface{}:
			for k, v := range r {
				fields[k] = v
			}
			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
			return nil
		}
	}

	return nil
}

func (t *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t.pointKeys = make([]string, len(t.pointMap))
	i := 0
	for _, v := range t.pointMap {
		t.pointKeys[i] = v.Address
		i++
	}
}

func (t *OPC) sendInitMsg(c net.Conn) error {
	b, _ := json.Marshal(map[string]interface{}{
		"cmd":             "init",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
		"opc_key_list":    t.pointKeys,
	})
	_, e := c.Write(b)
	if e != nil {
		return e
	}
	return nil
}

func (t *OPC) sendGetRealMsg(c net.Conn) error {
	b, _ := json.Marshal(map[string]interface{}{
		"cmd":             "real_time_data",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
	})
	_, e := c.Write(b)
	if e != nil {
		return e
	}
	return nil
}

func (t *OPC) receiveResponse() error {
	buf := make([]byte, 40960)
	tmpResp := opcServerResponse{}

	defer func() {
		t.client.Close()
		t.connected = false
		log.Println("defer")
	}()

	for {
		n, e := t.client.Read(buf)
		if e != nil {
			log.Println(e)
			return e
		}
		e = json.Unmarshal(buf[:n], &tmpResp)
		if e != nil {
			return e
		}

		if tmpResp.Cmd == "init" {
			if !tmpResp.Success {
				return fmt.Errorf("init failed")
			}
			log.Println("init success", tmpResp)
			continue
		}

		t.receiveData <- buf[:n]
	}
}

func (t *OPC) Start() error {
	l, e := net.DialTimeout("udp", t.Address, t.Timeout.Duration)
	if e != nil {
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))

	if e := t.sendInitMsg(l); e != nil {
		l.Close()
		return e
	}

	buf := make([]byte, 4096)
	tmpResp := opcServerResponse{}
	n, e := l.Read(buf)
	if e != nil {
		log.Println(e)
		return e
	}
	e = json.Unmarshal(buf[:n], &tmpResp)
	if e != nil {
		return e
	}

	if tmpResp.Cmd == "init" && !tmpResp.Success {
		return fmt.Errorf("init failed")
	}

	t.client = l
	t.connected = true

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
			originName:  "opc",
			quality:     deviceAgent.QualityGood,
			receiveData: make(chan []byte),

			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
