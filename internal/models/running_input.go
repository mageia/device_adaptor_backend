package models

import (
	"deviceAdaptor"
	"deviceAdaptor/metric"
	"deviceAdaptor/selfstat"
	"log"
	"time"
)

var GlobalMetricsGathered = selfstat.Register("agent", "metrics_gathered", map[string]string{})

type InputConfig struct {
	Name            string        
	PointMapPath    string        
	PointMapContent string        
	Interval        time.Duration 
}

type RunningInput struct {
	Config          *InputConfig
	Input           deviceAgent.Input
	PointMap        map[string]deviceAgent.PointDefine
	MetricsGathered selfstat.Stat
}

func NewRunningInput(input deviceAgent.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input:  input,
		Config: config,
		MetricsGathered: selfstat.Register(
			input.Name(),
			"metric_count",
			map[string]string{"input": config.Name},
		),
	}
}

func (r *RunningInput) Name() string {
	return "inputs." + r.Config.Name
}

func (r *RunningInput) MakeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	quality deviceAgent.Quality,
	mType deviceAgent.MetricType,
	t time.Time,
) deviceAgent.Metric {
	m, err := metric.New(measurement, tags, fields, quality, t, mType)
	if err != nil {
		log.Printf("Error adding point [%s]: %s\n", measurement, err.Error())
		return nil
	}

	r.MetricsGathered.Incr(1)
	GlobalMetricsGathered.Incr(1)
	return m
}
