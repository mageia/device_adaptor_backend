package fake

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/inputs"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

type Fake struct {
	connected     bool
	pointMap      map[string]deviceAgent.PointDefine
	quality       deviceAgent.Quality
	mockKeyList   []string
	mockCsvReader *csv.Reader

	originName   string
	FieldPrefix  string
	FieldSuffix  string
	NameOverride string
}

func (f *Fake) FlushPointMap(acc deviceAgent.Accumulator) error {
	pointMapFields := make(map[string]interface{})
	for k, v := range f.pointMap {
		pointMapFields[k] = v
	}
	acc.AddFields(f.Name()+"_point_map", pointMapFields, nil, f.SelfCheck())
	return nil
}

func (f *Fake) Start() error {
	if _csvFile, e := os.Open("../configs/mock_data_opc.csv"); e != nil {
		return e
	} else {
		_mockCsvReader := csv.NewReader(_csvFile)
		_mockKeyList, e := _mockCsvReader.Read()
		if e != nil {
			return e
		}

		f.connected = true
		f.mockCsvReader = _mockCsvReader
		f.mockKeyList = _mockKeyList
	}
	return nil
}
func (f *Fake) Stop() error {
	f.connected = false
	return nil
}
func (f *Fake) Gather(acc deviceAgent.Accumulator) error {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	f.quality = deviceAgent.QualityGood

	defer func(fake *Fake) {
		if e := recover(); e != nil {
			acc.AddError(fmt.Errorf("%v", e))
		}
		if fake.NameOverride != "" {
			acc.AddFields(fake.NameOverride, fields, tags, f.SelfCheck())
		} else {
			acc.AddFields(f.Name(), fields, tags, f.SelfCheck())
		}
	}(f)

	//for k, v := range f.pointMap {
	//	if v.PointType == deviceAgent.PointState {
	//		fields[k] = rand.Intn(10) % 2
	//	} else {
	//		fields[k] = utils.Round(rand.Float64()/rand.Float64(), 2)
	//	}
	//}
	row, e := f.mockCsvReader.Read()
	if e != nil {
		if e == io.EOF {
			f.connected = false
		}
		return e
	}

	for i, k := range f.mockKeyList {
		fields[k], _ = strconv.ParseFloat(row[i], 64)
	}

	return nil
}
func (f *Fake) SelfCheck() deviceAgent.Quality {
	return f.quality
}
func (f *Fake) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
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
	panic("implement me")
}
func (f *Fake) RetrievePointMap(keys []string) map[string]deviceAgent.PointDefine {
	if len(keys) == 0 {
		return f.pointMap
	}
	result := make(map[string]deviceAgent.PointDefine, len(keys))
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
	inputs.Add("fake", func() deviceAgent.Input {
		return &Fake{
			originName: "fake",
			quality:    deviceAgent.QualityGood,
		}
	})
}
