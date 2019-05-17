package vibration

import (
	"device_adaptor"
	"device_adaptor/metric"
	"device_adaptor/utils"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
	"strings"
	"time"
)

type Parser struct {
}

func round(f float64, n int) float64 {
	pow10 := math.Pow10(n)
	return math.Trunc((f+0.5/pow10)*pow10) / pow10
}
func calcAcc(o []byte) float64 {
	if len(o) != 2 {
		return 0
	}

	if o[1] > 128 {
		f := -round(float64(0xFFFF-int(o[1])*256-int(o[0])+1)/1024, 3)
		return f
	}
	return round(float64(binary.BigEndian.Uint16(o))/1024, 3)
}

func (p *Parser) Parse([]byte) (interface{}, error) {
	return nil, nil
}
func (p *Parser) validation(cmd string, line []byte) bool {
	if len(cmd) < 7 || len(line) < 7 {
		return false
	}
	if line[0] == 0xa5 && line[1] == 0x5a && line[len(line)-2] == 0x0d && line[len(line)-1] == 0x0a {
		if int(binary.BigEndian.Uint16(line[2:4])) == len(line) {
			return true
		}
	}
	return false
}
func (p *Parser) ParseCmd(cmd string, line []byte) (interface{}, error) {
	if !p.validation(cmd, line) {
		return nil, errors.New("invalid content to parse")
	}
	cmdB, _ := hex.DecodeString(cmd)
	switch cmdB[4] {
	case 0x01, 0x02:
		return hex.EncodeToString(line[5:7]), nil
	case 0x03:
		return utils.Round(float64(binary.BigEndian.Uint16(line[5:7]))/1000, 2), nil
	case 0x04, 0x05:
		x := line[5 : 5+512]
		y := line[5+512 : 5+2*512]
		z := line[5+2*512 : 5+3*512]
		r := [3][256]float32{}
		for i := 0; i < len(x); i += 2 {
			r[0][i/2] = float32(calcAcc(x[i : i+2]))
			r[1][i/2] = float32(calcAcc(y[i : i+2]))
			r[2][i/2] = float32(calcAcc(z[i : i+2]))
		}
		return r, nil
	default:
		return nil, errors.New("unsupported cmdId")
	}

	return nil, nil
}

