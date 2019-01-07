package opc

import (
	"bytes"
	"context"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"time"
)

type OPC struct {
	Address            string            `json:"address"`
	Interval           internal.Duration `json:"interval"`
	Timeout            internal.Duration `json:"timeout"`
	OPCServerName      string            `json:"opc_server_name"`
	FieldPrefix        string            `json:"field_prefix"`
	FieldSuffix        string            `json:"field_suffix"`
	NameOverride       string            `json:"name_override"`
	originName         string
	quality            deviceAgent.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	ctx                context.Context
	cancel             context.CancelFunc
}

func (t *OPC) OriginName() string {
	return t.originName
}

func (t *OPC) SetValue(kv map[string]interface{}) error {
	return t.sendCommand("control", kv)
}

func (t *OPC) UpdatePointMap(map[string]interface{}) error {
	return nil
}

func (t *OPC) RetrievePointMap(keys []string) map[string]points.PointDefine {
	if len(keys) == 0 {
		return t.pointMap
	}
	result := make(map[string]points.PointDefine, len(keys))
	for _, key := range keys {
		if p, ok := t.pointMap[key]; ok {
			result[key] = p
		}
	}
	return result
}

type opcServerResponse struct {
	Cmd     string      `json:"cmd"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func (t *OPC) sendCommand(cmdId string, param interface{}) error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		log.Error().Err(e).Msg("sendInitMsg.DialTimeout")
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))
	defer l.Close()

	body := make(map[string]interface{})

	switch cmdId {
	case "init":
		i := 0
		keyList := make([]string, len(t.pointMap))
		for k := range t._pointAddressToKey {
			keyList[i] = k
			i++
		}
		body = map[string]interface{}{
			"cmd":             cmdId,
			"opc_server_host": "localhost",
			"opc_server_name": t.OPCServerName,
			"opc_key_list":    keyList,
		}
	case "real_time_data":
		body = map[string]interface{}{
			"cmd":             cmdId,
			"opc_server_host": "localhost",
			"opc_server_name": t.OPCServerName,
		}
	case "control":
		controlPairs := make([]map[string]interface{}, 0)
		pairs, ok := param.(map[string]interface{})
		if !ok {
			return errors.New("invalid control paris format")
		}

		for k, v := range pairs {
			if _, ok := t._pointAddressToKey[k]; ok {
				controlPairs = append(controlPairs, map[string]interface{}{"key": k, "value": v})
			} else {
				if pM, ok := t.pointMap[k]; ok {
					controlPairs = append(controlPairs, map[string]interface{}{"key": pM.Address, "value": v})
				}
			}
		}

		body = map[string]interface{}{
			"cmd":             "control",
			"opc_server_host": "localhost",
			"opc_server_name": t.OPCServerName,
			"control_pair":    controlPairs,
		}
	default:
		return errors.New("unsupported cmd: " + cmdId)
	}

	b, e := jsoniter.Marshal(body)
	if e != nil {
		return e
	}
	writeBuf := new(bytes.Buffer)
	binary.Write(writeBuf, binary.LittleEndian, uint32(len(b)))
	writeBuf.Write(b)
	if n, e := l.Write(writeBuf.Bytes()); e != nil || n != len(b)+4 {
		log.Error().Int("len(n)", n).Int("len(b)", len(b)).Msg("write error")
		return errors.New("write command " + cmdId + " failed")
	}

	buf := make([]byte, 0)
	tmpBuf := make([]byte, 1024)
	var packetLen int32 = -1

	for {
		n, e := io.ReadFull(l, tmpBuf)
		buf = append(buf, tmpBuf[:n]...)

		log.Debug().Int("n", n).Bytes("tmpBuf", tmpBuf[:n]).Msg("tmpBuf")
		if n == 0 || e != nil {
			break
		}
		if n > 4 && packetLen <= 0 {
			binary.Read(bytes.NewBuffer(tmpBuf[:4]), binary.LittleEndian, &packetLen)
		}
		if int32(len(buf)) >= packetLen+4 {
			break
		}
	}

	if len(buf) <= 4 {
		return errors.New("can't get response")
	}

	tmpResp := opcServerResponse{}
	e = jsoniter.Unmarshal(buf[4:], &tmpResp)
	if e != nil {
		return e
	}

	if !tmpResp.Success {
		if tmpResp.Cmd == "real_time_data" {
			go func() {
				time.Sleep(time.Second * 3)
				t.sendCommand("init", nil)
			}()
		}
		return fmt.Errorf("parse response failed")
	}

	//switch r := tmpResp.Result.(type) {
	//case map[string]interface{}:
	//	for k, v := range r {
	//		if pKey, ok := t._pointAddressToKey[k]; ok {
	//			fields[pKey] = v
	//		}
	//	}
	//
	//	acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
	//	return nil
	//}

	log.Debug().Interface("tmpResp", tmpResp).Msg("sendInitMsg")

	switch tmpResp.Cmd {
	case "init":
	case "control":
	case "real_time_data":
		fields := make(map[string]interface{})
		acc, ok := param.(deviceAgent.Accumulator)
		if !ok {
			return errors.New("invalid real_time_data acc format")
		}

		switch r := tmpResp.Result.(type) {
		case map[string]interface{}:
			for k, v := range r {
				if pKey, ok := t._pointAddressToKey[k]; ok {
					fields[pKey] = v
				}
			}

			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
			return nil
		}
	default:
		return fmt.Errorf("parse response failed")
	}

	return nil
}

func (t *OPC) sendInitMsg1() error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		log.Error().Err(e).Msg("sendInitMsg.DialTimeout")
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))

	i := 0
	keyList := make([]string, len(t.pointMap))
	for k := range t._pointAddressToKey {
		keyList[i] = k
		i++
	}
	b, _ := jsoniter.Marshal(map[string]interface{}{
		"cmd":             "init",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
		"opc_key_list":    keyList,
	})
	b = append(b, 0x0a)
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
		if tmpBuf[n-1] == 0x0a {
			buf = append(buf, tmpBuf[:n-1]...)
			break
		}
		buf = append(buf, tmpBuf[:n]...)
	}
	l.Close()

	if len(buf) <= 0 {
		return nil
	}

	e = jsoniter.Unmarshal(buf, &tmpResp)
	if e != nil {
		return e
	}

	if tmpResp.Cmd == "init" && !tmpResp.Success {
		return fmt.Errorf("init failed")
	}
	log.Debug().Interface("tmpResp", tmpResp).Msg("sendInitMsg")

	return nil
}
func (t *OPC) sendGetRealMsg1(acc deviceAgent.Accumulator) error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))

	b, _ := jsoniter.Marshal(map[string]interface{}{
		"cmd":             "real_time_data",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
	})
	b = append(b, 0x0a)
	_, e = l.Write(b)
	if e != nil {
		return e
	}

	buf := make([]byte, 0)
	tmpBuf := make([]byte, 4096)
	tmpResp := opcServerResponse{}

	for {
		n, e := l.Read(tmpBuf)
		if n == 0 || e != nil {
			break
		} else if tmpBuf[n-1] == 0x0a {
			buf = append(buf, tmpBuf[:n-1]...)
			break
		}
		buf = append(buf, tmpBuf[:n]...)
	}
	l.Close()

	if len(buf) <= 0 {
		return nil
	}

	e = jsoniter.Unmarshal(buf, &tmpResp)
	if e != nil {
		log.Error().Err(e).Msg("Unmarshal")
		return e
	}
	fields := make(map[string]interface{})
	if tmpResp.Cmd == "real_time_data" && tmpResp.Success {
		switch r := tmpResp.Result.(type) {
		case map[string]interface{}:
			for k, v := range r {
				if pKey, ok := t._pointAddressToKey[k]; ok {
					fields[pKey] = v
				}
			}

			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
			return nil
		}
	} else {
		log.Debug().Msg("Resend init message")

		go func() {
			time.Sleep(time.Second * 3)
			t.sendInitMsg1()
		}()
	}

	return nil
}
func (t *OPC) sendControlMsg1(pairs map[string]interface{}) error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		return e
	}
	l.SetReadDeadline(time.Now().Add(t.Timeout.Duration * 2))

	controlPairs := make([]map[string]interface{}, 0)

	for k, v := range pairs {
		if _, ok := t._pointAddressToKey[k]; ok {
			controlPairs = append(controlPairs, map[string]interface{}{"key": k, "value": v})
		} else {
			if pM, ok := t.pointMap[k]; ok {
				controlPairs = append(controlPairs, map[string]interface{}{"key": pM.Address, "value": v})
			}
		}
	}

	b, _ := jsoniter.Marshal(map[string]interface{}{
		"cmd":             "control",
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
		"control_pair":    controlPairs,
	})
	b = append(b, 0x0a)
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
		if tmpBuf[n-1] == 0x0a {
			buf = append(buf, tmpBuf[:n-1]...)
			break
		}
	}
	l.Close()
	if len(buf) <= 0 {
		return nil
	}

	e = jsoniter.Unmarshal(buf, &tmpResp)
	if e != nil {
		return e
	}

	log.Debug().Interface("rmpResp", tmpResp).Msg("Control")

	if tmpResp.Cmd != "control" || !tmpResp.Success {
		return fmt.Errorf("%v", tmpResp.Result)
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
	if e := t.sendCommand("real_time_data", acc); e != nil {
		acc.AddError(e)
		return e
	}

	return nil
}
func (t *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t._pointAddressToKey = make(map[string]string, len(t.pointMap))
	i := 0
	for _, v := range t.pointMap {
		t._pointAddressToKey[v.Address] = v.PointKey
		i++
	}
}
func (t *OPC) Start() error {
	if e := t.sendCommand("init", nil); e != nil {
		return e
	}

	go func() {
		ticker := time.NewTicker(time.Second * 60)
		for {
			select {
			case <-ticker.C:
				t.sendCommand("init", nil)
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
			originName: "opc",
			quality:    deviceAgent.QualityGood,
			ctx:        ctx,
			cancel:     cancel,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
