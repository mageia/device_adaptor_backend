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
	//Set(cmdId string, kv map[string]interface{}) error
	//Get(cmdId string, key []string) interface{}
	UpdatePointMap(map[string]interface{}) error
	RetrievePointMap([]string) map[string]PointDefine
}
