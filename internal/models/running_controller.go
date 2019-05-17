package models

import "device_adaptor"

type RunningController struct {
	Name       string
	Controller device_adaptor.Controller
}

func NewRunningController(name string, controller device_adaptor.Controller) *RunningController {
	return &RunningController{
		Name:       name,
		Controller: controller,
	}
}
