package modbus_tcp

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/plugins/inputs"
	"encoding/binary"
	"fmt"
	"git.leaniot.cn/publicLib/go-modbus"
	"github.com/json-iterator/go"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const HoleWidth = 200

var defaultTimeout = internal.Duration{Duration: 15 * time.Second}

type ModbusTCP struct {
	Address string
	SlaveId int

	client    modbus.Client
	connected bool
	done      chan struct{}
	pointMap  map[string]deviceAgent.PointDefine
	addrMap   map[string][]int

	FieldPrefix  string
	FieldSuffix  string
	NameOverride string
}

func Round(f float64, n int) float64 {
	pow10 := math.Pow10(n)
	return math.Trunc((f+0.5/pow10)*pow10) / pow10
}

func GetBit(word []byte, bit uint) byte {
	return (word[bit/8]) >> (bit % 8) & 0x01
}

func MinInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
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

func (*ModbusTCP) Description() string {
	return "Get Modbus data by point map"
}
func (*ModbusTCP) SampleConfig() string {
	return "Modbus sample config"
}

func (m *ModbusTCP) Start(acc deviceAgent.Accumulator) error {
	m.done = make(chan struct{})
	m.connected = false
	return m.connect()
}

func (m *ModbusTCP) connect() error {
	_handler := modbus.NewTCPClientHandler(m.Address)
	_handler.SlaveId = uint8(m.SlaveId)
	_handler.IdleTimeout = defaultTimeout.Duration
	defer _handler.Close()

	if e := _handler.Connect(); e != nil {
		return e
	}
	m.client = modbus.NewClient(_handler)
	m.connected = true

	return nil
}

func (m *ModbusTCP) gatherServer(client modbus.Client, acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tmpDataMap := make(map[string][]interface{})
	now := time.Now().UnixNano() / 1e6
	for k, l := range m.addrMap {
		sort.Ints(l)
		switch k {
		case "1":
			for _, param := range getParamList(l, HoleWidth, 1000) {
				r, e := client.ReadDiscreteInputs(uint16(param[0]-1), uint16(param[1]))
				if e != nil {
					acc.AddError(e)
					log.Printf("ReadDiscreteInputs error: %s", e)
				}
				for i := 0; i < MinInt(len(r)*8, param[1]); i++ {
					tmpDataMap[k] = append(tmpDataMap[k], GetBit(r, uint(i)))
				}
			}
		case "4":
			for _, param := range getParamList(l, HoleWidth, 125) {
				r, e := client.ReadHoldingRegisters(uint16(param[0]-1), uint16(param[1]))
				if e != nil {
					log.Fatalln(e)
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
			case "1":
				if i > 0 && a-l[i-1]-1 <= HoleWidth {
					x1 += a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr], _ = jsoniter.MarshalToString(map[string]interface{}{
					"Value":     m.TranslateOption(pointAddr, tmpDataMap[k][i+x1].(byte)),
					"Timestamp": now,
				})
			case "4":
				if i > 0 && a-l[i-1]-1 <= HoleWidth {
					x4 += a - l[i-1] - 1 //计算并剔除被忽略的小空洞
				}
				fields[pointAddr], _ = jsoniter.MarshalToString(map[string]interface{}{
					"Value":     m.TranslateParameter(pointAddr, tmpDataMap[k][i+x4].(int16)),
					"Timestamp": now,
				})
			}
		}
	}

	if m.NameOverride != "" {
		acc.AddFields(m.NameOverride, fields, tags)
	}else{
		acc.AddFields("modbus_tcp", fields, tags)
	}
	return nil
}

func (m *ModbusTCP) Gather(acc deviceAgent.Accumulator) error {
	if !m.connected {
		if e := m.connect(); e != nil {
			return e
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(client modbus.Client) {
		defer wg.Done()
		acc.AddError(m.gatherServer(client, acc))
	}(m.client)

	wg.Wait()
	return nil
}

func (m *ModbusTCP) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	m.pointMap = pointMap
	m.addrMap = make(map[string][]int, 0)

	for a := range m.pointMap {
		addrSplit := strings.Split(a, "x")
		readAddr, _ := strconv.Atoi(addrSplit[1])
		m.addrMap[addrSplit[0]] = append(m.addrMap[addrSplit[0]], readAddr)
	}
}

func (m *ModbusTCP) TranslateOption(pointAddr string, source byte) string {
	option := m.pointMap[pointAddr].Option
	sourceStr := strconv.Itoa(int(source))
	if option != nil && option[sourceStr] != "" {
		return option[sourceStr]
	}

	return sourceStr
}

func (m *ModbusTCP) TranslateParameter(pointAddr string, source int16) float64 {
	parameter := m.pointMap[pointAddr].Parameter
	if parameter != 0 {
		return Round(parameter*float64(source), 2)
	}
	return Round(float64(source), 2)
}

func (m *ModbusTCP) FlushPointMap(acc deviceAgent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range m.pointMap {
		pointMapFields[k], _ = jsoniter.MarshalToString(v)
	}
	acc.AddFields("modbus_tcp_point_map", pointMapFields, nil)
	return nil
}

func init() {
	inputs.Add("modbus_tcp", func() deviceAgent.Input {
		return &ModbusTCP{}
	})
}
