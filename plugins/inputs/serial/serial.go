package serial

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/parsers"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/tarm/serial"
	"io"
	"time"
)

type HexString string

type Serial struct {
	Address      string            `json:"address"`
	BaudRate     int               `json:"baud_rate"`
	Interval     internal.Duration `json:"interval"`
	Timeout      internal.Duration `json:"timeout"`
	Interactive  bool              `json:"interactive"`
	StartFlag    string            `json:"start_flag"`
	StopFlag     string            `json:"stop_flag"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`

	pointMap   map[string]points.PointDefine
	originName string
	client     *serial.Port
	connected  bool
	parsers    map[string]parsers.Parser
}

func (s *Serial) SetParser(parsers map[string]parsers.Parser) {
	s.parsers = parsers
	//log.Debug().Interface("parsers", parsers).Msg("SetParser")
}

func (s *Serial) Start() error {
	if s.Address == "" {
		return fmt.Errorf("invalid serial address")
	}
	c, err := serial.OpenPort(&serial.Config{Name: s.Address, Baud: s.BaudRate, ReadTimeout: s.Timeout.Duration})
	if err != nil {
		return err
	}
	s.client = c
	s.connected = true
	return nil
}
func (s *Serial) Stop() {
	if s.connected {
		s.client.Close()
	}
}
func (s *Serial) Name() string {
	if s.NameOverride != "" {
		return s.NameOverride
	}
	return s.originName
}
func (s *Serial) OriginName() string {
	return s.originName
}
func (s *Serial) CheckGather(acc device_adaptor.Accumulator) error {
	defer func() {
		if err := recover(); err != nil {
			switch e := err.(type) {
			case error:
				acc.AddError(e)
			case string:
				acc.AddError(fmt.Errorf(e))
			default:
				acc.AddError(fmt.Errorf("error occured"))
			}
		}
	}()

	if !s.connected {
		if e := s.Start(); e != nil {
			panic(e)
		}
	}

	//log.Debug().Interface("pointMap", s.pointMap).Bool("interactive", s.Interactive).Int("parse_count", len(s.parsers)).Msg("CheckGather")

	if s.Interactive {
		fields := make(map[string]interface{})
		for _, v := range s.pointMap {
			a, e := hex.DecodeString(v.Address)
			if e != nil {
				panic(e)
			}
			if _, e = s.client.Write(a); e != nil {
				panic(e)
			}

			buf := make([]byte, 0)
			var cmdLen uint16 = 0
			for {
				b := make([]byte, 512)
				n, err := s.client.Read(b)
				if len(b) <= 0 || err == io.EOF {
					break
				}
				buf = append(buf, b[:n]...)

				//startFlag, e := hex.DecodeString(s.StartFlag)
				//if e != nil {
				//	break
				//}
				//stopFlag, e := hex.DecodeString(s.StopFlag)
				//if e != nil {
				//	break
				//}
				//if n < len(startFlag)+len(stopFlag) {
				//	break
				//}
				//
				//for i, v := range startFlag {
				//	if b[i] != v {
				//		break
				//	}
				//}
				//for i, v := range stopFlag {
				//	if b[i] != v {
				//		break
				//	}
				//}

				if n > 4 && b[0] == 0xa5 && b[1] == 0x5a {
					cmdLen = binary.BigEndian.Uint16(b[2:4])
				}

				if len(buf) == int(cmdLen) || (n > 2 && b[len(b)-1] == 0x0a && b[len(b)-2] == 0xd) {
					break
				}
			}
			if len(s.parsers) == 1 {
				for _, p := range s.parsers {
					if pV, e := p.ParseCmd(v.Address, buf); e != nil {
						log.Error().Err(e).Str("address", v.Address).Msg("ParseCmd")
					} else {
						fields[v.Name] = pV
					}
				}
			}
		}

		acc.AddFields(s.Name(), fields, nil, s.SelfCheck())
	}

	return nil
}
func (s *Serial) SelfCheck() device_adaptor.Quality {
	return device_adaptor.QualityGood
}
func (s *Serial) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
}

func init() {
	inputs.Add("serial", func() device_adaptor.Input {
		return &Serial{
			Interactive: true,
			Timeout:     internal.Duration{Duration: time.Second * 5},
			parsers:     make(map[string]parsers.Parser),
		}
	})
}
