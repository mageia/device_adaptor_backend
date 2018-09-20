package models

import (
	"deviceAgent.General/interfaces"
	"time"
)

type InputConfig struct {
	Name              string
	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Interval          time.Duration
}

type RunningInput struct {
	Input  interfaces.Input
	Config *InputConfig
}

func (r *RunningInput) Name() string {
	return "inputs." + r.Config.Name
}

func NewRunningInput(input interfaces.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input: input,
		Config: config,
	}
}