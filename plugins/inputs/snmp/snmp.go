package snmp

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	g "git.leaniot.cn/publicLib/go-snmp"
	"github.com/rs/zerolog/log"
	"time"
)

type S struct {
	Address      string            `json:"address"`
	Version      string            `json:"version"`
	Internal     internal.Duration `json:"internal"`
	Timeout      internal.Duration `json:"timeout"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`
	client       *g.GoSNMP
	connected    bool
	originName   string
	quality      device_agent.Quality
	pointMap     map[string]points.PointDefine
	oidList      []string
	oidBulkList  []string
	oidMap       map[string]string
}

func (s *S) Name() string {
	if s.NameOverride != "" {
		return s.NameOverride
	}
	return s.originName
}

func (s *S) CheckGather(acc device_agent.Accumulator) error {
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

	if r, e := s.client.Get(s.oidList); e == nil {
		for _, v := range r.Variables {
			pointKey := s.oidMap[v.Name]
			switch v.Type {
			case g.OctetString:
				fields[pointKey] = v.Value.([]byte)
			default:
				fields[pointKey] = g.ToBigInt(v.Value)
			}
		}
	}

	for _, v := range s.oidBulkList {
		walkResult := make(map[string]interface{})
		if r, e := s.client.WalkAll(v); e == nil {
			for _, pV := range r {
				switch pV.Type {
				case g.OctetString:
					walkResult[pV.Name] = pV.Value.([]byte)

					//log.Debug().Str("name", pV.Name).Interface("v", net.HardwareAddr(pV.Value.([]byte)).String()).Msg("name")
				case g.IPAddress:
					walkResult[pV.Name] = pV.Value.(string)
				default:
					walkResult[pV.Name] = g.ToBigInt(pV.Value)
				}
			}
			fields[s.oidMap[v]] = walkResult
		}
	}

	return nil
}

func (s *S) SelfCheck() device_agent.Quality {
	return s.quality
}

func (s *S) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
	for _, v := range s.pointMap {
		s.oidMap[v.Address] = v.PointKey
		switch v.PointType {
		case points.PointArray:
			s.oidBulkList = append(s.oidBulkList, v.Address)
		default:
			s.oidList = append(s.oidList, v.Address)
		}
	}
}

func (s *S) Start() error {
	s.client = &g.GoSNMP{
		Target:    s.Address,
		Port:      161,
		Community: "public",
		Version:   g.Version2c,
		Timeout:   time.Second * 2,
		Retries:   3,
		MaxOids:   g.MaxOids,
	}
	if e := s.client.Connect(); e != nil {
		return e
	}
	s.connected = true
	return nil
}

func (s *S) Stop() {
	if s.connected {
		s.client.Conn.Close()
		s.connected = false
	}
}

func (s *S) walkCallback(p g.SnmpPDU) error {
	log.Debug().Str("name", p.Name).Interface("type", p.Type).Interface("value", p.Value).Msg("pdu")
	return nil
}

func init() {
	inputs.Add("snmp", func() device_agent.Input {
		return &S{
			originName: "snmp",
			quality:    device_agent.QualityGood,
			oidMap:     make(map[string]string),
		}
	})
}
