package rabbitmq

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
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
	panic("implement me")
}

func (r *RabbitMQ) Close() error {
	panic("implement me")
}

func (r *RabbitMQ) Write(metrics []deviceAgent.Metric) error {
	panic("implement me")
}
