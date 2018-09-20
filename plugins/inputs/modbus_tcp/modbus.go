package modbus_tcp

import (
	"deviceAgent.General/interfaces"
	"deviceAgent.General/plugins/inputs"
	"log"
)

type ModbusRawData struct {
	RawData []string
}

func (*ModbusRawData) Description() string {
	return "Get Modbus data by point map"
}
func (*ModbusRawData) SampleConfig() string {
	return "Modbus sample config"
}

func (m *ModbusRawData) Gather(acc interfaces.Accumulator) error {
	log.Println("Modbus Gather ....")
	return nil
}

func init() {
	inputs.Add("modbus_tcp", func() interfaces.Input {
		return &ModbusRawData{
			RawData: []string{"test1", "test2"},
		}
	})
}
