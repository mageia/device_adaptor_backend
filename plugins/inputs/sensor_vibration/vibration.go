package sensor_vibration

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/parsers"
	"encoding/binary"
	"encoding/hex"
	"github.com/json-iterator/go"
	"io/ioutil"
	"math"
	"strings"
	"time"
)

type Vibration struct {
	mockData  map[string][]string
	done      chan struct{}
	connected bool
	parser    parsers.Parser
	count     int
}

type VibrationData struct {
	DeviceId     string
	DeviceUsedId string
	Temperature  float64
	Acc          [3][256]float32
	Freq         [3][256]float32
}

func (*Vibration) round(f float64, n int) float64 {
	pow10 := math.Pow10(n)
	return math.Trunc((f+0.5/pow10)*pow10) / pow10
}
func (v *Vibration) calcAcc1(o []byte) float64 {
	if len(o) != 2 {
		return 0
	}

	if o[1] > 128 {
		f := -v.round(float64(0xFFFF-int(o[1])*256-int(o[0])+1)/1024, 3)
		return f
	}
	return v.round(float64(binary.BigEndian.Uint16(o))/1024, 3)
}

func (*Vibration) Name() string {
	return "Sensor Vibration"
}
func (v *Vibration) Gather(acc device_agent.Accumulator) error {
	vData := VibrationData{}
	if deviceIdArr, ok := v.mockData["deviceId"]; ok {
		vData.DeviceId = strings.Join(deviceIdArr[5:15], "")
	}

	if deviceUsedId, ok := v.mockData["deviceUsedId"]; ok {
		vData.DeviceUsedId = strings.Join(deviceUsedId[5:15], "")
	}

	if TemperatureArr, ok := v.mockData["temperature"]; ok {
		var TempArr = make([]float64, len(TemperatureArr)/9)

		for i := 0; i < len(TemperatureArr)-1; i += 9 {
			h, _ := hex.DecodeString(strings.Join(TemperatureArr[i+5:i+7], ""))
			if h[1] < 100 {
				TempArr[i/9] = v.round(float64(h[1])/100+float64(h[0]), 2)
			} else {
				TempArr[i/9] = v.round(float64(h[1])/1000+float64(h[0]), 2)
			}
		}
		vData.Temperature = TempArr[v.count%len(TempArr)]
	}

	if AccArr, ok := v.mockData["acceleration"]; ok {
		accTmp, _ := hex.DecodeString(strings.TrimSpace(AccArr[v.count%len(AccArr)]))
		x := accTmp[5 : 5+512]
		y := accTmp[5+512 : 5+2*512]
		z := accTmp[5+2*512 : 5+3*512]

		for i := 0; i < len(x); i += 2 {
			vData.Acc[0][i/2] = float32(v.calcAcc1(x[i : i+2]))
			vData.Acc[1][i/2] = float32(v.calcAcc1(y[i : i+2]))
			vData.Acc[2][i/2] = float32(v.calcAcc1(z[i : i+2]))
		}
	}

	if FreqArr, ok := v.mockData["frequency"]; ok {
		freqTmp, _ := hex.DecodeString(strings.TrimSpace(FreqArr[v.count%len(FreqArr)]))
		x := freqTmp[5 : 5+512]
		y := freqTmp[5+512 : 5+2*512]
		z := freqTmp[5+2*512 : 5+3*512]
		for i := 0; i < len(x); i += 2 {
			vData.Freq[0][i/2] = float32(v.calcAcc1(x[i : i+2]))
			vData.Freq[1][i/2] = float32(v.calcAcc1(y[i : i+2]))
			vData.Freq[2][i/2] = float32(v.calcAcc1(z[i : i+2]))
		}
	}
	v.count += 1

	acceleration, _ := jsoniter.MarshalToString(vData.Acc)
	frequency, _ := jsoniter.MarshalToString(vData.Freq)

	acc.AddFields("sensor_vibration", map[string]interface{}{
		"deviceId":     vData.DeviceId,
		"deviceUsedId": vData.DeviceUsedId,
		"temperature":  vData.Temperature,
		"acceleration": acceleration,
		"frequency":    frequency,
		"timestamp":    time.Now().UnixNano() / 1e6,
	}, nil, device_agent.QualityGood)
	return nil
}
func (*Vibration) SetPointMap(map[string]points.PointDefine) {
}
func (*Vibration) FlushPointMap(device_agent.Accumulator) error {
	return nil
}
func (v *Vibration) Start() error {
	v.done = make(chan struct{})
	v.mockData = make(map[string][]string)

	b, _ := ioutil.ReadFile("../plugins/inputs/sensor_vibration/mock_data/01设备ID.txt")
	v.mockData["deviceId"] = strings.Split(string(b), " ")

	b, _ = ioutil.ReadFile("../plugins/inputs/sensor_vibration/mock_data/02使用设备ID.txt")
	v.mockData["deviceUsedId"] = strings.Split(string(b), " ")

	b, _ = ioutil.ReadFile("../plugins/inputs/sensor_vibration/mock_data/03温度数据.txt")
	v.mockData["temperature"] = strings.Split(string(b), " ")

	b, _ = ioutil.ReadFile("../plugins/inputs/sensor_vibration/mock_data/04加速度数据.txt")
	v.mockData["acceleration"] = strings.Split(strings.Join(strings.Split(string(b), " "), ""), "\n")

	b, _ = ioutil.ReadFile("../plugins/inputs/sensor_vibration/mock_data/05频谱数据.txt")
	v.mockData["frequency"] = strings.Split(strings.Join(strings.Split(string(b), " "), ""), "\n")

	v.connected = true
	return nil
}
func (*Vibration) Stop() {
}
func (*Vibration) SelfCheck() device_agent.Quality {
	return device_agent.QualityGood
}

func init() {
	inputs.Add("sensor_vibration", func() device_agent.Input {
		return &Vibration{}
	})
}
