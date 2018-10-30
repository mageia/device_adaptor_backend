package mqtt

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/serializers"
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"log"
	"net/url"
	"strings"
)

type Mqtt struct {
	UrlAddress    string
	OutputChannel string
	client        mqtt.Client
	serializer    serializers.Serializer
}

func (mt *Mqtt) Connect() error {
	opt, err := url.Parse(mt.UrlAddress)
	if err != nil {
		return fmt.Errorf("failed to parse mqtt url: %s", err)
	}
	if strings.ToLower(opt.Scheme) != "mqtt" {
		return fmt.Errorf("invalid mqtt scheme: %s", opt.Scheme)
	}
	passwd, _ := opt.User.Password()
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s", opt.Host))
	opts.SetUsername(opt.User.Username())
	opts.SetPassword(passwd)
	// TODO: check this
	opts.SetClientID(uuid.New().String())
	mt.client = mqtt.NewClient(opts)
	return nil
}

func (mt *Mqtt) Close() error {
	mt.client.Disconnect(3000)
	return nil
}

func (mt *Mqtt) Write(metrics []deviceAgent.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	for _, metric := range metrics {
		m, err := mt.serializer.SerializeMap(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}

		sV, err := jsoniter.MarshalToString(m)
		if err != nil {
			return err
		}

		if mt.OutputChannel == "" {
			mt.OutputChannel = "output_test"
		}
		token := mt.client.Publish(mt.OutputChannel, 2, true, sV)
		if token.Error() != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func init() {
	outputs.Add("mqtt", func() deviceAgent.Output {
		return &Mqtt{}
	})
}
