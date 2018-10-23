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
	//Set(cmdId string, kv map[string]interface{}) error
	Get(cmdId string, key []string) interface{}
	UpdatePointMap(cmdId string, kv map[string]interface{}) error
	RetrievePointMap(cmdId string, key []string) map[string]PointDefine
}

type PointMap struct {
	Name      string                 `json:"name"`
	UniqueId  string                 `json:"unique_id"`
	Parameter float64                `json:"parameter"`
	Options   map[string]string      `json:"options"`
	Unit      string                 `json:"unit"`
	Tags      []string               `json:"tags"`
	Extra     map[string]interface{} `json:"extra"`
}
