package redis

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/serializers"
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

type Redis struct {
	Address    string
	Password   string
	Timeout    internal.Duration
	Queue      string
	client     *redis.Client
	serializer serializers.Serializer
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

		result := r.client.HMSet(metric.Name(), m)
		if result.Err() != nil {
			log.Printf("failed to write message: %s", result.Err())
			return fmt.Errorf("failed to write message: %s", result.Err())
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
