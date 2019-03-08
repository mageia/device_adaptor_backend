package snmp

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	"github.com/k-sone/snmpgo"
	"github.com/rs/zerolog/log"
)

type S struct {
	Address      string            `json:"address"`
	Version      string            `json:"version"`
	Internal     internal.Duration `json:"internal"`
	Timeout      internal.Duration `json:"timeout"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`
	client       *snmpgo.SNMP
	connected    bool
	originName   string
	quality      device_agent.Quality
	pointMap     map[string]points.PointDefine
	oidMap       map[string]*snmpgo.Oid
	oidList      []*snmpgo.Oid
	oidBulkList  []*snmpgo.Oid
}

func (s *S) Name() string {
	if s.NameOverride != "" {
		return s.NameOverride
	}
	return s.originName
}

func (s *S) Gather(acc device_agent.Accumulator) error {
	if !s.connected {
		if e := s.Start(); e != nil {
			return e
		}
	}
	fields := make(map[string]interface{})
	s.quality = device_agent.QualityGood

	defer func(snmp *S) {
		if e := recover(); e != nil {
			snmp.quality = device_agent.QualityDisconnect
			snmp.connected = false
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(snmp.Name(), fields, nil, snmp.SelfCheck())
	}(s)

	for k, v := range s.pointMap {
		switch v.PointType {
		case points.PointArray:
			pdu, e := s.client.GetBulkRequest(s.oidBulkList, int(v.Extra["nonRepeaters"].(float64)), int(v.Extra["maxRepetitions"].(float64)))
			if e != nil || pdu.ErrorStatus() != snmpgo.NoError {
				log.Error().Err(e).Interface("errorStatus", pdu.ErrorStatus()).Interface("errorIndex", pdu.ErrorIndex()).Msg("PDU Error")
				return e
			}
			value := make([]string, 0)
			for _, v := range pdu.VarBinds().MatchBaseOids(s.oidMap[k]) {
				value = append(value, v.Variable.String())
			}
			fields[k] = value

		default:
			pdu, e := s.client.GetRequest(s.oidList)
			if e != nil || pdu.ErrorStatus() != snmpgo.NoError {
				log.Error().Err(e).Interface("errorStatus", pdu.ErrorStatus()).Interface("errorIndex", pdu.ErrorIndex()).Msg("PDU Error")
				return e
			}
			fields[k] = pdu.VarBinds().MatchOid(s.oidMap[k]).Variable.String()
		}
	}
	return nil
}

func (s *S) SelfCheck() device_agent.Quality {
	return s.quality
}

func (s *S) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
	for k, v := range s.pointMap {
		if o, e := snmpgo.NewOid(v.Address); e == nil {
			s.oidMap[k] = o
			switch v.PointType {
			case points.PointArray:
				s.oidBulkList = append(s.oidBulkList, o)
			default:
				s.oidList = append(s.oidList, o)
			}
		}
	}
}

func (s *S) Start() error {
	snmp, e := snmpgo.NewSNMP(snmpgo.SNMPArguments{
		Version:   snmpgo.V2c,
		Address:   s.Address,
		Retries:   1,
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
			oidMap:     make(map[string]*snmpgo.Oid),
		}
	})
}
