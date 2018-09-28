package models

import (
	"deviceAdaptor"
	"deviceAdaptor/metric"
	"deviceAdaptor/selfstat"
	"log"
	"time"
)

type InputConfig struct {
	Name         string
	PointMapPath string
	Interval     time.Duration
}

type RunningInput struct {
	Config          *InputConfig
	Input           deviceAgent.Input
	PointMap        map[string]deviceAgent.PointDefine
	MetricsGathered selfstat.Stat
}

func NewRunningInput(input deviceAgent.Input, config *InputConfig) *RunningInput {
	return &RunningInput{Input: input, Config: config}
}

func (r *RunningInput) Name() string {
	return "inputs." + r.Config.Name
}

func (r *RunningInput) MakeMetric(measurement string, fields map[string]interface{}, tags map[string]string, t time.Time) deviceAgent.Metric {
	m, err := metric.New(measurement, tags, fields, t)
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return nil
	}
	return m
}

//func (r *RunningInput) SetDefaultTags(tags map[string]string) {
//	r.defaultTags = tags
//}
