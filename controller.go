package deviceAgent

import "context"

type Controller interface {
	Name() string
	Start(context.Context) error
	Stop(context.Context) error
	RegisterInput(string, ControllerInput)
}

type ControllerInput interface {
	Name() string
	Set(cmdId string, key string, value interface{}) error
}
