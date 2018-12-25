package models

import "device_adaptor"

type RunningController struct {
	Name       string
	Controller deviceAgent.Controller
}

func NewRunningController(name string, controller deviceAgent.Controller) *RunningController {
	return &RunningController{
		Name:       name,
		Controller: controller,
	}
}
