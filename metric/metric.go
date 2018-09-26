package metric

import (
	"deviceAdaptor"
	"fmt"
	"time"
)

type metric struct {
	name   string
	tags   []*deviceAgent.Tag
	fields []*deviceAgent.Field
	tm     time.Time
}

func New(name string, tags map[string]string, fields map[string]interface{}, tm time.Time) (deviceAgent.Metric, error) {
	m := &metric{
		name:   name,
		tags:   nil,
		fields: nil,
		tm:     tm,
	}
	if len(tags) > 0 {
		m.tags = make([]*deviceAgent.Tag, 0, len(tags))
		for k, v := range tags {
			m.tags = append(m.tags, &deviceAgent.Tag{Key: k, Value: v})
		}
	}
	m.fields = make([]*deviceAgent.Field, 0, len(fields))
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
func (m *metric) TagList() []*deviceAgent.Tag {
	return m.tags
}
func (m *metric) Fields() map[string]interface{} {
	fields := make(map[string]interface{}, len(m.fields))
	for _, field := range m.fields {
		fields[field.Key] = field.Value
	}
	return fields
}
func (m *metric) FieldList() []*deviceAgent.Field {
	return m.fields
}
func (m *metric) Time() time.Time {
	return m.tm
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
		m.tags[i] = &deviceAgent.Tag{Key: key, Value: value}
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
			m.fields[i] = &deviceAgent.Field{Key: key, Value: convertField(value)}
			return
		}
	}
	m.fields = append(m.fields, &deviceAgent.Field{Key: key, Value: convertField(value)})
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
func (m *metric) Copy() deviceAgent.Metric {
	m2 := &metric{
		name:   m.name,
		tags:   make([]*deviceAgent.Tag, len(m.tags)),
		fields: make([]*deviceAgent.Field, len(m.fields)),
		tm:     m.tm,
	}
	for i, tag := range m.tags {
		m2.tags[i] = tag
	}
	for i, field := range m.fields {
		m2.fields[i] = field
	}
	return m2
}

func convertField(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return uint64(v)
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	default:
		return nil
	}
}
