package opc

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/inputs"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type OPC struct {
	Address           string            `json:"address"`
	Interval          internal.Duration `json:"interval"`
	Timeout           internal.Duration `json:"timeout"`
	OPCServerName     string            `json:"opc_server_name"`
	FieldPrefix       string            `json:"field_prefix"`
	FieldSuffix       string            `json:"field_suffix"`
	NameOverride      string            `json:"name_override"`
	originName        string
	quality           deviceAgent.Quality
	pointMap          map[string]points.PointDefine
	pointAddressToKey map[string]string
	ctx               context.Context
	cancel            context.CancelFunc
}

type opcServerResponse struct {
	Cmd     string      `json:"cmd"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func (t *OPC) sendInitMsg() error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))
	defer func() {
		l.Close()
	}()

	i := 0
	keyList := make([]string, len(t.pointMap))
	for k := range t.pointAddressToKey {
		keyList[i] = k
		i++
	}

	b, _ := json.Marshal(map[string]interface{}{
		"cmd":             "init",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
		"opc_key_list":    keyList,
	})
	_, e = l.Write(b)
	if e != nil {
		return e
	}

	buf := make([]byte, 0)
	tmpBuf := make([]byte, 1024)
	tmpResp := opcServerResponse{}

	for {
		n, e := l.Read(tmpBuf)
		if n == 0 || e != nil {
			break
		}
		buf = append(buf, tmpBuf[:n]...)
	}
	if len(buf) <= 0 {
		return nil
	}

	e = json.Unmarshal(buf, &tmpResp)
	if e != nil {
		return e
	}

	if tmpResp.Cmd == "init" && !tmpResp.Success {
		return fmt.Errorf("init failed")
	}
	l.Close()

	return nil
}
func (t *OPC) sendGetRealMsg(acc deviceAgent.Accumulator) error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))

	defer func() {
		l.Close()
	}()

	b, _ := json.Marshal(map[string]interface{}{
		"cmd":             "real_time_data",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
	})
	_, e = l.Write(b)
	if e != nil {
		return e
	}

	buf := make([]byte, 0)
	tmpBuf := make([]byte, 40960)
	tmpResp := opcServerResponse{}

	for {
		n, e := l.Read(tmpBuf)
		if n == 0 || e != nil {
			break
		}
		buf = append(buf, tmpBuf[:n]...)
	}
	if len(buf) <= 0 {
		return nil
	}

	e = json.Unmarshal(buf, &tmpResp)
	if e != nil {
		return e
	}
	fields := make(map[string]interface{})
	if tmpResp.Cmd == "real_time_data" && tmpResp.Success {
		switch r := tmpResp.Result.(type) {
		case map[string]interface{}:
			for k, v := range r {
				if pKey, ok := t.pointAddressToKey[k]; ok {
					fields[pKey] = v
				}
			}

			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
			return nil
		}
	} else {
		t.sendInitMsg()
	}

	return nil
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
	if e := t.sendGetRealMsg(acc); e != nil {
		acc.AddError(e)
		return e
	}

	return nil
}
func (t *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t.pointAddressToKey = make(map[string]string, len(t.pointMap))
	i := 0
	for _, v := range t.pointMap {
		t.pointAddressToKey[v.Address] = v.PointKey
		i++
	}
}
func (t *OPC) Start() error {
	if e := t.sendInitMsg(); e != nil {
		return e
	}

	go func() {
		ticker := time.NewTicker(time.Second * 60)
		for {
			select {
			case <-ticker.C:
				t.sendInitMsg()
			case <-t.ctx.Done():
				return
			}
		}
	}()

	return nil
}
func (t *OPC) Stop() {
	t.cancel()
}

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	inputs.Add("opc", func() deviceAgent.Input {
		return &OPC{
			originName:  "opc",
			quality:     deviceAgent.QualityGood,
			ctx:         ctx,
			cancel:      cancel,
			Timeout:     internal.Duration{Duration: time.Second * 5},
		}
	})
}
