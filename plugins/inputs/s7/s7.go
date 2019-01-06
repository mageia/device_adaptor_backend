package s7

import (
	"bytes"
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/utils"
	"encoding/binary"
	"fmt"
	"github.com/robinson/gos7"
	"github.com/rs/zerolog/log"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type S7 struct {
	Address string `json:"address"`
	Rack    int    `json:"rack"`
	Slot    int    `json:"slot"`

	client             gos7.Client
	_handler           *gos7.TCPClientHandler
	buf                map[string][]byte
	connected          bool
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	addrMap            map[string]map[string][][2]int
	quality            deviceAgent.Quality
	acc                deviceAgent.Accumulator

	originName string

	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

var defaultTimeout = internal.Duration{Duration: 3 * time.Second}

func (s *S7) getParamList() map[string][3]int {
	var areaNumber, startAddr, endAddr int
	var result = make(map[string][3]int)
	var endOffset = 4

	for areaType, o := range s.addrMap {
		if strings.HasPrefix(strings.ToLower(areaType), "db") {
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
		} else if strings.HasPrefix(strings.ToLower(areaType), "m") {

		}
	}

	return result
}
func (s *S7) gatherServer(acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	s.quality = deviceAgent.QualityGood

	defer func(s7 *S7) {
		if e := recover(); e != nil {
			debug.PrintStack()
			s7.quality = deviceAgent.QualityDisconnect
			s7.connected = false
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(s7.Name(), fields, tags, s7.SelfCheck())
	}(s)

	paramMap := s.getParamList()
	log.Debug().Interface("paramMap", paramMap).Msg("getParamList")
	for k, v := range paramMap {
		switch k[:2] {
		case "db":
			if s.buf[k] == nil || len(s.buf[k]) != v[2] {
				s.buf[k] = make([]byte, v[2])
			}
			if e := s.client.AGReadDB(v[0], v[1], v[2], s.buf[k]); e != nil {
				return e
			}
		}
	}

	for area, o := range s.addrMap {
		if strings.HasPrefix(strings.ToLower(area), "db") {
			for dataType, addrList := range o {
				for _, addr := range addrList {
					key := area + "." + dataType + fmt.Sprintf("%d", addr[0])
					switch dataType[2:] {
					case "b":
					case "w":
						valueByteArr := s.buf[area][addr[0]-paramMap[area][1] : addr[0]-paramMap[area][1]+2]
						fields[key] = binary.BigEndian.Uint16(valueByteArr)
					case "d":
						valueByteArr := s.buf[area][addr[0]-paramMap[area][1] : addr[0]-paramMap[area][1]+4]
						var v float32
						binary.Read(bytes.NewReader(valueByteArr), binary.BigEndian, &v)
						fields[key] = v
					case "x":
						key = area + "." + dataType + fmt.Sprintf("%d.%d", addr[0], addr[1])
						valueByteArr := s.buf[area][addr[0]-paramMap[area][1]]
						fields[key] = utils.GetBit([]byte{valueByteArr}, uint(addr[1]))
					}
				}
			}
		} else if strings.HasPrefix(strings.ToLower(area), "m") {

		}
	}

	return nil
}

func (s *S7) Name() string {
	if s.NameOverride != "" {
		return s.NameOverride
	}
	return s.originName
}
func (s *S7) OriginName() string {
	return s.originName
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
	handler.IdleTimeout = defaultTimeout.Duration * 100
	handler.Timeout = defaultTimeout.Duration
	s._handler = handler

	if e := handler.Connect(); e != nil {
		return e
	}
	s.client = gos7.NewClient(handler)
	s.connected = true
	return nil
}
func (s *S7) Stop() {
	if s.connected {
		s._handler.Close()
		s.connected = false
	}
}
func (s *S7) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap
	s._pointAddressToKey = make(map[string]string, len(s.pointMap))

	pattern := `^(md|MD)\d+$|^[mM]\d+\.\d{1,2}$|^(db|DB)\d+\.(db|DB)[dwxDWX]\d+(.\d{1,2}?)$`

	for _, a := range pointMap {
		if ok, err := regexp.Match(pattern, []byte(a.Address)); !ok || err != nil {
			log.Error().Err(err).Bool("matched", ok).Str("addr", a.Address).Msg("reg match address")
			continue
		}

		if strings.HasPrefix(strings.ToLower(a.Address), "db") { //DB area
			addrSplit := strings.SplitN(strings.TrimSpace(a.Address), ".", 2)
			areaStr := strings.ToLower(addrSplit[0])

			if _, ok := s.addrMap[areaStr]; !ok {
				s.addrMap[areaStr] = make(map[string][][2]int)
			}

			offsetSplit := strings.Split(addrSplit[1], ".")
			bit := -1
			if len(offsetSplit) == 2 {
				bit, _ = strconv.Atoi(offsetSplit[1])
			}
			offset, _ := strconv.Atoi(offsetSplit[0][3:])
			s.addrMap[areaStr][addrSplit[1][:3]] = append(s.addrMap[areaStr][addrSplit[1][:3]], [2]int{offset, bit})
		} else if strings.HasPrefix(strings.ToLower(a.Address), "m") { //M area
			addrSplit := strings.SplitN(strings.TrimSpace(a.Address), ".", 2)
			areaStr := "m"
			if _, ok := s.addrMap[areaStr]; !ok {
				s.addrMap[areaStr] = make(map[string][][2]int)
			}

			if strings.HasPrefix(strings.ToLower(a.Address), "md") {
				offset, _ := strconv.Atoi(addrSplit[0][2:])
				s.addrMap[areaStr]["md"] = append(s.addrMap[areaStr]["md"], [2]int{offset, -1})
			} else {
				offset, _ := strconv.Atoi(addrSplit[0][1:])
				bit, _ := strconv.Atoi(addrSplit[1])
				s.addrMap[areaStr]["m"] = append(s.addrMap[areaStr]["m"], [2]int{offset, bit})
			}
		}
	}

	//log.Debug().Interface("addrMap", s.addrMap).Msg("SetPointMap")

	i := 0
	for _, v := range s.pointMap {
		s._pointAddressToKey[v.Address] = v.PointKey
		i++
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
func (s *S7) SetValue(map[string]interface{}) error {
	time.Sleep(2 * time.Second)
	return nil
}
func (s *S7) UpdatePointMap(kv map[string]interface{}) error {
	var errors []error

NEXT:
	for key, value := range kv {
		pD, ok := s.pointMap[key]
		if !ok {
			errors = append(errors, fmt.Errorf("no such point: %s\n", key))
			continue NEXT
		}

		itemList := []string{"label", "name"}
		switch value.(type) {
		case map[string]interface{}:
			for _, k := range itemList {
				if v, ok := value.(map[string]interface{})[k]; ok {
					if e := utils.SetField(&pD, strings.Title(k), v); e != nil {
						errors = append(errors, e)
						continue NEXT
					}
				}
			}
		}
		s.pointMap[key] = pD
	}

	if len(errors) != 0 {
		var ss string
		for _, s := range errors {
			ss += s.Error() + "\n"
		}
		return fmt.Errorf(ss)
	}
	return nil
}
func (s *S7) RetrievePointMap(keys []string) map[string]points.PointDefine {
	if len(keys) == 0 {
		return s.pointMap
	}
	result := make(map[string]points.PointDefine, len(keys))
	for _, key := range keys {
		if p, ok := s.pointMap[key]; ok {
			result[key] = p
		}
	}
	return result
}

func init() {
	inputs.Add("s7", func() deviceAgent.Input {
		return &S7{
			originName: "s7",
			buf:        make(map[string][]byte),
			addrMap:    make(map[string]map[string][][2]int),
			quality:    deviceAgent.QualityGood,
		}
	})
}
