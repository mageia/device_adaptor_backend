package models

import "device_adaptor"

type RunningController struct {
	Name       string
	Controller device_agent.Controller
}

func NewRunningController(name string, controller device_agent.Controller) *RunningController {
	return &RunningController{
		Name:       name,
		Controller: controller,
	}
}
