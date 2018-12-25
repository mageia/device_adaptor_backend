package deviceAgent

import (
	"device_adaptor/internal/points"
	"math"
)

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
	SetPointMap(map[string]points.PointDefine)
	//FlushPointMap(Accumulator) error
}

type InteractiveInput interface {
	Name() string
	Gather(Accumulator) error
	SetPointMap(map[string]points.PointDefine)
	Start() error
	Stop()
}
