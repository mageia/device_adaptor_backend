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
	Address     string
	Password    string
	Timeout     internal.Duration
	ChannelName string
	client      *redis.Client
	serializer  serializers.Serializer
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

		if err := r.client.Set(metric.Name(), sV, 0).Err(); err != nil {
			log.Println(err)
			return err
		}

		if r.ChannelName != "" {
			if err := r.client.Publish(r.ChannelName, sV).Err(); err != nil {
				log.Println(err)
				return err
			}
		}
	}

	return nil
}

func (r *Redis) Connect() error {
	if r.Address == "" {
		r.Address = "localhost:6379"
	}
	c := redis.NewClient(&redis.Options{
		Addr:         r.Address,
		ReadTimeout:  r.Timeout.Duration,
		WriteTimeout: r.Timeout.Duration,
	})
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
