package vibration

import (
	"device_adaptor"
	"device_adaptor/plugins/processors"
)

type Vibration struct {
}

func (v *Vibration) Apply(in ...deviceAgent.Metric) []deviceAgent.Metric {
	return nil
}

func init() {
	processors.Add("vibration", func() deviceAgent.Processor {
		return &Vibration{}
	})
}
