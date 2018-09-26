package models

import (
	"deviceAdaptor"
	"sync"
)

type ProcessorConfig struct {
	Name string
}
type RunningProcessor struct {
	Name string
	sync.Mutex
	Config *ProcessorConfig
	Processor deviceAgent.Processor
}

type RunningProcessors []*RunningProcessor
