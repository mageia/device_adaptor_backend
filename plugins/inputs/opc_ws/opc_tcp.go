package opc_ws

import (
	"bufio"
	"context"
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

type OpcTcp struct {
	Address            string            `json:"address"`
	Interval           internal.Duration `json:"interval"`
	Timeout            internal.Duration `json:"timeout"`
	OPCServerName      string            `json:"opc_server_name"`
	FieldPrefix        string            `json:"field_prefix"`
	FieldSuffix        string            `json:"field_suffix"`
	NameOverride       string            `json:"name_override"`
	originName         string
	quality            device_adaptor.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	ctx                context.Context
	cancel             context.CancelFunc
}

func (t *OpcTcp) OriginName() string {
	return t.originName
}

func (t *OpcTcp) SetValue(kv map[string]interface{}) error {
	return t.sendCommand("control", kv)
}

func (t *OpcTcp) UpdatePointMap(map[string]interface{}) error {
	return nil
}

func (t *OpcTcp) RetrievePointMap(keys []string) map[string]points.PointDefine {
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

//
//func (t *OpcTcp) receiveCommand() {
//	var totalLen uint32
//
//	for {
//		binary.Read(t.client, binary.LittleEndian, &totalLen)
//		if totalLen <= 4 {
//			//return errors.New("can't get response")
//		}
//		buf := make([]byte, totalLen-4)
//		if _, err := io.ReadFull(t.client, buf); err != nil {
//			//return err
//		}
//
//		log.Debug().Str("buf", string(buf)).Msg("buf")
//	}
//}

func (t *OpcTcp) sendCommand(cmdId string, param interface{}) error {
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
	b = append(b, byte('\n'))
	if n, e := l.Write(b); e != nil || n != len(b) {
		log.Error().Int("len(n)", n).Int("len(b)", len(b)).Msg("write error")
		return errors.New("write command " + cmdId + " failed")
	}

	r := bufio.NewReader(l)
	buf, _ := r.ReadBytes(byte('\n'))
	log.Debug().Str("buf", string(buf)).Msg("buf")

	tmpResp := opcServerResponse{}
	if e := jsoniter.Unmarshal(buf, &tmpResp); e != nil {
		return e
	}

	log.Debug().Interface("tmpResp", tmpResp).Msg("tmpResp")

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

func (t *OpcTcp) SelfCheck() device_adaptor.Quality {
	return t.quality
}
func (t *OpcTcp) Name() string {
	if t.NameOverride != "" {
		return t.NameOverride
	}
	return t.originName
}
func (t *OpcTcp) CheckGather(acc device_adaptor.Accumulator) error {
	if e := t.sendCommand("real_time_data", acc); e != nil {
		t.Stop()
		return e
	}

	return nil
}
func (t *OpcTcp) SetPointMap(pointMap map[string]points.PointDefine) {
	t.pointMap = pointMap
	t._pointAddressToKey = make(map[string]string, len(t.pointMap))
	for k, v := range t.pointMap {
		t._pointAddressToKey[v.Address] = k
	}
}
func (t *OpcTcp) Start() error {
	go func() {
		ticker := time.NewTicker(time.Second * 2)
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
func (t *OpcTcp) Stop() {
	t.cancel()
}

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	inputs.Add("opc_tcp", func() device_adaptor.Input {
		return &OpcTcp{
			originName: "opc_tcp",
			quality:    device_adaptor.QualityGood,
			ctx:        ctx,
			cancel:     cancel,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
