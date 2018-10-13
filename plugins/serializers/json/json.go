package json

import (
	"deviceAdaptor"
	"encoding/json"
	"time"
)

type serializer struct {
	TimestampUnits time.Duration
}

func NewSerializer(timestampUnits time.Duration) (*serializer, error) {
	return &serializer{
		TimestampUnits: truncateDuration(timestampUnits),
	}, nil
}

func (s *serializer) Serialize(metric deviceAgent.Metric) ([]byte, error) {
	m := map[string]map[string]interface{}{
		metric.Name(): metric.Fields(),
	}
	serialized, err := json.Marshal(m)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')
	return serialized, err
}

func (s *serializer) SerializeBatch(metrics []deviceAgent.Metric) ([]byte, error) {
	objects := make([]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		objects = append(objects, s.createObject(metric))
	}
	obj := map[string]interface{}{
		"metrics": objects,
	}
	serialized, err := json.Marshal(obj)
	if err != nil {
		return []byte{}, err
	}
	return serialized, nil
}

func (s *serializer) SerializeMap(metric deviceAgent.Metric) (map[string]interface{}, error) {
	r := make(map[string]interface{})

	for k, v := range metric.Fields() {
		serialized, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		r[k] = serialized
	}
	return r, nil
}

func (s *serializer) createObject(metric deviceAgent.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 4)
	m["tags"] = metric.Tags()
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.Time().Unix()
	return m
}

func truncateDuration(units time.Duration) time.Duration {
	if units <= 0 {
		return time.Second
	}
	d := time.Nanosecond
	for {
		if d*10 > units {
			return d
		}
		d = d * 10
	}
}
