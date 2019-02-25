package modbus

import (
	"device_adaptor"
	"device_adaptor/internal"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/utils"
	"encoding/binary"
	"errors"
	"fmt"
	"git.leaniot.cn/publicLib/go-modbus"
	"github.com/rs/zerolog/log"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HoleWidth = 200

var defaultTimeout = internal.Duration{Duration: 3 * time.Second}

type Modbus struct {
	Address string `json:"address"`
	SlaveId int    `json:"slave_id"`

	client             modbus.Client
	_handler           *modbus.TCPClientHandler
	connected          bool
	pointMap           map[string]points.PointDefine
	_pointAddressToKey map[string]string
	addrMap            map[string]map[int][]int
	addrMapKeys        map[string][]int
	quality            device_agent.Quality

	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func getParamList(addrList []int, HoleWidth int, WinWidth int) [][2]int {
	lastIndex := -1
	lastStartAddress := 0
	R := make([][2]int, 0)

	for i, d := range addrList {
		if lastIndex == -1 {
			lastStartAddress = d
			lastIndex = i
		}

		if i == len(addrList)-1 || //遍历结束，剩余的必然小于窗口宽度
			(addrList[i+1]-d > HoleWidth) || //相邻两个地址相差大于空洞大小
			(lastIndex >= 0 && (addrList[i+1]-addrList[lastIndex] > WinWidth-1)) { //左边界确定，当前游标的下一个地址与左边界地址距离超过窗口宽度

			R = append(R, [2]int{lastStartAddress, addrList[i] - addrList[lastIndex] + 1})
			lastIndex = -1
		}
	}
	return R
}
func (m *Modbus) parseAddress(address string) (area, base, bit string, err error) {
	addrSplit := strings.Split(address, "x")
	if len(addrSplit) != 2 || len(addrSplit[0]) != 1 {
		err = errors.New("invalid address format")
		return
	}
	area = addrSplit[0]

	secondSplit := strings.Split(addrSplit[1], ".")
	if len(secondSplit) == 1 {
		base = fmt.Sprintf("%04s", addrSplit[1])
		bit = ""
	} else if len(secondSplit) == 2 {
		base = fmt.Sprintf("%04s", secondSplit[0])
		bit = secondSplit[1]
	} else {
		err = errors.New("invalid address format")
	}

	return
}

func (m *Modbus) gatherServer(acc device_agent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	rawData := make(map[string][][]interface{})
	rawDataMux := sync.Mutex{}
	m.quality = device_agent.QualityGood

	defer func(md *Modbus) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
			md.quality = device_agent.QualityDisconnect
			md.Stop()
			trace := make([]byte, 2048)
			runtime.Stack(trace, true)
			log.Error().Msgf("Input [modbus] panicked: %s, Stack:\n%s\n", e, trace)
		}
		acc.AddFields(md.Name(), fields, tags, md.SelfCheck())
	}(m)

	var wg sync.WaitGroup

	for k := range m.addrMapKeys {
		switch k {
		case "0", "1":
			pList := getParamList(m.addrMapKeys[k], HoleWidth, 1500)
			rawData[k] = make([][]interface{}, len(pList))
			wg.Add(len(pList))
			for taskIdx, param := range pList {
				go func(taskIdx int, k string, param [2]int) {
					defer wg.Done()

					r, e := m.client.ReadDiscreteInputs(uint16(param[0]), uint16(param[1]))
					if e != nil {
						log.Error().Err(e)

						m.quality = device_agent.QualityDisconnect
						m.Stop()
						return
					}

					rawDataMux.Lock()
					for i := 0; i < utils.MinInt(len(r)*8, param[1]); i++ {
						rawData[k][taskIdx] = append(rawData[k][taskIdx], utils.GetBit(r, uint(i)))
					}
					rawDataMux.Unlock()
				}(taskIdx, k, param)
			}

		case "4":
			pList := getParamList(m.addrMapKeys[k], HoleWidth, 100)
			rawData[k] = make([][]interface{}, len(pList))
			wg.Add(len(pList))
			for taskIdx, param := range pList {
				go func(taskIdx int, k string, param [2]int) {
					defer wg.Done()

					r, e := m.client.ReadHoldingRegisters(uint16(param[0]), uint16(param[1]))
					if e != nil {
						log.Error().Err(e)

						m.quality = device_agent.QualityDisconnect
						m.Stop()
						return
					}
					rawDataMux.Lock()
					for i := 0; i < len(r); i += 2 {
						rawData[k][taskIdx] = append(rawData[k][taskIdx], int16(binary.BigEndian.Uint16(r[i:i+2])))
					}
					rawDataMux.Unlock()
				}(taskIdx, k, param)
			}
		}
	}

	wg.Wait()
	tmpDataMap := make(map[string][]interface{})
	for k := range rawData {
		for _, taskResult := range rawData[k] {
			tmpDataMap[k] = append(tmpDataMap[k], taskResult...)
		}
	}

	for k, l := range m.addrMapKeys {
		if len(m.addrMapKeys[k]) > len(tmpDataMap[k]) {
			continue
		}

		x1, x4 := 0, 0
		for i, a := range l {
			pointAddr := m.FieldPrefix + fmt.Sprintf("%sx%04d", k, a) + m.FieldSuffix
			switch k {
			case "0", "1":
				if i > 0 && a-l[i-1] < HoleWidth {
					x1 = a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr] = tmpDataMap[k][i+x1].(byte)
			case "4":
				if i > 0 && a-l[i-1] < HoleWidth {
					x4 = a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}

				if p, ok := m.addrMap[k][a]; ok {
					pA := fmt.Sprintf("%sx%04d", k, a)

					for _, bit := range p {
						if bit == -1 {
							if key, ok := m._pointAddressToKey[pA]; ok {
								fields[m.FieldPrefix+key+m.FieldSuffix] = m.TranslateParameter(pointAddr, tmpDataMap[k][i+x4].(int16))
							}
						} else {
							if key, ok := m._pointAddressToKey[fmt.Sprintf("%s.%d", pA, bit)]; ok {
								fields[m.FieldPrefix+key+m.FieldSuffix] = (tmpDataMap[k][i+x4].(int16)>>uint(bit))&1
							}
						}
					}
				}
			}
		}
	}

	return nil
}
func (m *Modbus) Gather(acc device_agent.Accumulator) error {
	if !m.connected {
		if e := m.Start(); e != nil {
			return e
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		e := m.gatherServer(acc)
		if e != nil {
			acc.AddError(e)
			m.Stop()
		}
	}()

	wg.Wait()

	return nil
}
func (m *Modbus) TranslateOption(pointAddr string, source byte) interface{} {
	if _, ok := m.pointMap[pointAddr]; !ok {
		return source
	}

	if o, ok := m.pointMap[pointAddr].Option[strconv.Itoa(int(source))]; ok {
		return o
	}

	return source
}
func (m *Modbus) TranslateParameter(pointAddr string, source int16) float64 {
	parameter := m.pointMap[pointAddr].Parameter
	if parameter != 0 {
		return utils.Round(parameter*float64(source), 2)
	}
	return utils.Round(float64(source), 2)
}

func (m *Modbus) Name() string {
	if m.NameOverride != "" {
		return m.NameOverride
	}
	return m.originName
}
func (m *Modbus) OriginName() string {
	return m.originName
}
func (m *Modbus) Start() error {
	_handler := modbus.NewTCPClientHandler(m.Address)
	_handler.SlaveId = uint8(m.SlaveId)
	_handler.IdleTimeout = defaultTimeout.Duration * 100
	_handler.Timeout = defaultTimeout.Duration
	m._handler = _handler

	if e := _handler.Connect(); e != nil {
		return e
	}
	m.client = modbus.NewClient(_handler)
	m.connected = true
	return nil
}
func (m *Modbus) Stop() {
	if m.connected {
		m._handler.Close()
		m.connected = false
	}
}

func (m *Modbus) SetPointMap(pointMap map[string]points.PointDefine) {
	m.pointMap = pointMap
	m._pointAddressToKey = make(map[string]string)
	m.addrMap = make(map[string]map[int][]int)
	m.addrMapKeys = make(map[string][]int)

	for k, p := range m.pointMap {
		area, base, bit, err := m.parseAddress(p.Address)
		if err != nil {
			log.Error().Err(err).Msg("parseAddress")
			continue
		}
		if m.addrMap[area] == nil {
			m.addrMap[area] = make(map[int][]int)
		}

		readAddr, e := strconv.Atoi(base)
		if e != nil {
			log.Error().Err(e).Msg("Atoi base")
			continue
		}
		bitInt := -1
		if bit != "" {
			bitInt, e = strconv.Atoi(bit)
			if e != nil {
				log.Error().Err(e).Msg("Atoi bit")
				continue
			}
		}
		m.addrMap[area][readAddr] = append(m.addrMap[area][readAddr], bitInt)

		pointKey := fmt.Sprintf("%sx%s.%s", area, base, bit)
		if bit == "" {
			pointKey = fmt.Sprintf("%sx%s", area, base)
		}
		m._pointAddressToKey[pointKey] = k
	}
	for k, v := range m.addrMap {
		for kk := range v {
			m.addrMapKeys[k] = append(m.addrMapKeys[k], kk)
		}
		sort.Ints(m.addrMapKeys[k])
	}
}
func (m *Modbus) FlushPointMap(acc device_agent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range m.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields(m.Name()+"_point_map", pointMapFields, nil, m.SelfCheck())
	return nil
}
func (m *Modbus) SetValue(kv map[string]interface{}) error {
	var errorList []error

NEXT:
	for key, value := range kv {
		addrSplit := strings.Split(strings.TrimSpace(key), "x")
		if len(addrSplit) != 2 {
			errorList = append(errorList, fmt.Errorf("invalid point key: %s", key))
			continue NEXT
		}

		readAddr, _ := strconv.Atoi(addrSplit[1])
		switch addrSplit[0] {
		case "4":
			if v, ok := value.(float64); ok {
				_, e := m.client.WriteSingleRegister(uint16(readAddr), uint16(v))
				if e != nil {
					errorList = append(errorList, e)
					continue NEXT
				}
				//TODO: write result check
				//if binary.BigEndian.Uint16(r) == uint16(v) {
				//	return nil
				//}
				//return nil
			} else {
				errorList = append(errorList, fmt.Errorf("invalid value format: %s", value))
				continue NEXT
			}
		default:
			errorList = append(errorList, fmt.Errorf("unsupported modbus address type: %s", addrSplit[0]))
			continue NEXT
		}
	}
	if len(errorList) != 0 {
		var ss string
		for _, s := range errorList {
			ss += s.Error() + "\n"
		}
		return fmt.Errorf(ss)
	}
	return nil
}
func (m *Modbus) SelfCheck() device_agent.Quality {
	return m.quality
}
func (m *Modbus) UpdatePointMap(kv map[string]interface{}) error {
	var errorList []error

NEXT:
	for key, value := range kv {
		pD, ok := m.pointMap[key]
		if !ok {
			errorList = append(errorList, fmt.Errorf("no such point: %s\n", key))
			continue NEXT
		}

		itemList := []string{"label", "name"}
		switch value.(type) {
		case map[string]interface{}:
			for _, k := range itemList {
				if v, ok := value.(map[string]interface{})[k]; ok {
					if e := utils.SetField(&pD, strings.Title(k), v); e != nil {
						errorList = append(errorList, e)
						continue NEXT
					}
				}
			}
		}
		m.pointMap[key] = pD
	}

	if len(errorList) != 0 {
		var ss string
		for _, s := range errorList {
			ss += s.Error() + "\n"
		}
		return fmt.Errorf(ss)
	}
	m.SetPointMap(m.pointMap)
	return nil
}
func (m *Modbus) RetrievePointMap(keys []string) map[string]points.PointDefine {
	if len(keys) == 0 {
		return m.pointMap
	}
	result := make(map[string]points.PointDefine, len(keys))
	for _, key := range keys {
		if p, ok := m.pointMap[key]; ok {
			result[key] = p
		}
	}
	return result
}

func init() {
	inputs.Add("modbus", func() device_agent.Input {
		return &Modbus{
			originName: "modbus",
			quality:    device_agent.QualityGood,
		}
	})
}
