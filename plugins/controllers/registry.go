package controllers

import "deviceAdaptor"

//type ControllerInput interface {
//	SetController(controller Controller)
//}
//
//type Controller interface {
//	Set(cmdId string, key string, value interface{}) error
//}
//
//type Config struct {
//	Name string
//}
//
//func NewController(c *Config) (Controller, error) {
//	log.Printf("Controller name: %s\n", c.Name)
//	switch c.Name {
//	default:
//		return newBasicController()
//	}
//	return nil, nil
//}
//
//func newBasicController() (Controller, error) {
//	return &http.HTTP{}, nil
//}

type Creator func() deviceAgent.Controller

var Controllers = map[string]Creator{}

func Add(name string, creator Creator) {
	Controllers[name] = creator
}
