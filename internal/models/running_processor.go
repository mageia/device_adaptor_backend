package models

import (
	"device_adaptor"
	"sync"
)

type RunningProcessor struct {
	Name string
	sync.Mutex
	Processor deviceAgent.Processor
}

type RunningProcessors []*RunningProcessor
