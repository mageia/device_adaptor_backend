package deviceAgent

import "time"

type Tag struct {
	Key   string
	Value string
}

type Field struct {
	Key   string
	Value interface{}
}

type Metric interface {
	Name() string
	Tags() map[string]string
	TagList() []*Tag
	Fields() map[string]interface{}
	FieldList() []*Field
	Time() time.Time

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
