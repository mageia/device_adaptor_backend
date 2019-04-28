package opc

import (
	"bufio"
	"bytes"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type OPC struct {
	Address            string            `json:"address"`
	Interval           internal.Duration `json:"interval"`
	OPCServerName      string            `json:"opc_server_name"`
	FieldPrefix        string            `json:"field_prefix"`
	FieldSuffix        string            `json:"field_suffix"`
	NameOverride       string            `json:"name_override"`
	originName         string
	quality            device_agent.Quality
	pointMap           map[string]points.PointDefine
	connected          bool
	_opcCmd            *exec.Cmd
	_baseParamList     []string
	_cmdReader         *bufio.Reader
	_fields            map[string]interface{}
	_pointAddressToKey map[string]string
}

func (o *OPC) Name() string {
	if o.NameOverride != "" {
		return o.NameOverride
	}
	return o.originName
}

func (o *OPC) CheckGather(acc device_agent.Accumulator) error {
	if len(o._fields) > 0 {
		acc.AddFields("opc", o._fields, nil, o.quality)
	}

	return nil
}

func (o *OPC) SelfCheck() device_agent.Quality {
	return o.quality
}

func (o *OPC) SetPointMap(pointMap map[string]points.PointDefine) {
	o.pointMap = pointMap
	o._pointAddressToKey = make(map[string]string, len(o.pointMap))
	for k, v := range o.pointMap {
		o._pointAddressToKey[v.Address] = k
	}
}

func (o *OPC) Listen(ctx context.Context, acc device_agent.Accumulator) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			l, _, e := o._cmdReader.ReadLine()
			if e != nil {
				log.Error().Err(e).Msg("opc ReadLine")
				o.DisConnect()
				o.Connect()
				break
			}

			if len(l) > 0 {
				item := bytes.Split(l, []byte{','})
				if len(item) != 4 {
					break
				}
				if pKey, ok := o._pointAddressToKey[string(item[0])]; ok {
					o._fields[o.FieldPrefix+pKey+o.FieldSuffix] = string(item[1])
				}
			}
		}
	}
}

func (o *OPC) Connect() error {
	address := strings.Split(o.Address, ":")
	host := "localhost"
	port := "7766"
	if len(address) == 1 {
		host = address[0]
	} else if len(address) == 2 {
		host = address[0]
		port = address[1]
	}

	loop := fmt.Sprintf("%.1f", float32((o.Interval.Duration)/time.Second))
	o._baseParamList = []string{"-o", "csv", "-H", host, "-P", port, "-s", o.OPCServerName, "-L", loop, "-r"}
	var paramList = o._baseParamList
	for k := range o._pointAddressToKey {
		paramList = append(paramList, k)
	}

	o._opcCmd = exec.Command("opc", paramList...)
	stdoutPipe, e := o._opcCmd.StdoutPipe()
	if e != nil {
		log.Error().Err(e).Msg("Assign stdoutPipe")
		return e
	}
	o._cmdReader = bufio.NewReader(stdoutPipe)

	if e := o._opcCmd.Start(); e != nil {
		log.Error().Err(e).Msg("Start CMD")
		return e
	}
	o.connected = true

	return nil
}

func (o *OPC) DisConnect() error {
	if e := o._opcCmd.Process.Kill(); e != nil {
		log.Error().Err(e).Msg("Process.Kill")
		return e
	}
	o.connected = false

	return nil
}

func init() {
	inputs.Add("opc", func() device_agent.Input {
		return &OPC{
			originName: "opc",
			_fields:    make(map[string]interface{}),
			quality:    device_agent.QualityGood,
			Interval:   internal.Duration{Duration: time.Second * 3},
		}
	})
}
