package ws

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/parsers"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type WebSocket struct {
	Address      string `json:"address"`
	client       *websocket.Conn
	connected    bool
	quality      device_agent.Quality
	parser       map[string]parsers.Parser
	msgChan      chan []byte
	errChan      chan error
	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (w *WebSocket) Name() string {
	if w.NameOverride != "" {
		return w.NameOverride
	}
	return w.originName
}

func (w *WebSocket) CheckGather(device_agent.Accumulator) error {
	if !w.connected {
		if e := w.Connect(); e != nil {
			w.errChan <- e
			return e
		}
	}
	return nil
}

func (w *WebSocket) SelfCheck() device_agent.Quality {
	return w.quality
}

func (w *WebSocket) SetPointMap(map[string]points.PointDefine) {}

func (w *WebSocket) Connect() error {
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: time.Second,
	}

	c, _, e := dialer.Dial(w.Address, nil)
	if e != nil {
		return e
	}

	w.connected = true
	w.client = c

	go w.receiveMsg()

	return nil
}
func (w *WebSocket) DisConnect() error {
	if w.connected {
		w.client.Close()
	}
	w.connected = false
	w.client = nil
	return nil
}
func (w *WebSocket) SetParser(parser map[string]parsers.Parser) {
	w.parser = parser
}

func (w *WebSocket) receiveMsg() {
	for {
		if _, m, e := w.client.ReadMessage(); e != nil {
			w.errChan <- e
			return
		} else {
			w.msgChan <- m
		}
	}
}

func (w *WebSocket) Listen(ctx context.Context, acc device_agent.Accumulator) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-w.errChan:
			log.Error().Err(e).Msg("receive Error")
			w.DisConnect()
		case m := <-w.msgChan:
			if p, ok := w.parser["kj66"]; ok {
				p.Parse(m)
				//log.Debug().Interface("dV", dV).Msg("CmdId")
			}
		}
	}
}

func init() {
	inputs.Add("ws", func() device_agent.Input {
		return &WebSocket{
			msgChan: make(chan []byte),
			errChan: make(chan error),
		}
	})
}
