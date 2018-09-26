package outputs

import "deviceAdaptor"

type Creator func() deviceAgent.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
