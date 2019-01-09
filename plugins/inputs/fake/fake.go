package fake

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"fmt"
	"math/rand"
	"time"
)

type Fake struct {
	connected bool
	pointMap  map[string]points.PointDefine
	quality   device_agent.Quality
	//mockKeyList map[string]interface{}
	//mockCsvReader *csv.Reader

	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (f *Fake) FlushPointMap(acc device_agent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range f.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields(f.Name()+"_point_map", pointMapFields, nil, f.SelfCheck())
	return nil
}

func (f *Fake) Start() error {
	//fmt.Print(f,"\n")
	//sFs, e := fs.New()
	//if e != nil {
	//	return e
	//}
	//
	//_csvFile, e := sFs.Open("/configs/mock_data_opc.csv")
	//if e != nil {
	//	return e
	//}
	//_mockCsvReader := csv.NewReader(_csvFile)
	//_mockKeyList, e := _mockCsvReader.Read()
	//if e != nil {
	//	return e
	//}
	//
	//f.connected = true
	//f.mockCsvReader = _mockCsvReader
	//f.mockKeyList = _mockKeyList
	return nil
}
func (f *Fake) Stop() {
	f.connected = false
}
func (f *Fake) Gather(acc device_agent.Accumulator) error {
	//if !f.connected {
	//	f.Start()
	//}
	rand.Seed(time.Now().Unix())

	fields := make(map[string]interface{})
	tags := make(map[string]string)
	f.quality = device_agent.QualityGood

	defer func(fake *Fake) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
		}
		acc.AddFields(fake.Name(), fields, tags, f.SelfCheck())
	}(f)

	//f.mockKeyList = make(map[string]interface{})
	for k, v := range f.pointMap {
		switch v.PointType {
		case points.PointAnalog:
			fields[k] = rand.Float64() * 100
		case points.PointDigital:
			fields[k] = rand.Intn(2)
		case points.PointInteger:
			fields[k] = rand.Intn(100)
		case points.PointString:
			fields[k] = string(rand.Intn(100))
		default:
			fields[k] = "unsupported random value type"
		}
	}

	//row, e := f.mockCsvReader.Read()
	//if e != nil {
	//	if e == io.EOF {
	//		f.connected = false
	//	}
	//	panic(e)
	//}
	//
	//for i, k := range f.mockKeyList {
	//	fields[i] = k
	//}

	return nil
}
func (f *Fake) SelfCheck() device_agent.Quality {
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
	inputs.Add("fake", func() device_agent.Input {
		return &Fake{
			originName: "fake",
			quality:    device_agent.QualityGood,
		}
	})
}
