package snmp

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"github.com/k-sone/snmpgo"
	"github.com/rs/zerolog/log"
)

type S struct {
	Address      string             `json:"address"`
	Version      snmpgo.SNMPVersion `json:"version"`
	Internal     internal.Duration  `json:"internal"`
	Timeout      internal.Duration  `json:"timeout"`
	FieldPrefix  string             `json:"field_prefix"`
	FieldSuffix  string             `json:"field_suffix"`
	NameOverride string             `json:"name_override"`
	client       *snmpgo.SNMP
	connected    bool
	originName   string
	quality      device_agent.Quality
	pointMap     map[string]points.PointDefine
}

func (s *S) Name() string {
	if s.NameOverride != "" {
		return s.NameOverride
	}
	return s.originName
}

func (s *S) Gather(acc device_agent.Accumulator) error {
	log.Debug().Msg("ABCD")
	return nil
}

func (s *S) SelfCheck() device_agent.Quality {
	return s.quality
}

func (s *S) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
}

func (s *S) Start() error {
	snmp, e := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   s.Address,
		Retries:   3,
		Community: "public",
	})
	if e != nil {
		return e
	}
	if e := snmp.Open(); e != nil {
		return e
	}
	s.client = snmp
	s.connected = true
	return nil
}

func (s *S) Stop() {
	if s.connected {
		s.client.Close()
		s.connected = false
	}
}

func init() {
	inputs.Add("snmp", func() device_agent.Input {
		return &S{
			originName: "snmp",
			quality:    device_agent.QualityGood,
		}
	})

	for k, v := range inputs.Inputs {

		log.Debug().Str("k", k).Str("v", v().Name()).Msg("test")
	}
}
