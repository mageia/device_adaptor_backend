package redis

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/serializers"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/json-iterator/go"
	"log"
)

type Redis struct {
	UrlAddress    string
	Timeout       internal.Duration
	OutputChannel string
	client        *redis.Client
	serializer    serializers.Serializer
}

func (r *Redis) Write(metrics []deviceAgent.Metric) error {
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

		if r.OutputChannel == "" {
			r.OutputChannel = "output_test"
		}

		if err := r.client.Publish(r.OutputChannel, sV).Err(); err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (r *Redis) Connect() error {
	if r.UrlAddress == "" {
		r.UrlAddress = "redis://localhost:6379/0"
	}
	rO, _ := redis.ParseURL(r.UrlAddress)
	rO.ReadTimeout = r.Timeout.Duration
	rO.WriteTimeout = r.Timeout.Duration

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
