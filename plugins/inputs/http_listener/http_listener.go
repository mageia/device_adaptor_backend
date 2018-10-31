package http_listener

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/parsers"
	"log"
)

type HTTPListener struct {
	Address      string
	Parsers      map[string]parsers.Parser
	NameOverride string

	parsers map[string]parsers.Parser
}

func (h *HTTPListener) SetParser(parsers map[string]parsers.Parser) {
	h.parsers = parsers
}

func (h *HTTPListener) SelfCheck() deviceAgent.Quality {
	return deviceAgent.QualityGood
}

func (h *HTTPListener) FlushPointMap(deviceAgent.Accumulator) error {
	return nil
}

func (h *HTTPListener) Name() string {
	return "http_listener"
}

func (h *HTTPListener) Gather(deviceAgent.Accumulator) error {
	for k, v := range h.parsers {
		log.Println(k, v)
	}
	return nil
}

func (h *HTTPListener) SetPointMap(map[string]deviceAgent.PointDefine) {

}

func (h *HTTPListener) Start() error {
	return nil
}

func (h *HTTPListener) Stop() error {
	return nil
}

func init() {
	inputs.Add("http_listener", func() deviceAgent.Input {
		return &HTTPListener{}
	})
}
