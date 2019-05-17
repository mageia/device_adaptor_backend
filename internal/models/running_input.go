package models

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/metric"
	"device_adaptor/selfstat"
	"log"
	"time"
)

var GlobalMetricsCheckGathered = selfstat.Register("agent", "metrics_CheckGathered", map[string]string{})

type InputConfig struct {
	Name            string
	//PointMapPath    string
	//PointMapContent string
	Interval        time.Duration
}

type RunningInput struct {
	Config               *InputConfig
	Input                device_adaptor.Input
	PointMap             map[string]points.PointDefine
	MetricsCheckGathered selfstat.Stat
}

func NewRunningInput(input device_adaptor.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input:  input,
		Config: config,
		MetricsCheckGathered: selfstat.Register(
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
	quality device_adaptor.Quality,
	mType device_adaptor.MetricType,
	t time.Time,
) device_adaptor.Metric {
	m, err := metric.New(measurement, tags, fields, quality, t, mType)
	if err != nil {
		log.Printf("Error adding point [%s]: %s", measurement, err.Error())
		return nil
	}

	r.MetricsCheckGathered.Incr(1)
	GlobalMetricsCheckGathered.Incr(1)
	return m
}
