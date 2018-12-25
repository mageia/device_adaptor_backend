package outputs

import "device_adaptor"

type Creator func() deviceAgent.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
