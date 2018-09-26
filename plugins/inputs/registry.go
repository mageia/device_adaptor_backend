package inputs

import "deviceAdaptor"

type Creator func() deviceAgent.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator)  {
	Inputs[name] = creator
}

