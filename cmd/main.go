package main

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/configs"
	_ "deviceAdaptor/plugins/controllers/all"
	_ "deviceAdaptor/plugins/inputs/all"
	_ "deviceAdaptor/plugins/outputs/all"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	c := configs.NewConfig()
	if err := c.LoadConfig(""); err != nil {
		log.Fatal(err)
		return
	}

	ag, _ := agent.NewAgent(c)
	ag.Connect()
	defer ag.Cancel()

	//go func() {
	//	time.Sleep(time.Second * 3)
	//	ag.Cancel()
	//}()

	ag.Run()
}
