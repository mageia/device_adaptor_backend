package redis

import (
	"context"
	"device_adaptor"
)

type Redis struct {
}

func (r *Redis) Name() string {
	panic("implement me")
}

func (r *Redis) Start(context.Context) error {
	panic("implement me")
}

func (r *Redis) Stop(context.Context) error {
	panic("implement me")
}

func (r *Redis) RegisterInput(string, device_adaptor.ControllerInput) {
	panic("implement me")
}
