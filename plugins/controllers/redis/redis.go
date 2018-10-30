package redis

import (
	"context"
	"deviceAdaptor"
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

func (r *Redis) RegisterInput(string, deviceAgent.ControllerInput) {
	panic("implement me")
}
