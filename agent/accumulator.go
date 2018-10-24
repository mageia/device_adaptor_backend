package agent

import (
	"deviceAdaptor"
	"deviceAdaptor/selfstat"
	"log"
	"runtime"
	"strings"
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
		quality deviceAgent.Quality,
		mType deviceAgent.MetricType,
		t time.Time,
	) deviceAgent.Metric
}

func NewAccumulator(maker MetricMaker, metrics chan deviceAgent.Metric) deviceAgent.Accumulator {
	acc := accumulator{
		maker:     maker,
		metrics:   metrics,
		precision: time.Nanosecond,
	}
	return &acc
}

type accumulator struct {
	metrics   chan deviceAgent.Metric
	maker     MetricMaker
	precision time.Duration
}

func (ac *accumulator) AddError(err error) {
	if err == nil {
		return
	}
	NErrors.Incr(1)
	_, f, l, ok := runtime.Caller(1)
	if ok {
		fL := strings.Split(f, "/")
		f = fL[len(fL)-1]
		log.Printf("E! Error in plugin [%s][%s:%d]: %v", ac.maker.Name(), f, l, err)
	} else {
		log.Printf("E! Error in plugin [%s]: %v", ac.maker.Name(), err)
	}
}

func (ac *accumulator) AddFields(measurement string, fields map[string]interface{}, tags map[string]string, quality deviceAgent.Quality, t ...time.Time) {
	if m := ac.maker.MakeMetric(measurement, fields, tags, quality, deviceAgent.Untyped, ac.getTime(t)); m != nil {
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
