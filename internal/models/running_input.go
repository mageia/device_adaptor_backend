package models

import (
	"deviceAdaptor"
	"deviceAdaptor/metric"
	"deviceAdaptor/selfstat"
	"log"
	"time"
)

type InputConfig struct {
	Name              string
	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	PointMapPath      string
	Tags              map[string]string
	Interval          time.Duration
}

type RunningInput struct {
	Input           deviceAgent.Input
	Config          *InputConfig
	trace           bool
	defaultTags     map[string]string
	MetricsGathered selfstat.Stat
	PointMap        map[string]deviceAgent.PointDefine
}

func NewRunningInput(input deviceAgent.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input:       input,
		Config:      config,
		defaultTags: make(map[string]string),
	}
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

func (r *RunningInput) Trace() bool {
	return r.trace
}
func (r *RunningInput) SetTrace(trace bool) {
	r.trace = trace
}

func (r *RunningInput) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
