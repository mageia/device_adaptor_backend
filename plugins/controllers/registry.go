package controllers

import "device_adaptor"

type Creator func() deviceAgent.Controller

var Controllers = map[string]Creator{}

func Add(name string, creator Creator) {
	Controllers[name] = creator
}
