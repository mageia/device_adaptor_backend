package all

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/parsers"
)

type Interactive struct{}

func (i *Interactive) Name() string {
	panic("implement me")
}

func (i *Interactive) Gather(deviceAgent.Accumulator) error {
	panic("implement me")
}

func (i *Interactive) SetPointMap(map[string]points.PointDefine) {
	panic("implement me")
}

func (i *Interactive) Start() error {
	panic("implement me")
}

func (i *Interactive) Stop()  {
	panic("implement me")
}

func (i *Interactive) SetParser(parsers map[string]parsers.Parser) {
	panic("implement me")
}

