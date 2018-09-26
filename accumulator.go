package deviceAgent

import "time"

type Accumulator interface {
	AddFields(measurement string, fields map[string]interface{}, tags map[string]string, t ...time.Time)
	SetPrecision(precision, interval time.Duration)
	AddError(err error)
}
