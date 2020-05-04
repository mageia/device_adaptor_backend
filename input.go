package device_adaptor

import (
	"device_adaptor/internal/points"
	"golang.org/x/net/context"
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
	Start() error
	Stop()
	Name() string                  //plugin name or data source name
	CheckGather(Accumulator) error //check connect or gather data
	SelfCheck() Quality            //check data quality
	SetPointMap(map[string]points.PointDefine)
}

type SimpleInput interface {
	ProbePointMap() map[string]points.PointDefine //input 尝试探测生成点表
}

type InteractiveInput interface {
	Input
}

type PassiveInput interface {
	Input
	StartListen(ctx context.Context, accumulator Accumulator) (bool, error)
	GetListening() bool
}
