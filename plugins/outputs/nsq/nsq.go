package nsq

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/plugins/outputs"
	"github.com/json-iterator/go"
	"github.com/nsqio/go-nsq"
	"github.com/rs/zerolog/log"
	"time"
)

type NSQ struct {
	UrlAddress string            `json:"url_address"`
	Timeout    internal.Duration `json:"timeout"`
	Topic      string
	client     *nsq.Producer
	connected  bool
}

func (n *NSQ) Connect() error {
	c := nsq.NewConfig()
	c.DialTimeout = n.Timeout.Duration
	p, err := nsq.NewProducer(n.UrlAddress, c)
	if err != nil {
		return err
	}
	p.SetLogger(nil, 0)
	n.client = p
	n.connected = true
	return nil
}

func (n *NSQ) Close() error {
	if n.connected {
		n.client.Stop()
		n.connected = false
	}
	return nil
}

func (n *NSQ) Write(metrics []device_adaptor.Metric) error {
	for _, metric := range metrics {
		body := make(map[string]interface{})
		for k, v := range metric.Fields() {
			body[k] = map[string]interface{}{"Desc": k, "Value": v}
		}
		sV, err := jsoniter.Marshal(body)
		if err != nil {
			return err
		}
		if err := n.client.Publish(n.Topic, sV); err != nil {
			log.Error().Err(err).Msg("nsq.Publish")
			n.connected = false
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("nsq", func() device_adaptor.Output {
		return &NSQ{Timeout: internal.Duration{Duration: time.Second * 5}}
	})
}
