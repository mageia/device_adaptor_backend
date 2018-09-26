package main

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/configs"
	_ "deviceAdaptor/logger"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
	"log"
)

func main() {
	c := configs.NewConfig()
	if err := c.LoadConfig(""); err != nil {
		log.Fatal(err)
	}

	ag, _ := agent.NewAgent(c)
	ag.Connect()

	ag.Run(nil)
}
