package deviceAgent

import "context"

type ControllerConfig struct {
}

type Controller interface {
	Start(context.Context) error
}

type ControllerInput interface {
	Set(cmdId, key string, value interface{}) error
}

