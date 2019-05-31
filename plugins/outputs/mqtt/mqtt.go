package mqtt

import (
	"device_adaptor"
	"device_adaptor/plugins/outputs"
	"device_adaptor/plugins/serializers"
	"errors"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"net/url"
	"strings"
)

type Mqtt struct {
	UrlAddress string `json:"url_address"`
	client     mqtt.Client
	connected  bool
	serializer serializers.Serializer
}

func (mt *Mqtt) Connect() error {
	if mt.connected {
		return nil
	}

	opt, err := url.Parse(mt.UrlAddress)
	if err != nil {
		return fmt.Errorf("failed to parse mqtt url: %s", err)
	}
	if strings.ToLower(opt.Scheme) != "mqtt" {
		return fmt.Errorf("invalid mqtt scheme: %s", opt.Scheme)
	}
	p, _ := opt.User.Password()
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", opt.Host))
	opts.SetUsername(opt.User.Username())
	opts.SetPassword(p)
	opts.SetClientID(uuid.New().String())
	_client := mqtt.NewClient(opts)

	if token := _client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	if _client.IsConnected() {
		mt.client = _client
		mt.connected = true
	}
	return nil
}

func (mt *Mqtt) Close() error {
	if mt.connected {
		mt.client.Disconnect(0)
		mt.connected = false
	}
	return nil
}

func (mt *Mqtt) Write(metrics []device_adaptor.Metric) error {
	if !mt.connected {
		return errors.New("disconnected")
	}
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		m, err := mt.serializer.SerializeMap(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}

		sV, err := jsoniter.Marshal(m)
		if err != nil {
			return err
		}

		token := mt.client.Publish(metric.Name(), 0, true, sV)
		if token.Error() != nil {
			log.Error().Err(token.Error()).Msg("mqtt.Publish")
			return err
		}
	}

	return nil
}

func (mt *Mqtt) SetSerializer(s serializers.Serializer) {
	mt.serializer = s
}
func (mt *Mqtt) WritePointMap(pointMap device_adaptor.PointMap) error {
	return nil
}

func init() {
	outputs.Add("mqtt", func() device_adaptor.Output {
		return &Mqtt{}
	})
}
