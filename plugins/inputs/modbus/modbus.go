package modbus

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/utils"
	"encoding/binary"
	"fmt"
	"git.leaniot.cn/publicLib/go-modbus"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HoleWidth = 200

var defaultTimeout = internal.Duration{Duration: 3 * time.Second}

type Modbus struct {
	Address string
	SlaveId int

	client    modbus.Client
	_handler  *modbus.TCPClientHandler
	connected bool
	pointMap  map[string]deviceAgent.PointDefine
	addrMap   map[string][]int
	quality   deviceAgent.Quality
	acc       deviceAgent.Accumulator

	originName   string
	FieldPrefix  string
	FieldSuffix  string
	NameOverride string
}

func getParamList(addrList []int, HoleWidth int, WinWidth int) [][2]int {
	tmpR := [2]int{}
	var lastIndex = -1
	R := make([][2]int, 0)

	for i, d := range addrList {
		if lastIndex == -1 {
			tmpR[0] = addrList[i]
			lastIndex = i
		}

		if i == len(addrList)-1 || //遍历结束，剩余的必然小于窗口宽度
			(addrList[i+1]-d >= HoleWidth) || //相邻两个地址相差大于空洞大小
			(lastIndex >= 0 && (i-lastIndex+1 >= WinWidth || //左边界确定，当前游标处地址与左边界地址距离超过窗口宽度
				addrList[i+1]-addrList[lastIndex]+1 >= WinWidth)) { //左边界确定，当前游标的下一个地址与左边界地址距离超过窗口宽度

			tmpR[1] = addrList[i] - addrList[lastIndex] + 1
			R = append(R, tmpR)
			lastIndex = -1
		}
	}

	return R
}
func (m *Modbus) gatherServer(acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tmpDataMap := make(map[string][]interface{})
	m.quality = deviceAgent.QualityGood

	defer func(modbus *Modbus) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
			m.quality = deviceAgent.QualityDisconnect
			m.connected = false
		}
		if modbus.NameOverride != "" {
			acc.AddFields(modbus.NameOverride, fields, tags, modbus.SelfCheck())
		} else {
			acc.AddFields(m.Name(), fields, tags, modbus.SelfCheck())
		}
	}(m)

	for k, l := range m.addrMap {
		sort.Ints(l)
		switch k {
		case "0", "1":
			for _, param := range getParamList(l, HoleWidth, 1000) {
				r, e := m.client.ReadDiscreteInputs(uint16(param[0]), uint16(param[1]))
				if e != nil {
					m.quality = deviceAgent.QualityDisconnect
					m.connected = false
					return e
				}
				for i := 0; i < utils.MinInt(len(r)*8, param[1]); i++ {
					tmpDataMap[k] = append(tmpDataMap[k], utils.GetBit(r, uint(i)))
				}
			}
		case "4":
			for _, param := range getParamList(l, HoleWidth, 125) {
				r, e := m.client.ReadHoldingRegisters(uint16(param[0]), uint16(param[1]))
				if e != nil {
					m.quality = deviceAgent.QualityDisconnect
					m.connected = false
					return e
				}
				for i := 0; i < len(r); i += 2 {
					tmpDataMap[k] = append(tmpDataMap[k], int16(binary.BigEndian.Uint16(r[i:i+2])))
				}
			}
		}
	}

	for k, l := range m.addrMap {
		x1, x4 := 0, 0

		for i, a := range l {
			pointAddr := m.FieldPrefix + fmt.Sprintf("%sx%04d", k, a) + m.FieldSuffix
			switch k {
			case "0", "1":
				if i > 0 && a-l[i-1]-1 <= HoleWidth {
					x1 += a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr] = tmpDataMap[k][i+x1].(byte)
			case "4":
				if i > 0 && a-l[i-1]-1 <= HoleWidth {
					x4 += a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr] = m.TranslateParameter(pointAddr, tmpDataMap[k][i+x4].(int16))
			}
		}
	}

	return nil
}
func (m *Modbus) Gather(acc deviceAgent.Accumulator) error {
	if !m.connected {
		if e := m.connect(); e != nil {
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
			m.Start()
		}
	}()

	wg.Wait()

	return nil
}
func (m *Modbus) connect() error {
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
	m.connected = false
	return m.connect()
}
func (m *Modbus) Stop() {
	if m.connected {
		m._handler.Close()
		m.connected = false
	}
}
func (m *Modbus) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	m.pointMap = pointMap
	m.addrMap = make(map[string][]int, 0)

	for a := range m.pointMap {
		addrSplit := strings.Split(a, "x")
		readAddr, _ := strconv.Atoi(addrSplit[1])
		m.addrMap[addrSplit[0]] = append(m.addrMap[addrSplit[0]], readAddr)
	}
}
func (m *Modbus) FlushPointMap(acc deviceAgent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range m.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields(m.Name()+"_point_map", pointMapFields, nil, m.SelfCheck())
	return nil
}
func (m *Modbus) SetValue(kv map[string]interface{}) error {
	var errors []error

NEXT:
	for key, value := range kv {
		addrSplit := strings.Split(strings.TrimSpace(key), "x")
		if len(addrSplit) != 2 {
			errors = append(errors, fmt.Errorf("invalid point key: %s", key))
			continue NEXT
		}

		readAddr, _ := strconv.Atoi(addrSplit[1])
		switch addrSplit[0] {
		case "4":
			if v, ok := value.(float64); ok {
				log.Println(readAddr, v)
				_, e := m.client.WriteSingleRegister(uint16(readAddr), uint16(v))
				if e != nil {
					errors = append(errors, e)
					continue NEXT
				}
				//TODO: write result check
				//if binary.BigEndian.Uint16(r) == uint16(v) {
				//	return nil
				//}
				//return nil
			} else {
				errors = append(errors, fmt.Errorf("invalid value format: %s", value))
				continue NEXT
			}
		default:
			errors = append(errors, fmt.Errorf("unsupported modbus address type: %s", addrSplit[0]))
			continue NEXT
		}
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
func (m *Modbus) SelfCheck() deviceAgent.Quality {
	return m.quality
}
func (m *Modbus) UpdatePointMap(kv map[string]interface{}) error {
	var errors []error

NEXT:
	for key, value := range kv {
		pD, ok := m.pointMap[key]
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
		m.pointMap[key] = pD
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
func (m *Modbus) RetrievePointMap(keys []string) map[string]deviceAgent.PointDefine {
	if len(keys) == 0 {
		return m.pointMap
	}
	result := make(map[string]deviceAgent.PointDefine, len(keys))
	for _, key := range keys {
		if p, ok := m.pointMap[key]; ok {
			result[key] = p
		}
	}
	return result
}

func init() {
	inputs.Add("modbus", func() deviceAgent.Input {
		return &Modbus{
			originName: "modbus",
			quality:    deviceAgent.QualityGood,
		}
	})
}
