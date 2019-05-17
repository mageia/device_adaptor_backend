package vibration

import (
	"device_adaptor"
	"device_adaptor/plugins/processors"
)

type Vibration struct {
}

func (v *Vibration) Apply(in ...device_adaptor.Metric) []device_adaptor.Metric {
	return nil
}

func init() {
	processors.Add("vibration", func() device_adaptor.Processor {
		return &Vibration{}
	})
}
