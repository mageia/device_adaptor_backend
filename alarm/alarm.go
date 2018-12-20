package alarm

import (
	"time"
)

type Alarm struct {
	Name      string `json:"name"`
	InputName string `json:"input_name"`
	Timestamp string `json:"timestamp"`
}

var ChanAlarm chan Alarm

type RealTime struct {
	PluginName string                 `json:"plugin_name"`
	Metric     map[string]interface{} `json:"metric"`
}

var ChanRealTime chan RealTime

func init() {
	ChanAlarm = make(chan Alarm, 100)
	ChanRealTime = make(chan RealTime, 10)

	go func() {
		for range time.Tick(time.Second * 3) {
			ChanAlarm <- Alarm{
				Name:      "TestAlarm",
				Timestamp: time.Now().Format(time.RFC3339),
				InputName: "inputs.opc",
			}
		}
	}()
}
