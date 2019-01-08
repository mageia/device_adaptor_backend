package main

import (
	"device_adaptor/agent"
	_ "device_adaptor/plugins/controllers/all"
	_ "device_adaptor/plugins/inputs/all"
	_ "device_adaptor/plugins/outputs/all"
	_ "device_adaptor/plugins/processors/all"
	"device_adaptor/router"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatCaller: func(i interface{}) string {
			l := strings.Split(i.(string), "/")
			return "[" + fmt.Sprintf("%-20s", l[len(l)-1]) + "]"
		},
	}).With().Caller().Timestamp().Logger()

	go func() {
		address := ":8080"
		if runtime.GOOS == "linux" {
			address = ":80"
		}

		ConfigServer := &http.Server{
			Addr:    address,
			Handler: router.InitRouter(true),
		}
		gin.SetMode(gin.ReleaseMode)
		ConfigServer.ListenAndServe()
	}()

	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	//go func() {
	//	for range time.Tick(time.Second) {
	//		log.Println(runtime.NumGoroutine())
	//	}
	//}()

	go func() {
		var e error
		agent.A, e = agent.NewAgent()
		if e != nil {
			log.Error().Err(e)
			return
		}

		agent.A.Run()
	}()

	for {
		select {
		case signal := <-agent.Signal:
			switch s := signal.(type) {
			case agent.ReloadSignal:
				agent.A.Reload()
				//agent.A.Cancel()

			case agent.PointDefineUpdateSignal:
				agent.A.OnPointDefineUpdate(s.Input)
			}
		}
	}
}
