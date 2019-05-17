package opc_ws

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type OpcWs struct {
	Address       string `json:"address"`
	OPCServerName string `json:"opc_server_name"`

	client       *websocket.Conn
	connected    bool
	quality      device_adaptor.Quality
	msgChan      chan []byte
	errChan      chan error
	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (o *OpcWs) Name() string {
	if o.NameOverride != "" {
		return o.NameOverride
	}
	return o.originName
}

func (o *OpcWs) CheckGather(device_adaptor.Accumulator) error {
	if !o.connected {
		if e := o.Start(); e != nil {
			o.errChan <- e
			return e
		}
	}

	body := map[string]interface{}{
		"cmd":             "init",
		"opc_server_host": "localhost",
		"opc_server_name": o.OPCServerName,
		"params":          []string{"a", "b"},
	}
	if e := o.client.WriteJSON(body); e != nil {
		log.Error().Err(e).Msg("WriteJSON")
		return e
	}

	return nil
}

func (o *OpcWs) SelfCheck() device_adaptor.Quality {
	return o.quality
}

func (o *OpcWs) SetPointMap(map[string]points.PointDefine) {}

func (o *OpcWs) Start() error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second,
	}

	c, _, e := dialer.Dial(o.Address, nil)
	if e != nil {
		return e
	}

	o.connected = true
	o.client = c

	return nil
}

func (o *OpcWs) Stop() {
	if o.connected {
		o.client.Close()
	}
	o.connected = false
	o.client = nil
}


func init() {
	inputs.Add("opc_ws", func() device_adaptor.Input {
		return &OpcWs{
			msgChan: make(chan []byte),
			errChan: make(chan error),
		}
	})
}
