package controllers

import "deviceAdaptor"

type Creator func() deviceAgent.Controller

var Controllers = map[string]Creator{}

func Add(name string, creator Creator) {
	Controllers[name] = creator
}
