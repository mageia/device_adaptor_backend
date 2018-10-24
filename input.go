package deviceAgent

import "math"

type PointDefine struct {
	Name      string                 `json:"name" yaml:"name"`
	Unit      string                 `json:"unit" yaml:"unit"`
	PointType PointType              `json:"point_type" yaml:"point_type"`
	Parameter float64                `json:"parameter,omitempty" yaml:"parameter"`
	Option    map[string]string      `json:"option,omitempty" yaml:"option"`
	Control   map[string]string      `json:"control,omitempty" yaml:"control"`
	Tags      []string               `json:"tags,omitempty" yaml:"tags"`
	Extra     map[string]interface{} `json:"extra,omitempty" yaml:"extra"`
}

type PointType uint8
type Quality uint8

const (
	_ PointType = iota
	PointAnalog
	PointState
)
const (
	_ Quality = iota
	QualityGood
	QualityDisconnect

	QualityUnknown = math.MaxUint8
)

type Input interface {
	Name() string
	Gather(Accumulator) error
	SelfCheck() Quality
	SetPointMap(map[string]PointDefine)
	FlushPointMap(Accumulator) error
}

type ServiceInput interface {
	Name() string
	Gather(Accumulator) error
	SetPointMap(map[string]PointDefine)
	Start() error
	Stop() error
}


