package metric

import (
	"device_adaptor"
	"fmt"
	"time"
)

type metric struct {
	name    string
	tags    []*device_adaptor.Tag
	fields  []*device_adaptor.Field
	tm      time.Time
	quality device_adaptor.Quality
	tp      device_adaptor.MetricType
}

func New(name string,
	tags map[string]string,
	fields map[string]interface{},
	quality device_adaptor.Quality,
	tm time.Time,
	tp device_adaptor.MetricType,
) (device_adaptor.Metric, error) {
	m := &metric{
		name:    name,
		tags:    nil,
		fields:  nil,
		quality: quality,
		tm:      tm,
		tp:      tp,
	}
	if len(tags) > 0 {
		m.tags = make([]*device_adaptor.Tag, 0, len(tags))
		for k, v := range tags {
			m.tags = append(m.tags, &device_adaptor.Tag{Key: k, Value: v})
		}
	}
	m.fields = make([]*device_adaptor.Field, 0, len(fields))

	for k, v := range fields {
		m.AddField(k, v)
	}
	return m, nil
}

func (m *metric) String() string {
	return fmt.Sprintf("%s %v %v %d", m.name, m.Tags(), m.Fields(), m.tm.UnixNano())
}
func (m *metric) Name() string {
	return m.name
}
func (m *metric) Tags() map[string]string {
	tags := make(map[string]string, len(m.tags))
	for _, tag := range m.tags {
		tags[tag.Key] = tag.Value
	}
	return tags
}
func (m *metric) TagList() []*device_adaptor.Tag {
	return m.tags
}
func (m *metric) Fields() map[string]interface{} {
	fields := make(map[string]interface{}, len(m.fields))
	for _, field := range m.fields {
		fields[field.Key] = field.Value
	}
	return fields
}
func (m *metric) FieldList() []*device_adaptor.Field {
	return m.fields
}
func (m *metric) Time() time.Time {
	return m.tm
}
func (m *metric) Quality() device_adaptor.Quality {
	return m.quality
}
func (m *metric) SetName(name string) {
	m.name = name
}
func (m *metric) AddPrefix(prefix string) {
	m.name = prefix + m.name
}
func (m *metric) AddSuffix(suffix string) {
	m.name = m.name + suffix
}
func (m *metric) GetTag(key string) (string, bool) {
	for _, tag := range m.tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}
func (m *metric) HasTag(key string) bool {
	for _, tag := range m.tags {
		if tag.Key == key {
			return true
		}
	}
	return false
}
func (m *metric) AddTag(key, value string) {
	for i, tag := range m.tags {
		if key > tag.Key {
			continue
		}
		if key == tag.Key {
			tag.Value = value
			return
		}
		m.tags = append(m.tags, nil)
		copy(m.tags[i+1:], m.tags[i:])
		m.tags[i] = &device_adaptor.Tag{Key: key, Value: value}
		return
	}
}
func (m *metric) RemoveTag(key string) {
	for i, tag := range m.tags {
		if tag.Key == key {
			copy(m.tags[i:], m.tags[i+1:])
			m.tags[len(m.tags)-1] = nil
			m.tags = m.tags[:len(m.tags)-1]
			return
		}
	}
}
func (m *metric) GetField(key string) (interface{}, bool) {
	for _, field := range m.fields {
		if field.Key == key {
			return field.Value, true
		}
	}
	return nil, false
}
func (m *metric) HasField(key string) bool {
	for _, field := range m.fields {
		if field.Key == key {
			return true
		}
	}
	return false
}
func (m *metric) AddField(key string, value interface{}) {
	for i, field := range m.fields {
		if key == field.Key {
			m.fields[i] = &device_adaptor.Field{Key: key, Value: value}
			return
		}
	}
	m.fields = append(m.fields, &device_adaptor.Field{Key: key, Value: value})
}
func (m *metric) RemoveField(key string) {
	for i, field := range m.fields {
		if field.Key == key {
			copy(m.fields[i:], m.fields[i+1:])
			m.fields[len(m.fields)-1] = nil
			m.fields = m.fields[:len(m.fields)-1]
			return
		}
	}
}
func (m *metric) SetTime(t time.Time) {
	m.tm = t
}
func (m *metric) Copy() device_adaptor.Metric {
	m2 := &metric{
		name:    m.name,
		tags:    make([]*device_adaptor.Tag, len(m.tags)),
		fields:  make([]*device_adaptor.Field, len(m.fields)),
		tm:      m.tm,
		quality: m.quality,
	}
	for i, tag := range m.tags {
		m2.tags[i] = tag
	}
	for i, field := range m.fields {
		m2.fields[i] = field
	}
	return m2
}
