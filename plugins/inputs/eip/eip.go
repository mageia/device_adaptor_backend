package eip

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	"git.leaniot.cn/publicLib/go-eip"
	"time"
)

type EIP struct {
	Address      string            `json:"address"`
	Slot         int               `json:"slot"`
	Internal     internal.Duration `json:"internal"`
	Timeout      internal.Duration `json:"timeout"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`
	client       go_eip.Client
	connected    bool
	originName   string
	quality      device_agent.Quality
	pointMap     map[string]points.PointDefine
}

var defaultTimeout = internal.Duration{Duration: 3 * time.Second}

func (e *EIP) SelfCheck() device_agent.Quality {
	return e.quality
}

func (e *EIP) Name() string {
	if e.NameOverride != "" {
		return e.NameOverride
	}
	return e.originName
}

func (e *EIP) Gather(acc device_agent.Accumulator) error {
	if !e.connected {
		if err := e.Start(); err != nil {
			return err
		}
	}

	fields := make(map[string]interface{})
	e.quality = device_agent.QualityGood

	defer func(eip *EIP) {
		if e := recover(); e != nil {
			eip.quality = device_agent.QualityDisconnect
			eip.connected = false
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(eip.Name(), fields, nil, eip.SelfCheck())
	}(e)

	for k, point := range e.pointMap {
		v, err := e.client.Read(point.Address)
		if err != nil {
			return err
		}
		fields[k] = v
	}
	return nil
}

func (e *EIP) SetPointMap(pointMap map[string]points.PointDefine) {
	e.pointMap = pointMap
}

func (e *EIP) Start() error {
	handler := go_eip.NewTCPClientHandler(e.Address)
	handler.IdleTimeout = defaultTimeout.Duration * 100
	handler.Timeout = defaultTimeout.Duration
	if e := handler.Connect(); e != nil {
		return e
	}
	e.client = go_eip.NewClient(handler, e.Slot)
	e.connected = true
	return nil
}

func (e *EIP) Stop() {
	if e.connected {
		e.client.Stop()
		e.connected = false
	}
}

func init() {
	inputs.Add("eip", func() device_agent.Input {
		return &EIP{
			originName: "eip",
			quality:    device_agent.QualityGood,
		}
	})
}
