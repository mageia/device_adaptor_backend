package s7

import (
	"bytes"
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/utils"
	"encoding/binary"
	"fmt"
	"github.com/robinson/gos7"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type S7 struct {
	Address string
	Rack    int
	Slot    int

	client    gos7.Client
	_handler  *gos7.TCPClientHandler
	buf       map[string][]byte
	connected bool
	pointMap  map[string]deviceAgent.PointDefine
	addrMap   map[string]map[string][][2]int
	quality   deviceAgent.Quality

	FieldPrefix  string
	FieldSuffix  string
	NameOverride string
}

var defaultTimeout = internal.Duration{Duration: 15 * time.Second}

func (*S7) Name() string {
	return "S7"
}
func (s *S7) getParamList() map[string][3]int {
	var areaNumber, startAddr, endAddr, endOffset int
	var result = make(map[string][3]int)

	for areaType, o := range s.addrMap {
		areaNumber, _ = strconv.Atoi(areaType[2:])
		for k, v := range o {
			for _, i := range v {
				if i[0] > endAddr {
					endAddr = i[0]
					switch k[2:] {
					case "d":
						endOffset = 4
					case "w":
						endOffset = 2
					case "x":
						endOffset = 1
					}
				}
				if i[0] < startAddr {
					startAddr = i[0]
				}
			}
		}
		result[areaType] = [3]int{areaNumber, startAddr, endAddr + endOffset - startAddr}
	}

	return result
}

func (s *S7) gatherServer(acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	s.quality = deviceAgent.QualityGood

	defer func(s7 *S7) {
		if e := recover(); e != nil {
			s7.quality = deviceAgent.QualityDisconnect
			acc.AddError(fmt.Errorf("%v", e))
		}
		if s7.NameOverride != "" {
			acc.AddFields(s7.NameOverride, fields, tags, s7.SelfCheck())
		} else {
			acc.AddFields("s7", fields, tags, s7.SelfCheck())
		}
	}(s)

	paramMap := s.getParamList()
	for k, v := range paramMap {
		switch k[:2] {
		case "db":
			s.buf[k] = make([]byte, v[2])
			s.client.AGReadDB(v[0], v[1], v[2], s.buf[k])
		}
	}


	for area, o := range s.addrMap {
		for dataType, addrList := range o {
			for _, addr := range addrList {
				key := area + "." + dataType + fmt.Sprintf("%d", addr[0])
				switch dataType[2:] {
				case "w":
					valueByteArr := s.buf[area][addr[0]-paramMap[area][1] : addr[0]-paramMap[area][1]+2]
					//log.Println(key, binary.BigEndian.Uint16(valueByteArr))
					fields[key] = binary.BigEndian.Uint16(valueByteArr)
				case "d":
					valueByteArr := s.buf[area][addr[0]-paramMap[area][1] : addr[0]-paramMap[area][1]+4]
					var v float32
					binary.Read(bytes.NewReader(valueByteArr), binary.BigEndian, &v)
					//log.Println(key, v)
					fields[key] = v
				case "x":
					key = area + "." + dataType + fmt.Sprintf("%d.%d", addr[0], addr[1])
					valueByteArr := s.buf[area][addr[0]-paramMap[area][1]]
					//log.Println(key, utils.GetBit([]byte{valueByteArr}, uint(addr[1])))
					fields[key] = utils.GetBit([]byte{valueByteArr}, uint(addr[1]))
				}
			}
		}
	}

	return nil
}

func (s *S7) Gather(acc deviceAgent.Accumulator) error {
	if !s.connected {
		if e := s.Start(); e != nil {
			return e
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := s.gatherServer(acc); e != nil {
			acc.AddError(e)
			s.Stop()
			s.Start()
		}
	}()
	wg.Wait()
	return nil
}

func (s *S7) Start() error {
	handler := gos7.NewTCPClientHandler(s.Address, s.Rack, s.Slot)
	handler.IdleTimeout = defaultTimeout.Duration
	handler.Timeout = defaultTimeout.Duration
	//handler.Logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	s._handler = handler

	if e := handler.Connect(); e != nil {
		return e
	}
	s.client = gos7.NewClient(handler)
	s.connected = true
	return nil
}

func (s *S7) Stop() error {
	if s.connected {
		s._handler.Close()
		s.connected = false
	}
	return nil
}

func (s *S7) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	s.pointMap = pointMap
	for a := range s.pointMap {
		addrSplit := strings.SplitN(strings.TrimSpace(a), ".", 2)
		if len(addrSplit) < 2 {
			log.Println("Invalid address format, continue")
			continue
		}

		if _, ok := s.addrMap[addrSplit[0]]; !ok {
			s.addrMap[addrSplit[0]] = make(map[string][][2]int)
		}

		offsetSplit := strings.Split(addrSplit[1], ".")
		bit := -1
		if len(offsetSplit) == 2 {
			bit, _ = strconv.Atoi(offsetSplit[1])
		}
		offset, _ := strconv.Atoi(offsetSplit[0][3:])
		s.addrMap[addrSplit[0]][addrSplit[1][:3]] = append(s.addrMap[addrSplit[0]][addrSplit[1][:3]], [2]int{offset, bit})
	}
}

func (s *S7) FlushPointMap(acc deviceAgent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range s.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields("s7_point_map", pointMapFields, nil, s.SelfCheck())
	return nil
}

func (s *S7) SelfCheck() deviceAgent.Quality {
	return s.quality
}

func init() {
	inputs.Add("s7", func() deviceAgent.Input {
		return &S7{
			buf:     make(map[string][]byte),
			addrMap: make(map[string]map[string][][2]int),
			quality: deviceAgent.QualityGood,
		}
	})
}
