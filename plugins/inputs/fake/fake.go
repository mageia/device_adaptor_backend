package fake

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/inputs"
)

type Fake struct {
}

func (f *Fake) Gather(deviceAgent.Accumulator) error {
	panic("implement me")
}

func (f *Fake) SetPointMap(map[string]deviceAgent.PointDefine) {
	panic("implement me")
}

func (f *Fake) FlushPointMap(deviceAgent.Accumulator) error {
	panic("implement me")
}

func (f *Fake) Name() string {
	panic("implement me")
}

func (f *Fake) Set(cmdId string, kv map[string]interface{}) error {
	panic("implement me")
}

func (f *Fake) Get(cmdId string, key []string) interface{} {
	panic("implement me")
}

func (f *Fake) UpdatePointMap(cmdId string, kv map[string]interface{}) error {
	panic("implement me")
}

func (f *Fake) RetrievePointMap(cmdId string, key []string) map[string]deviceAgent.PointDefine {
	panic("implement me")
}

func init() {
	inputs.Add("fake", func() deviceAgent.Input {
		return &Fake{}
	})
}
