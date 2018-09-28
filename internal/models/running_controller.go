package models

import "deviceAdaptor"

type ControllerConfig struct {
	Name string
}
type RunningController struct {
	Controller deviceAgent.Controller
	Config     *ControllerConfig
}

func NewRunningController(
	controller deviceAgent.Controller,
	config *ControllerConfig,
) *RunningController {
	return &RunningController{
		Config:     config,
		Controller: controller,
	}
}

func (r *RunningController) Name() string {
	return "controllers." + r.Config.Name
}