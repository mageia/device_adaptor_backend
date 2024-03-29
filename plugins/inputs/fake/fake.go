package fake

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type Fake struct {
	connected bool
	pointMap  map[string]points.PointDefine
	quality   device_adaptor.Quality

	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (f *Fake) FlushPointMap(acc device_adaptor.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range f.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields(f.Name()+"_point_map", pointMapFields, nil, f.SelfCheck())
	return nil
}

func (f *Fake) Start() error {
	return nil
}
func (f *Fake) Stop() {
	f.connected = false
}
func (f *Fake) CheckGather(acc device_adaptor.Accumulator) error {
	rand.Seed(time.Now().Unix())

	fields := make(map[string]interface{})
	tags := make(map[string]string)
	f.quality = device_adaptor.QualityGood

	defer func(fake *Fake) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(fake.Name(), fields, tags, f.SelfCheck())
	}(f)

	for k, v := range f.pointMap {
		fakeMax, fakeMaxExist := v.Extra["fakeMax"]
		fakeMin, fakeMinExist := v.Extra["fakeMin"]

		switch v.PointType {
		case points.PointAnalog:
			if fakeMaxExist && fakeMinExist {
				fields[k] = rand.Float64()*(fakeMax.(float64)-fakeMin.(float64)) + fakeMin.(float64) //never max
			} else {
				fields[k] = rand.Float64() * 100 //default [0,100)
			}
		case points.PointDigital:
			if fakeMaxExist && fakeMinExist {
				fields[k] = rand.Intn(int(fakeMax.(float64))-int(fakeMin.(float64))+1) + int(fakeMin.(float64))
			} else {
				if v.Option != nil && len(v.Option) > 0 {
					keyList := make([]int, 0, len(v.Option))
					for key := range v.Option {
						if idx, err := strconv.Atoi(key); err == nil {
							keyList = append(keyList, idx)
						}
					}
					fields[k] = keyList[rand.Intn(len(keyList))]
				} else {
					fields[k] = rand.Intn(2) //default [0,1]
				}
			}
		case points.PointInteger:
			fields[k] = rand.Intn(100)
		case points.PointString:
			fields[k] = string(rand.Intn(100))
		default:
			fields[k] = "unsupported random value type"
		}
	}

	return nil
}
func (f *Fake) SelfCheck() device_adaptor.Quality {
	return f.quality
}
func (f *Fake) SetPointMap(pointMap map[string]points.PointDefine) {
	f.pointMap = pointMap
}
func (f *Fake) Name() string {
	if f.NameOverride != "" {
		return f.NameOverride
	}
	return f.originName
}
func (f *Fake) OriginName() string {
	return f.originName
}
func (f *Fake) UpdatePointMap(map[string]interface{}) error {
	return nil
}
func (f *Fake) RetrievePointMap(keys []string) map[string]points.PointDefine {
	if len(keys) == 0 {
		return f.pointMap
	}
	result := make(map[string]points.PointDefine, len(keys))
	for _, key := range keys {
		if p, ok := f.pointMap[key]; ok {
			result[key] = p
		}
	}
	return result
}
func (f *Fake) SetValue(map[string]interface{}) error {
	time.Sleep(2 * time.Second)
	return nil
}

func init() {
	inputs.Add("fake", func() device_adaptor.Input {
		return &Fake{
			originName: "fake",
			quality:    device_adaptor.QualityGood,
		}
	})
}
