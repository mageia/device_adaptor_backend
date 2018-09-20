package models

import "deviceAgent.General/interfaces"

type OutputConfig struct {
	Name string
}
type RunningOutput struct {
	Name   string
	Output interfaces.Output
	Config *OutputConfig
}
