package models

import (
	"deviceAgent.General/interfaces"
	"sync"
)

type ProcessorConfig struct {
	Name string
}
type RunningProcessor struct {
	Name string
	sync.Mutex
	Config *ProcessorConfig
	Processor interfaces.Processor
}

type RunningProcessors []*RunningProcessor
