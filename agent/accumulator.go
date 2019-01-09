package agent

import (
	"device_adaptor"
	"device_adaptor/selfstat"
	"github.com/rs/zerolog/log"
	"time"
)

var (
	NErrors           = selfstat.Register("agent", "error_count", map[string]string{})
	MetricFieldsCount = selfstat.Register("agent", "field_count", nil)
)

type MetricMaker interface {
	Name() string
	MakeMetric(
		measurement string,
		fields map[string]interface{},
		tags map[string]string,
		quality device_agent.Quality,
		mType device_agent.MetricType,
		t time.Time,
	) device_agent.Metric
}

func NewAccumulator(maker MetricMaker, metrics chan device_agent.Metric) device_agent.Accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	return &acc
}

type accumulator struct {
	metrics   chan device_agent.Metric
	maker     MetricMaker
	precision time.Duration
}

func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	NErrors.Incr(1)
	log.Error().Err(err).Str("plugin", ac.maker.Name()).Msg("ACC ERROR")
}

func (ac *accumulator) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, quality device_agent.Quality, t ...time.Time) {
	if m := ac.maker.MakeMetric(measurement, fields, tags, quality, device_agent.Untyped, ac.getTime(t)); m != nil {
		ac.metrics <- m
	}
	MetricFieldsCount.Incr(int64(len(fields)))
}

func (ac *accumulator) SetPrecision(precision, interval time.Duration) {
	if precision > 0 {
		ac.precision = precision
		return
	}
	switch {
	case interval >= time.Second:
		ac.precision = time.Second
	case interval >= time.Millisecond:
		ac.precision = time.Millisecond
	case interval >= time.Microsecond:
		ac.precision = time.Microsecond
	default:
		ac.precision = time.Nanosecond
	}
}

func (ac accumulator) getTime(t []time.Time) time.Time {
	var timestamp time.Time
	if len(t) > 0 {
		timestamp = t[0]
	} else {
		timestamp = time.Now()
	}
	return timestamp.Round(ac.precision)
}
