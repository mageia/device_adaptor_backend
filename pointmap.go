package device_adaptor

import (
	"device_adaptor/internal/points"
	"time"
)

type PointMap struct {
	Time time.Time
	InputName string
	Points map[string]points.PointDefine
}
