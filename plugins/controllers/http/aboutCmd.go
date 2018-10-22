package http

import (
	"deviceAdaptor"
	"errors"
	"fmt"
	"strings"
)

type command struct {
	input       deviceAgent.ControllerInput
	cmdType     string
	cmdId       string
	subCmd      string
	value       interface{}
	callbackUrl string
}

type result struct {
	CmdId       string      `json:"cmd_id"`
	Success     bool        `json:"success"`
	CallbackUrl string      `json:"callback_url"`
	Msg         interface{} `json:"msg"`
}

func (c *command) success(msg interface{}) result {
	return result{
		CmdId:       c.cmdId,
		Success:     true,
		CallbackUrl: c.callbackUrl,
		Msg:         msg,
	}
}
func (c *command) failed(err error) result {
	return result{
		CmdId:       c.cmdId,
		Success:     false,
		CallbackUrl: c.callbackUrl,
		Msg:         err.Error(),
	}
}

func (c command) execute() result {
	switch strings.ToUpper(c.cmdType) {
	case "GET":
		switch c.subCmd {
		case "point_meta":
			p := c.input.RetrievePointMap(c.cmdId, c.value.([]string))
			if len(p) == 0 {
				return c.failed(fmt.Errorf("no such point: %v", c.value))
			}
			//g, err := jsoniter.MarshalToString(p)
			//if err != nil {
			//	return c.failed(err)
			//}
			return c.success(p)

		case "point_value":
			//g, err := jsoniter.MarshalToString(c.input.Get(c.cmdId, c.value.([]string)))
			//if err != nil {
			//	return c.failed(err)
			//}
			return c.success(c.input.Get(c.cmdId, c.value.([]string)))
		default:
			return c.failed(errors.New("unknown sub command: " + c.subCmd))
		}
	case "SET":
		switch c.subCmd {
		case "point_meta":
			if err := c.input.UpdatePointMap(c.cmdId, c.value.(map[string]interface{})); err != nil {
				return c.failed(err)
			}
			return c.success("")
		case "point_value":
			if err := c.input.Set(c.cmdId, c.value.(map[string]interface{})); err != nil {
				return c.failed(err)
			}
			return c.success("")
		default:
			return c.failed(errors.New("unknown sub command: " + c.subCmd))
		}
	default:
		return c.failed(errors.New("unsupported command"))
	}
	return c.failed(errors.New("unsupported command"))
}
