package device_agent

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
}

type InteractiveInput interface {
	Input
	Start() error
	Stop()
}
