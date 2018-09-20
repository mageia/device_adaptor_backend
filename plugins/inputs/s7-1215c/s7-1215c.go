package s7_1215c

import (
	"deviceAgent.General/interfaces"
	"deviceAgent.General/plugins/inputs"
	"log"
)

type S71215CRawData struct {
	RawData []string
}

func (*S71215CRawData) Description() string {
	return "Get S71215CRawData data by point map"
}
func (*S71215CRawData) SampleConfig() string {
	return "S71215CRawData sample config"
}

func (m *S71215CRawData) Gather(acc interfaces.Accumulator) error {
	log.Println("S71215CRawData Gather ....")
	return nil
}

func init() {
	inputs.Add("s7_1215c", func() interfaces.Input {
		return &S71215CRawData{
			RawData: []string{"test1", "test2"},
		}
	})
}
