package main

import (
	"deviceAgent.General/agent"
	"deviceAgent.General/configs"
	_ "deviceAgent.General/plugins/inputs/all"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	c := configs.NewConfig()
	c.LoadConfig("")

	ag, _ := agent.NewAgent(c)
	//ag.Connect()
	ag.Run(nil)
}
