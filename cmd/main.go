package main

import (
	"deviceAdaptor/agent"
	_ "deviceAdaptor/plugins/controllers/all"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
	"log"
	"runtime"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go func() {
		agent.A, _ = agent.NewAgent()
		agent.A.Run()
		defer agent.A.Cancel()
	}()

	go func() {
		for range time.Tick(time.Second) {
			log.Println(runtime.NumGoroutine())
		}
	}()

	for {
		select {
		case <-agent.ReloadSignal:
			agent.A.Reload()
		}
	}
}
