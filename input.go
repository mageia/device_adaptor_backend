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
	Name() string						//plugin name or data source name
	CheckGather(Accumulator) error		//check connect or gather data
	SelfCheck() Quality					//check data quality
	SetPointMap(map[string]points.PointDefine)
}

type InteractiveInput interface {
	Input
	Start() error
	Stop()
}

type PassiveInput interface {
	Input
	Connect() error
	DisConnect() error
	Listen(context.Context, Accumulator) error	//listen forever to obtain data from data source
}
