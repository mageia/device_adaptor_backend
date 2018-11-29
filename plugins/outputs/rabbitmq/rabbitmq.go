package rabbitmq

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/serializers"
)

type RabbitMQ struct {
	URLAddress string            `json:"url_address"`
	Timeout    internal.Duration `json:"timeout"`
	serializer serializers.Serializer
}

func (r *RabbitMQ) SetSerializer(serializer serializers.Serializer) {
	r.serializer = serializer
}

func (r *RabbitMQ) Connect() error {
	return nil
}

func (r *RabbitMQ) Close() error {
	return nil
}

func (r *RabbitMQ) Write(metrics []deviceAgent.Metric) error {
	return nil
}

func init() {
	outputs.Add("rabbitmq", func() deviceAgent.Output {
		return &RabbitMQ{}
	})
}
