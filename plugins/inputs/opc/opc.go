package opc

import (
	"bufio"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/utils"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
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

	client             net.Conn
	connected          bool
	reader             *bufio.Reader
	writer             *bufio.Writer
	originName         string
	quality            device_adaptor.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
}
type opcServerResponse struct {
	Cmd     string      `json:"cmd"`
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

func (t *OPC) sendCommand(cmdId string, param interface{}) error {
	if !t.connected {
		return errors.New("not connected")
	}

	body := map[string]interface{}{
		"cmd":             cmdId,
		"opc_server_host": "localhost",
		"opc_server_name": t.OPCServerName,
	}

	switch cmdId {
	case "init":
		i := 0
		keyList := make([]string, len(t.pointMap))
		for k := range t._pointAddressToKey {
			keyList[i] = k
			i++
		}
		body["params"] = keyList
	case "real_time_data":

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

		body["params"] = controlPairs
	default:
		return errors.New("unsupported cmd: " + cmdId)
	}

	b, e := jsoniter.Marshal(body)
	if e != nil {
		return e
	}
	b = append(b, byte('\n'))
	if _, e := t.writer.Write(b); e != nil {
		return e
	}
	t.writer.Flush()

	buf, _ := t.reader.ReadBytes(byte('\n'))
	tmpResp := opcServerResponse{}
	if e := jsoniter.Unmarshal(buf, &tmpResp); e != nil {
		return e
	}

	if !tmpResp.Success {
		if tmpResp.Cmd == "real_time_data" {
			go t.sendCommand("init", nil)
		}
		return fmt.Errorf("parse response failed, success == false")
	}

	switch tmpResp.Cmd {
	case "init":
	case "control":
	case "real_time_data":
		fields := make(map[string]interface{})
		acc, ok := param.(device_adaptor.Accumulator)
		if !ok {
			return errors.New("invalid real_time_data acc format")
		}
		if r, ok := tmpResp.Result.(map[string]interface{}); ok {
			for k, v := range r {
				if pKey, ok := t._pointAddressToKey[k]; ok {
					switch vf := v.(type) {
					case float64: //TODO: float32
						fields[pKey] = utils.Round(vf, 6)
					case bool:
						if vf {
							fields[pKey] = 1
						} else {
							fields[pKey] = 0
						}
					default:
						fields[pKey] = v
					}
				}
			}
			acc.AddFields(t.NameOverride, fields, nil, t.SelfCheck())
		}
	default:
		return fmt.Errorf("parse response failed")
	}

	return nil
}
func (t *OPC) OriginName() string {
	return t.originName
}
func (t *OPC) SetValue(kv map[string]interface{}) error {
	log.Debug().Interface("kv", kv).Msg("SetValue")
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
func (t *OPC) SelfCheck() device_adaptor.Quality {
	return t.quality
}
func (t *OPC) Name() string {
	if t.NameOverride != "" {
		return t.NameOverride
	}
	return t.originName
}
func (t *OPC) CheckGather(acc device_adaptor.Accumulator) error {
	if !t.connected {
		if e := t.Start(); e != nil {
			return e
		}
	}
	if e := t.sendCommand("real_time_data", acc); e != nil {
		t.Stop()
		return e
	}

	return nil
}
func (t *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t._pointAddressToKey = make(map[string]string, len(t.pointMap))
	for k, v := range t.pointMap {
		t._pointAddressToKey[v.Address] = k
	}
}
func (t *OPC) Start() error {
	l, e := net.DialTimeout("tcp", t.Address, t.Timeout.Duration)
	if e != nil {
		log.Error().Err(e).Msg("sendInitMsg.DialTimeout")
		return e
	}

	t.client = l
	t.connected = true

	t.reader = bufio.NewReader(l)
	t.writer = bufio.NewWriter(l)

	if e := t.sendCommand("init", nil); e != nil {
		t.Stop()
		return e
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
	inputs.Add("opc_tcp", func() device_adaptor.Input {
		return &OPC{
			originName: "opc_tcp",
			quality:    device_adaptor.QualityGood,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
