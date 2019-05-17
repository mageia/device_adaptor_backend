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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type S7 struct {
	Address string `json:"address"`
	Rack    int    `json:"rack"`
	Slot    int    `json:"slot"`

	client     gos7.Client
	handler    *gos7.TCPClientHandler
	buf        map[string][]byte
	connected  bool
	pointMap   map[string]points.PointDefine
	addrMap    map[string]map[int]map[string]utils.OffsetBitPair
	quality    device_adaptor.Quality
	acc        device_adaptor.Accumulator
	originName string

	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

var defaultTimeout = internal.Duration{Duration: 3 * time.Second}

func (s *S7) CheckGatherServer(acc device_adaptor.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	s.quality = device_adaptor.QualityGood

	defer func(s7 *S7) {
		if e := recover(); e != nil {
			debug.PrintStack()
			s7.quality = device_adaptor.QualityDisconnect
			s7.Stop()
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(s7.Name(), fields, tags, s7.SelfCheck())
	}(s)

	for areaType, v := range s.addrMap {
		for areaIndex, vv := range v {
			for valueType, vvv := range vv {
				sort.Sort(vvv)

				readOffset := 4
				startAddr := vvv[0][0].(int)
				endAddr := vvv[len(vvv)-1][0].(int)

				switch strings.ToLower(valueType) {
				case "", "dbb", "b", "dbx", "x":
					readOffset = 1
				case "dbd", "d":
					readOffset = 4
				case "dbw", "w":
					readOffset = 2
				}

				bufKey := fmt.Sprintf("%s_%d_%s", areaType, areaIndex, valueType)
				if len(s.buf[bufKey]) != endAddr+readOffset-startAddr {
					s.buf[bufKey] = make([]byte, endAddr+readOffset-startAddr)
				}

				//log.Debug().Str("bufKey", bufKey).Int("startAddr", startAddr).Int("endAddr", endAddr).Int("readOffset", readOffset).Msg("parse")
				switch strings.ToLower(areaType) {
				case "i":
				case "q":
				case "m":
					if e := s.client.AGReadMB(startAddr, endAddr-startAddr+readOffset, s.buf[bufKey]); e != nil {
						log.Error().Err(e).Msg("AGReadMB")
						return e
					}
				case "db":
					if e := s.client.AGReadDB(areaIndex, startAddr, endAddr+readOffset-startAddr, s.buf[bufKey]); e != nil {
						log.Error().Err(e).Msg("AGReadDB")
						return e
					}
				}

				for _, offsetBitPair := range vvv {
					valueByteArr := s.buf[bufKey][offsetBitPair[0].(int)-startAddr : offsetBitPair[0].(int)-startAddr+readOffset]
					//log.Debug().Bytes("buf", s.buf[bufKey]).Bytes("valueByteArr", valueByteArr).Msg("valueByteArr")
					switch strings.ToLower(valueType) {
					case "", "dbx", "x":
						fields[offsetBitPair[2].(string)] = utils.GetBit(valueByteArr, uint(offsetBitPair[1].(int))) == 1
					case "dbb", "b":
						var val uint8
						binary.Read(bytes.NewReader(valueByteArr), binary.BigEndian, &val)
						fields[offsetBitPair[2].(string)] = val
					case "dbw", "w":
						fields[offsetBitPair[2].(string)] = binary.BigEndian.Uint16(valueByteArr)
					case "dbd", "d":
						switch s.pointMap[offsetBitPair[2].(string)].PointType {
						case points.PointInteger:
							var val uint32
							binary.Read(bytes.NewReader(valueByteArr), binary.BigEndian, &val)
							fields[offsetBitPair[2].(string)] = val
						default:
							var val float32
							binary.Read(bytes.NewReader(valueByteArr), binary.BigEndian, &val)
							fields[offsetBitPair[2].(string)] = val
						}
					}
				}
			}
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
func (s *S7) CheckGather(acc device_adaptor.Accumulator) error {
	if !s.connected {
		if e := s.Start(); e != nil {
			return e
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := s.CheckGatherServer(acc); e != nil {
			acc.AddError(e)
			s.Stop()
		}
	}()
	wg.Wait()
	return nil
}
func (s *S7) Start() error {
	handler := gos7.NewTCPClientHandler(s.Address, s.Rack, s.Slot)
	handler.IdleTimeout = defaultTimeout.Duration * 100
	handler.Timeout = defaultTimeout.Duration
	//handler.Logger = log2.New(os.Stdout, "tcp: ", log2.LstdFlags)
	s.handler = handler

	if e := handler.Connect(); e != nil {
		return e
	}
	s.client = gos7.NewClient(handler)
	s.connected = true
	return nil
}
func (s *S7) Stop() {
	if s.connected {
		s.handler.Close()
		s.connected = false
	}
}

func (s *S7) parseAddress(pointKey, addr string) bool {
	patternDB := `(?P<areaType>db|DB)(?P<areaIndex>\d+)\.(?P<valueType>[dDwWbBxX]+)(?P<offset>\d+)\.?(?P<bit>\d*)`
	patternM := `(?P<areaType>m|M)(?P<areaIndex>)(?P<valueType>[dDwWbBxX]*)(?P<offset>\d*)\.?(?P<bit>\d*)`

	rDB := regexp.MustCompile(patternDB)
	rM := regexp.MustCompile(patternM)

	result := rDB.FindStringSubmatch(addr)
	if len(result) <= 1 {
		result = rM.FindStringSubmatch(addr)
		if len(result) <= 1 {
			return false
		}
	}
	areaType := result[1]
	areaIndex, e := strconv.Atoi(result[2])
	if e != nil {
		areaIndex = -1
	}
	valueType := result[3]
	offset, e := strconv.Atoi(result[4])
	if e != nil {
		offset = -1
	}
	bit, e := strconv.Atoi(result[5])
	if e != nil {
		bit = -1
	}

	if _, ok := s.addrMap[areaType]; !ok {
		s.addrMap[areaType] = make(map[int]map[string]utils.OffsetBitPair)
	}
	if _, ok := s.addrMap[areaType][areaIndex]; !ok {
		s.addrMap[areaType][areaIndex] = make(map[string]utils.OffsetBitPair)
	}

	s.addrMap[areaType][areaIndex][valueType] = append(s.addrMap[areaType][areaIndex][valueType], [3]interface{}{offset, bit, pointKey})

	return true
}
func (s *S7) SetPointMap(pointMap map[string]points.PointDefine) {
	s.pointMap = pointMap

	for pointKey, a := range pointMap {
		if !s.parseAddress(pointKey, a.Address) {
			log.Error().Str("address", a.Address).Str("plugin", s.Name()).Msg("parseAddress Error")
			continue
		}
	}
}
func (s *S7) FlushPointMap(acc device_adaptor.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range s.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields("s7_point_map", pointMapFields, nil, s.SelfCheck())
	return nil
}
func (s *S7) SelfCheck() device_adaptor.Quality {
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
	inputs.Add("s7", func() device_adaptor.Input {
		return &S7{
			originName: "s7",
			buf:        make(map[string][]byte),
			addrMap:    make(map[string]map[int]map[string]utils.OffsetBitPair),
			quality:    device_adaptor.QualityGood,
		}
	})
}
