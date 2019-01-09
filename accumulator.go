package device_agent

import "time"

type Accumulator interface {
	AddFields(measurement string, fields map[string]interface{}, tags map[string]string, quality Quality, t ...time.Time)
	SetPrecision(precision, interval time.Duration)
	AddError(err error)
}
