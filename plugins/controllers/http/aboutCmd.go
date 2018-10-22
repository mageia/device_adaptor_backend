package http

import (
	"deviceAdaptor"
	"errors"
	"strings"
)

type command struct {
	input   deviceAgent.ControllerInput
	cmdType string
	cmdId   string
	subCmd  string
	keys    interface{}
}

func (c command) execute() (interface{}, error) {
	switch strings.ToUpper(c.cmdType) {
	case "GET":
		switch c.subCmd {
		case "point_meta":
			return c.input.RetrievePointMap(c.cmdId, c.keys.([]string)), nil

		case "point_value":
			return c.input.Get(c.cmdId, c.keys.([]string)), nil
		default:
			return nil, errors.New("unknown sub command: " + c.subCmd)
		}
	default:
		return nil, errors.New("unsupported command")
	}

	return nil, errors.New("unsupported command")
}
