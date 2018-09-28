package s7_1215c

import (
	"deviceAdaptor"
	"deviceAdaptor/plugins/inputs"
	"log"
)

type S71215CRawData struct {
	pointMap map[string]deviceAgent.PointDefine
	RawData  []string
}

func (*S71215CRawData) Name() string {
	return "S71215C"
}

func (m *S71215CRawData) Gather(acc deviceAgent.Accumulator) error {
	log.Println("S71215CRawData Gather ....")
	return nil
}

func (m *S71215CRawData) SetPointMap(pointMap map[string]deviceAgent.PointDefine) {
	m.pointMap = pointMap
}

func (m *S71215CRawData) FlushPointMap(acc deviceAgent.Accumulator) error {
	log.Println("S71215CRawData FlushPointMap")
	return nil
}

func init() {
	inputs.Add("s7_1215c", func() deviceAgent.Input {
		return &S71215CRawData{
			RawData: []string{"test1", "test2"},
		}
	})
}
