package main

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/logger"
	_ "deviceAdaptor/plugins/controllers/all"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
	"log"
)

func main() {

	go func() {
		var e error
		agent.A, e = agent.NewAgent()
		if e != nil {
			log.Println(e)
			return
		}

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
