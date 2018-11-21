package serial

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/utils"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/tarm/serial"
	"io"
	"time"
)

type Serial struct {
	Address      string            `json:"address"`
	BaudRate     int               `json:"baud_rate"`
	Interval     internal.Duration `json:"interval"`
	Timeout      internal.Duration `json:"timeout"`
	Interactive  bool              `json:"interactive"`
	FieldPrefix  string            `json:"field_prefix"`
	FieldSuffix  string            `json:"field_suffix"`
	NameOverride string            `json:"name_override"`

	pointMap   map[string]points.PointDefine
	originName string
	client     *serial.Port
	connected  bool
}

func calcAcc(o []byte) float64 {
	if len(o) != 2 {
		return 0
	}

	if o[1] > 128 {
		f := -utils.Round(float64(0xFFFF-int(o[1])*256-int(o[0])+1)/1024, 3)
		return f
	}
	return utils.Round(float64(binary.BigEndian.Uint16(o))/1024, 3)
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
func (s *Serial) Gather(acc deviceAgent.Accumulator) error {
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
				if n > 4 && b[0] == 0xa5 && b[1] == 0x5a {
					cmdLen = binary.BigEndian.Uint16(b[2:4])
				}

				if len(buf) == int(cmdLen) || (n > 2 && b[len(b)-1] == 0x0a && b[len(b)-2] == 0xd) {
					break
				}
			}
			//fields[v.Name] = buf

			//TODO: delete later
			if len(v.Address) != 14 {
				continue
			}
			switch v.Address[8:10] {
			case "01", "02":
				fields[v.Name] = hex.EncodeToString(buf[5:7])
			case "03":
				fields[v.Name] = utils.Round(float64(binary.BigEndian.Uint16(buf[5:7])/1000), 2)
			case "04", "05":
				x := buf[5 : 5+512]
				y := buf[5+512 : 5+2*512]
				z := buf[5+2*512 : 5+3*512]
				acceleration := [3][256]float32{}

				for i := 0; i < len(x); i += 2 {
					acceleration[0][i/2] = float32(calcAcc(x[i : i+2]))
					acceleration[1][i/2] = float32(calcAcc(y[i : i+2]))
					acceleration[2][i/2] = float32(calcAcc(z[i : i+2]))
				}
				fields[v.Name] = acceleration
			default:
				continue
			}
			//TODO: delete later
		}

		acc.AddFields(s.Name(), fields, nil, s.SelfCheck())
	}

	return nil
}
func (s *Serial) SelfCheck() deviceAgent.Quality {
	return deviceAgent.QualityGood
}
func (s *Serial) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
}

func init() {
	inputs.Add("serial", func() deviceAgent.Input {
		return &Serial{
			Interactive: true,
			Timeout:     internal.Duration{Duration: time.Second * 5},
		}
	})
}
