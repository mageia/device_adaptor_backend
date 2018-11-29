package eip

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/inputs"
)

type EIP struct {
	Address      string            `json:"address"`
	Internal     internal.Duration `json:"internal"`
	Timeout      internal.Duration `json:"timeout"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`
	originName   string
	quality      deviceAgent.Quality
	pointMap     map[string]points.PointDefine
}

func (e *EIP) SelfCheck() deviceAgent.Quality {
	return e.quality
}

func (e *EIP) Name() string {
	if e.NameOverride != "" {
		return e.NameOverride
	}
	return e.originName
}

func (e *EIP) Gather(deviceAgent.Accumulator) error {
	return nil
}

func (e *EIP) SetPointMap(pointMap map[string]points.PointDefine) {
	e.pointMap = pointMap
}

func (e *EIP) Start() error {
	return nil
}

func (e *EIP) Stop() {
	return
}

func init() {
	inputs.Add("eip", func() deviceAgent.Input {
		return &EIP{
			originName: "eip",
			quality:    deviceAgent.QualityGood,
		}
	})
}
