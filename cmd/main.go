package main

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/logger"
	_ "deviceAdaptor/plugins/controllers/all"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

func main() {
	logger.SetupLogging(true, "")

	go func() {
		address := ":8080"
		if runtime.GOOS == "linux" {
			address = ":80"
		}

		ConfigServer := &http.Server{
			Addr:    address,
			Handler: agent.InitRouter(true),
		}
		gin.SetMode(gin.ReleaseMode)
		ConfigServer.ListenAndServe()
	}()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//go func() {
	//	for range time.Tick(time.Second) {
	//		log.Println(runtime.NumGoroutine())
	//	}
	//}()

	go func() {
		var e error
		agent.A, e = agent.NewAgent()
		if e != nil {
			log.Println(e)
			return
		}

		logger.SetupLogging(agent.A.Config.Global.Debug, "")

		agent.A.Run()
		//defer agent.A.Cancel()
	}()

	for {
		select {
		case <-agent.ReloadSignal:
			agent.A.Reload()
			//agent.A.Cancel()
		}
	}
}
