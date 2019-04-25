package opc

import (
	"bytes"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
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
	quality            device_agent.Quality
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	ctx                context.Context
	cancel             context.CancelFunc
	_baseParamList     []string
	paramList          []string
}

func (o *OPC) Name() string {
	if o.NameOverride != "" {
		return o.NameOverride
	}
	return o.originName
}

func (o *OPC) CheckGather(acc device_agent.Accumulator) error {
	if len(o.pointMap) <= 0 {
		return nil
	}
	cmd := exec.Command("opc", o.paramList...)
	outPipe, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msg("Start")
		return err
	}

	slurp, _ := ioutil.ReadAll(outPipe)

	fields := make(map[string]interface{})
	for _, r := range bytes.Split(slurp, []byte{'\n'}) {
		if len(r) > 0 {
			item := bytes.Split(r, []byte{','})
			if len(item) != 4 {
				continue
			}
			if pKey, ok := o._pointAddressToKey[string(item[0])]; ok {
				fields[pKey] = string(item[1])
			}
		}
	}
	acc.AddFields(o.NameOverride, fields, nil, o.SelfCheck())

	if err := cmd.Wait(); err != nil {
		log.Error().Err(err).Msg("Wait")
		return err
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

	var paramList = o._baseParamList
	for _, p := range o.pointMap {
		paramList = append(paramList, p.Address)
	}
	o.paramList = paramList
}

func (o *OPC) Start() error {
	address := strings.Split(o.Address, ":")
	host := "localhost"
	port := "7766"
	if len(address) == 1 {
		host = address[0]
	} else if len(address) == 2 {
		host = address[0]
		port = address[1]
	}

	o._baseParamList = []string{"-o", "csv", "-H", host, "-P", port, "-s", o.OPCServerName, "-r"}
	var paramList = o._baseParamList
	for k := range o._pointAddressToKey {
		paramList = append(paramList, k)
	}
	o.paramList = paramList

	return nil
}

func (o *OPC) Stop() {
	o.cancel()
}

func init() {
	ctx, cancel := context.WithCancel(context.Background())
	inputs.Add("opc", func() device_agent.Input {
		return &OPC{
			originName: "opc",
			quality:    device_agent.QualityGood,
			ctx:        ctx,
			cancel:     cancel,
			Timeout:    internal.Duration{Duration: time.Second * 5},
		}
	})
}
