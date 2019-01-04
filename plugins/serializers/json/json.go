package json

import (
	"device_adaptor"
	"encoding/json"
	"github.com/json-iterator/go"
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
	serialized, err := json.Marshal(s.createObject(metric))
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, '\n')
	return serialized, err
}

func (s *serializer) SerializeBatch(metrics []deviceAgent.Metric) ([]byte, error) {
	objects := make([]map[string]interface{}, 0, len(metrics))
	for _, metric := range metrics {
		objects = append(objects, s.createObject(metric))
	}
	serialized, err := json.Marshal(objects)
	if err != nil {
		return []byte{}, err
	}
	serialized = append(serialized, byte('\n'))
	return serialized, nil
}

func (s *serializer) SerializeMap(metric deviceAgent.Metric) (map[string]interface{}, error) {
	m := s.createObject(metric)
	return m, nil
}

func (s *serializer) SerializePoints(pointMap deviceAgent.PointMap) (map[string]interface{}, error) {
	points := make(map[string]interface{}, len(pointMap.Points))
	for key, point := range pointMap.Points {
		obj := make(map[string]interface{})
		bytes, err := jsoniter.Marshal(point)
		if err != nil {
			return nil, err
		}
		err = jsoniter.Unmarshal(bytes, &obj)
		if err != nil {
			return nil, err
		}
		points[key] = obj
	}

	m := make(map[string]interface{})
	m["points"] = points
	m["timestamp"] = pointMap.Time.UnixNano() / 1e6
	m["name"] = pointMap.InputName
	return m, nil
}

func (s *serializer) createObject(metric deviceAgent.Metric) map[string]interface{} {
	m := make(map[string]interface{}, 5)
	m["fields"] = metric.Fields()
	m["name"] = metric.Name()
	m["timestamp"] = metric.Time().UnixNano() / 1e6
	m["quality"] = metric.Quality()
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
