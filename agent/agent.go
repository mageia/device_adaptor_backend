package agent

import (
	"deviceAgent.General/configs"
	"deviceAgent.General/internal/models"
	"log"
	"sync"
	"time"
)

type Agent struct {
	Config *configs.Config
}

func NewAgent(config *configs.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}
	return a, nil
}

func (a *Agent) Connect() error {
	for _, o := range a.Config.Outputs {
		err := o.Output.Connect()
		if err != nil {
			log.Printf("E! Failed to connect to output %s, retrying in 15s, "+
				"error was '%s'\n", o.Name, err)
			time.Sleep(15 * time.Second)
			err = o.Output.Connect()
			if err != nil {
				return err
			}
		}
		log.Printf("D! Successfully connected to output: %s\n", o.Name)
	}
	return nil
}

func (a *Agent) Close() error {
	var err error
	for _, o := range a.Config.Outputs {
		err = o.Output.Close()
		log.Printf("D! Successfully connected to output: %s\n", o.Name)
	}
	return err
}

func (a *Agent)gatherer(input *models.RunningInput, interval time.Duration)  {
	input.Input.Gather(nil)
}

func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup
	wg.Add(len(a.Config.Inputs))
	for _, input  := range a.Config.Inputs {
		go func(in *models.RunningInput, interval time.Duration) {
			defer wg.Done()
			a.gatherer(in, interval)
		}(input, 10)
	}
	wg.Wait()
	a.Close()
	return nil
}
