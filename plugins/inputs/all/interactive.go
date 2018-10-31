package all

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/parsers"
)

type Interactive struct{}

func (i *Interactive) Name() string {
	panic("implement me")
}

func (i *Interactive) Gather(deviceAgent.Accumulator) error {
	panic("implement me")
}

func (i *Interactive) SetPointMap(map[string]deviceAgent.PointDefine) {
	panic("implement me")
}

func (i *Interactive) Start() error {
	panic("implement me")
}

func (i *Interactive) Stop() error {
	panic("implement me")
}

func (i *Interactive) SetParser(parsers map[string]parsers.Parser) {
	panic("implement me")
}

