package serializers

import (
	"device_adaptor"
	"device_adaptor/plugins/serializers/json"
	"fmt"
	"time"
)

type SerializerOutput interface {
	SetSerializer(serializer Serializer)
}

type Serializer interface {
	Serialize(metric device_agent.Metric) ([]byte, error)
	SerializeBatch(metrics []device_agent.Metric) ([]byte, error)
	SerializeMap(metric device_agent.Metric) (map[string]interface{}, error)
	SerializePoints(pointMap device_agent.PointMap) (map[string]interface{}, error)
}

type Config struct {
	DataFormat     string
	TimestampUnits time.Duration
	//Prefix         string
}

func NewSerializer(config *Config) (Serializer, error) {
	var err error
	var serializer Serializer
	switch config.DataFormat {
	case "json":
		serializer, err = NewJsonSerializer(config.TimestampUnits)
	default:
		err = fmt.Errorf("invalid data format: %s", config.DataFormat)
	}
	return serializer, err
}

func NewJsonSerializer(timestampUnits time.Duration) (Serializer, error) {
	return json.NewSerializer(timestampUnits)
}
