package amqp

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/plugins/outputs"
	"device_adaptor/plugins/serializers"
	"device_adaptor/utils"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	URLAddress   string `json:"url_address"`
	ExchangeName string `json:"exchange_name"`

	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto_delete"`
	Exclusive  bool   `json:"exclusive"`
	NoWait     bool   `json:"no_wait"`
	Mandatory  bool   `json:"mandatory"`
	Immediate  bool   `json:"immediate"`

	Timeout internal.Duration `json:"timeout"`

	serializer serializers.Serializer
	connected  bool
	client     *amqp.Connection
	channel    *amqp.Channel
}

func (r *RabbitMQ) SetSerializer(serializer serializers.Serializer) {
	r.serializer = serializer
}

func (r *RabbitMQ) Connect() error {
	if r.connected {
		return nil
	}
	conn, err := amqp.Dial(r.URLAddress)
	if err != nil {
		return fmt.Errorf("[%s]: %s", utils.GetLineNo(), err.Error())
	}
	r.client = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("RabbitMQ Channel")
		conn.Close()
		return fmt.Errorf("[%s]: %s", utils.GetLineNo(), err.Error())
	}
	if e := ch.ExchangeDeclare(r.ExchangeName, "topic", r.Durable, r.AutoDelete, false, r.NoWait, nil); e != nil {
		log.Error().Err(e).Msg("ExchangeDeclarePassive")
	}

	r.channel = ch
	r.connected = true

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.connected {
		r.channel.Close()
		r.client.Close()
		r.connected = false
	}
	return nil
}

func (r *RabbitMQ) Write(metrics []device_adaptor.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	if !r.connected {
		return r.Connect()
	}

	for _, metric := range metrics {
		m, err := r.serializer.SerializeMap(metric)
		if err != nil {
			return fmt.Errorf("[%s]: %s", utils.GetLineNo(), err.Error())
		}

		pV, err := jsoniter.Marshal(m)
		if err != nil {
			return fmt.Errorf("[%s]: %s", utils.GetLineNo(), err.Error())
		}
		err = r.channel.Publish(r.ExchangeName, metric.Name(), r.Mandatory, r.Immediate, amqp.Publishing{
			ContentType: "application/json",
			Body:        pV,
		})
		if err != nil {
			r.Close()
			r.Connect()
			return fmt.Errorf("[%s]: %s", utils.GetLineNo(), err.Error())
		}
	}
	return nil
}

func init() {
	outputs.Add("amqp", func() device_adaptor.Output {
		return &RabbitMQ{
			AutoDelete: true,
		}
	})
}
