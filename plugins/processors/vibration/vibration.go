package vibration

import (
	"device_adaptor"
	"device_adaptor/plugins/processors"
)

type Vibration struct {
}

func (v *Vibration) Apply(in ...device_agent.Metric) []device_agent.Metric {
	return nil
}

func init() {
	processors.Add("vibration", func() device_agent.Processor {
		return &Vibration{}
	})
}
