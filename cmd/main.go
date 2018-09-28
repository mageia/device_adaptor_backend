package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//go func() {
	//	time.Sleep(time.Second * 2)
	//	cancel()
	//}()

	ag.Run(ctx)
}
