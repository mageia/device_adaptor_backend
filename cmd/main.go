package main

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/logger"
	_ "deviceAdaptor/plugins/controllers/all"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
)

func main() {

	go func() {
		agent.A, _ = agent.NewAgent()
		logger.SetupLogging(agent.A.Config.Global.Debug, "")

		agent.A.Run()
		defer agent.A.Cancel()
	}()

	//go func() {
	//	for range time.Tick(time.Second) {
	//		log.Println(runtime.NumGoroutine())
	//	}
	//}()

	for {
		select {
		case <-agent.ReloadSignal:
			agent.A.Reload()
		}
	}
}
