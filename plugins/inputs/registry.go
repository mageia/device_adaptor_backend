package inputs

import "deviceAgent.General/interfaces"

type Creator func() interfaces.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator)  {
	Inputs[name] = creator
}

