package configs

import (
	"deviceAgent.General/internal/models"
	"deviceAgent.General/plugins/inputs"
	"fmt"
	"time"
)

type Config struct {
	Tags          map[string]string
	Agent      *AgentConfig
	Inputs     []*models.RunningInput
	Outputs    []*models.RunningOutput
}

func NewConfig() *Config {
	c := &Config{
		Agent: &AgentConfig{
			Interval: 10 * time.Second,
		},
		Tags:          make(map[string]string),
		Inputs:        make([]*models.RunningInput, 0),
		Outputs:       make([]*models.RunningOutput, 0),
		//Processors:    make([]*models.RunningProcessor, 0),
		//InputFilters:  make([]string, 0),
		//OutputFilters: make([]string, 0),
	}
	return c
}

type AgentConfig struct {
	Interval time.Duration
}

func (c *Config) LoadConfig(path string) error {
	c.addInput("modbus_tcp")
	c.addInput("s7_1215c")
	return nil
}

func (c *Config) addInput (name string) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("Undefined but requested input: %s", name)
	}
	input := creator()

	pluginConfig, err := buildInput(name)
	if err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) addOutput (name string) error {
	return nil
}

func buildInput(name string) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}
	return cp, nil
}