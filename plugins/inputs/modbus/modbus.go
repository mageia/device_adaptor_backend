package modbus

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/utils"
	"encoding/binary"
	"fmt"
	"git.leaniot.cn/publicLib/go-modbus"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HoleWidth = 200

var defaultTimeout = internal.Duration{Duration: 15 * time.Second}

type Modbus struct {
	Address string
	SlaveId int

	client    modbus.Client
	_handler  *modbus.TCPClientHandler
	connected bool
	pointMap  map[string]deviceAgent.PointDefine
	addrMap   map[string][]int

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

func (m *Modbus) Name() string {
	return "modbus"
}

func (m *Modbus) Start() error {
	m.connected = false
	return m.connect()
}

func (m *Modbus) Stop() error {
	if m.connected {
		m._handler.Close()
		m.connected = false
	}
	return nil
}

func (m *Modbus) connect() error {
	_handler := modbus.NewTCPClientHandler(m.Address)
	_handler.SlaveId = uint8(m.SlaveId)
	_handler.IdleTimeout = defaultTimeout.Duration
	_handler.Timeout = defaultTimeout.Duration
	m._handler = _handler

	if e := _handler.Connect(); e != nil {
		return e
	}
	m.client = modbus.NewClient(_handler)
	m.connected = true
	return nil
}

func (m *Modbus) gatherServer(acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tmpDataMap := make(map[string][]interface{})
	now := time.Now().UnixNano() / 1e6
	for k, l := range m.addrMap {
		sort.Ints(l)
		switch k {
		case "0", "1":
			for _, param := range getParamList(l, HoleWidth, 1000) {
				r, e := m.client.ReadDiscreteInputs(uint16(param[0]), uint16(param[1]))
				if e != nil {
					acc.AddError(e)
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
					acc.AddError(e)
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
				fields[pointAddr] = map[string]interface{}{
					"value":     m.TranslateOption(pointAddr, tmpDataMap[k][i+x1].(byte)),
					"timestamp": now,
					//"point_define": m.pointMap[pointAddr],
				}
			case "4":
				if i > 0 && a-l[i-1]-1 <= HoleWidth {
					x4 += a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr] = map[string]interface{}{
					"value":     m.TranslateParameter(pointAddr, tmpDataMap[k][i+x4].(int16)),
					"timestamp": now,
					//"point_define": m.pointMap[pointAddr],
				}
			}
		}
	}

	if m.NameOverride != "" {
		acc.AddFields(m.NameOverride, fields, tags)
	} else {
		acc.AddFields("modbus", fields, tags)
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

func (m *Modbus) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	m.pointMap = pointMap
	m.addrMap = make(map[string][]int, 0)

	for a := range m.pointMap {
		addrSplit := strings.Split(a, "x")
		readAddr, _ := strconv.Atoi(addrSplit[1])
		m.addrMap[addrSplit[0]] = append(m.addrMap[addrSplit[0]], readAddr)
	}
}

func (m *Modbus) TranslateOption(pointAddr string, source byte) string {
	option := m.pointMap[pointAddr].Option
	sourceStr := strconv.Itoa(int(source))
	if option != nil && option[sourceStr] != "" {
		return option[sourceStr]
	}

	return sourceStr
}

func (m *Modbus) TranslateParameter(pointAddr string, source int16) float64 {
	parameter := m.pointMap[pointAddr].Parameter
	if parameter != 0 {
		return utils.Round(parameter*float64(source), 2)
	}
	return utils.Round(float64(source), 2)
}

func (m *Modbus) FlushPointMap(acc deviceAgent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range m.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields("modbus_point_map", pointMapFields, nil)
	return nil
}

func (m *Modbus) Set(cmdId string, key string, value interface{}) error {
	//time.Sleep(10 * time.Second)

	addrSplit := strings.Split(strings.TrimSpace(key), "x")
	if len(addrSplit) != 2 {
		return fmt.Errorf("invalid point key: %s", key)
	}
	readAddr, _ := strconv.Atoi(addrSplit[1])
	switch addrSplit[0] {
	case "4":
		if v, ok := value.(float64); ok {
			r, e := m.client.WriteSingleRegister(uint16(readAddr), uint16(v))
			if e != nil {
				return e
			}
			if binary.BigEndian.Uint16(r) == uint16(v) {
				return nil
			}
			return nil
		} else {
			return fmt.Errorf("invalid value format: %s", value)
		}
	case "1":
	default:
		return fmt.Errorf("unsupported modbus address type: %s", addrSplit[0])
	}

	return nil
}

func (m *Modbus) Get(cmdId string, key string) interface{} {
	return nil
}


func (m *Modbus) UpdatePointMap(cmdId string, key string, value interface{}) error {
	pD, ok := m.pointMap[key]
	if !ok {
		return fmt.Errorf("no such point: %s\n", key)
	}

	//TODO: convert map[string]interface{} to struct by tag

	itemList := []string{"label", "name"}
	switch value.(type) {
	case map[string]interface{}:
		for _, k := range itemList {
			if v, ok := value.(map[string]interface{})[k]; ok {
				if e := utils.SetField(&pD, strings.Title(k), v); e != nil {
					return e
				}
			}
		}
	}
	m.pointMap[key] = pD
	return nil
}
func (m *Modbus) RetrievePointMap(cmdId string, key string) interface{} {
	if p, ok := m.pointMap[key]; ok {
		return p
	}
	return nil
}

func init() {
	inputs.Add("modbus", func() deviceAgent.Input {
		return &Modbus{}
	})
}
