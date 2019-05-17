package device_adaptor

import (
	"time"
)

type MetricType int8

const (
	_ MetricType = iota
	Untyped
	Counter
	Gauge
	Summary
	Histogram
)

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Field struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type Metric interface {
	Name() string
	Tags() map[string]string
	TagList() []*Tag
	Fields() map[string]interface{}
	FieldList() []*Field
	Time() time.Time
	Quality() Quality

	SetName(name string)
	AddPrefix(prefix string)
	AddSuffix(suffix string)

	GetTag(key string) (string, bool)
	HasTag(key string) bool
	AddTag(key, value string)
	RemoveTag(key string)

	GetField(key string) (interface{}, bool)
	HasField(key string) bool
	AddField(key string, value interface{})
	RemoveField(key string)

	SetTime(t time.Time)
	Copy() Metric
}
