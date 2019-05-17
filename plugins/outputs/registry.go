package outputs

import "device_adaptor"

type Creator func() device_adaptor.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
