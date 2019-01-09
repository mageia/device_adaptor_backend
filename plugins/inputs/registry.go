package inputs

import "device_adaptor"

type Creator func() device_agent.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