func (p *Parser) ParserString(line []byte) (interface{}, error) {
	dataMap := make(map[string][]interface{}, 0)

	for _, l := range strings.Split(string(line), "\n") {
		lS := strings.Replace(l, " ", "", -1)
		lS = strings.Replace(lS, "\r", "", -1)

		if strings.HasPrefix(lS, "A55A") && strings.HasSuffix(lS, "0D0A") {
			msgDat, _ := hex.DecodeString(lS[10 : len(lS)-4])

			switch lS[8:10] {
			case "01":
				if _, ok := dataMap["deviceId"]; !ok {
					dataMap["deviceId"] = make([]interface{}, 0)
				}
				dataMap["deviceId"] = append(dataMap["deviceId"], lS[10:len(lS)-4])
			case "02":
				if _, ok := dataMap["deviceUsedId"]; !ok {
					dataMap["deviceUsedId"] = make([]interface{}, 0)
				}
				dataMap["deviceUsedId"] = append(dataMap["deviceUsedId"], lS[10:len(lS)-4])
			case "03":
				if _, ok := dataMap["temperature"]; !ok {

					dataMap["temperature"] = make([]interface{}, 0)
				}
				var temp float64
				if msgDat[1] < 100 {
					temp = round(float64(msgDat[1])/100+float64(msgDat[0]), 2)
				} else {
					temp = round(float64(msgDat[1])/1000+float64(msgDat[0]), 2)
				}
				dataMap["temperature"] = append(dataMap["temperature"], temp)
			case "04":
				accTmp, _ := hex.DecodeString(lS[10 : len(lS)-4])
				x := accTmp[5 : 5+512]
				y := accTmp[5+512 : 5+2*512]
				z := accTmp[5+2*512 : 5+3*512]

				acceleration := [3][256]float32{}

				for i := 0; i < len(x); i += 2 {
					acceleration[0][i/2] = float32(calcAcc(x[i : i+2]))
					acceleration[1][i/2] = float32(calcAcc(y[i : i+2]))
					acceleration[2][i/2] = float32(calcAcc(z[i : i+2]))
				}

				if _, ok := dataMap["acceleration"]; !ok {
					dataMap["acceleration"] = make([]interface{}, 0)
				}
				dataMap["acceleration"] = append(dataMap["acceleration"], acceleration)
			case "05":
				freqTmp, _ := hex.DecodeString(lS[10 : len(lS)-4])
				x := freqTmp[5 : 5+512]
				y := freqTmp[5+512 : 5+2*512]
				z := freqTmp[5+2*512 : 5+3*512]

				frequency := [3][256]float32{}

				for i := 0; i < len(x); i += 2 {
					frequency[0][i/2] = float32(calcAcc(x[i : i+2]))
					frequency[1][i/2] = float32(calcAcc(y[i : i+2]))
					frequency[2][i/2] = float32(calcAcc(z[i : i+2]))
				}

				if _, ok := dataMap["frequency"]; !ok {
					dataMap["frequency"] = make([]interface{}, 0)
				}
				dataMap["frequency"] = append(dataMap["frequency"], frequency)
			}
		}
	}
	return dataMap, nil
}
func (p *Parser) Parse2(line []byte) ([]device_adaptor.Metric, error) {
	fields := make(map[string]interface{})
	dataMap := make(map[string][]interface{}, 0)

	for _, l := range strings.Split(string(line), "\n") {
		lS := strings.Replace(l, " ", "", -1)
		lS = strings.Replace(lS, "\r", "", -1)

		if strings.HasPrefix(lS, "A55A") && strings.HasSuffix(lS, "0D0A") {
			msgDat, _ := hex.DecodeString(lS[10 : len(lS)-4])

			switch lS[8:10] {
			case "01":
				if _, ok := dataMap["deviceId"]; !ok {
					dataMap["deviceId"] = make([]interface{}, 0)
				}
				dataMap["deviceId"] = append(dataMap["deviceId"], lS[10:len(lS)-4])
			case "02":
				if _, ok := dataMap["deviceUsedId"]; !ok {
					dataMap["deviceUsedId"] = make([]interface{}, 0)
				}
				dataMap["deviceUsedId"] = append(dataMap["deviceUsedId"], lS[10:len(lS)-4])
			case "03":
				if _, ok := dataMap["temperature"]; !ok {

					dataMap["temperature"] = make([]interface{}, 0)
				}
				var temp float64
				if msgDat[1] < 100 {
					temp = round(float64(msgDat[1])/100+float64(msgDat[0]), 2)
				} else {
					temp = round(float64(msgDat[1])/1000+float64(msgDat[0]), 2)
				}
				dataMap["temperature"] = append(dataMap["temperature"], temp)
			case "04":
				if _, ok := dataMap["acceleration"]; !ok {
					dataMap["acceleration"] = make([]interface{}, 0)
				}
				dataMap["acceleration"] = append(dataMap["acceleration"], "")
			case "05":
				if _, ok := dataMap["frequency"]; !ok {
					dataMap["frequency"] = make([]interface{}, 0)
				}
				dataMap["frequency"] = append(dataMap["frequency"], "")
			}
		}
	}
	for k, v := range dataMap {
		fields[k] = v
	}

	m, _ := metric.New("", nil, fields, device_adaptor.QualityGood, time.Now(), device_adaptor.Untyped)
	return []device_adaptor.Metric{m}, nil
}
func (p *Parser) ParseLine(line string) (device_adaptor.Metric, error) {
	return nil, nil
}

func (p *Parser) parseId(line []byte) int {
	return 0
}
func (p *Parser) parseTemp(line []byte) int {
	return 0
}
func (p *Parser) parseAcc(line []byte) int {
	return 0
}
func (p *Parser) parseFreq(line []byte) int {
	return 0
}

