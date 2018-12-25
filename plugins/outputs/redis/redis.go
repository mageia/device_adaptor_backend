package redis

import (
	"device_adaptor"
	"device_adaptor/alarm"
	"device_adaptor/internal"
	"device_adaptor/plugins/outputs"
	"device_adaptor/plugins/serializers"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/json-iterator/go"
	"log"
	"time"
)

type Redis struct {
	UrlAddress string            `json:"url_address"`
	Timeout    internal.Duration `json:"timeout"`
	client     *redis.Client
	serializer serializers.Serializer
}

func (r *Redis) Write(metrics []deviceAgent.Metric) error {
	if r.client == nil {
		return errors.New("disconnected")
	}

	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		m, err := r.serializer.SerializeMap(metric)
		if err != nil {
			return fmt.Errorf("failed to serialize message: %s", err)
		}

		sV, err := jsoniter.MarshalToString(m)
		if err != nil {
			return err
		}

		if err := r.client.Publish(metric.Name(), sV).Err(); err != nil {
			log.Println(err)
			return err
		}
		alarm.ChanRealTime <- alarm.RealTime{PluginName: metric.Name(), Metric: m}
	}

	return nil
}

func (r *Redis) Connect() error {
	if r.UrlAddress == "" {
		r.UrlAddress = "redis://localhost:6379/0"
	}
	rO, _ := redis.ParseURL(r.UrlAddress)
	if r.Timeout.Duration == 0 {
		r.Timeout.Duration = time.Second * 10
	}
	rO.ReadTimeout = r.Timeout.Duration
	rO.WriteTimeout = r.Timeout.Duration
	rO.IdleCheckFrequency = time.Second * 3

	c := redis.NewClient(rO)
	if _, err := c.Ping().Result(); err != nil {
		return fmt.Errorf("failed to connect redis: %s", err)
	}
	r.client = c
	return nil
}

func (r *Redis) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *Redis) SetSerializer(s serializers.Serializer) {
	r.serializer = s
}

func init() {
	outputs.Add("redis", func() deviceAgent.Output {
		return &Redis{}
	})
}
