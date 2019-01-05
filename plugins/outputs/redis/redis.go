/*
	redis 输出（支持输出点表）

	实时数据
    以 Redis PUB/SUB 形式发布，Channel 名字就是 Input 名字。

    点表信息
    1. 会将每个 Input 的点表 HSET 到 Redis 中，键通过 PointsKey 配置，Field 就是 Input 名字；
    2. 会将每个 Input 的点表版本（时间戳）HSET 到 Redis 中，键通过 PointsVersionKey 配置，Field 就是 Input 名字；
    3. 点表变更时，会将对应的 Input 名字通过 PUB/SUB 形式发布出来，Channel 通过 PointsKey 配置。
 */
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
	PointsKey  string            `json:"points_key"`
	PointsVersionKey string      `json:"points_version_key"`
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

func (r *Redis) WritePointMap(pointMap deviceAgent.PointMap) error {
	if r.client == nil {
		return errors.New("disconnected")
	}

	obj, err := r.serializer.SerializePoints(pointMap)
	if err != nil {
		return err
	}

	str, err := jsoniter.MarshalToString(obj)
	if err != nil {
		return err
	}

	// 将点表内容覆盖到指定键（用于同步查询）
	if err := r.client.HSet(r.PointsKey, pointMap.InputName, str).Err(); err != nil {
		log.Println(err)
		return err
	}

	// 将点表的版本（时间戳）覆盖到指定键
	if err := r.client.HSet(r.PointsVersionKey, pointMap.InputName, pointMap.Time.UnixNano() / 1e6).Err(); err != nil {
		log.Println(err)
		return err
	}

	// 将发送变更的 Input 名称发布到 redis 通道（用于异步通知）
	if err := r.client.Publish(r.PointsKey, pointMap.InputName).Err(); err != nil {
		log.Println(err)
		return err
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

	r.client.Del(r.PointsKey)
	r.client.Del(r.PointsVersionKey)

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
