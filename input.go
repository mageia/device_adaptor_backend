package deviceAgent

import "math"

type PointDefine struct {
	Name      string                 `json:"name"`
	Unit      string                 `json:"unit"`
	Parameter float64                `json:"parameter,omitempty"`
	Option    map[string]string      `json:"option,omitempty"`
	Control   map[string]string      `json:"control,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type Quality uint8

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
	FlushPointMap(Accumulator) error

	Start() error
	Stop() error
}
