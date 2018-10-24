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
	OriginName() string
	SetValue(map[string]interface{}) error
	UpdatePointMap(map[string]interface{}) error
	RetrievePointMap([]string) map[string]PointDefine
}
