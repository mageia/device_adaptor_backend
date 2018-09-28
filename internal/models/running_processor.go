package models

import (
	"deviceAdaptor"
	"sync"
)

type RunningProcessor struct {
	Name string
	sync.Mutex
	Processor deviceAgent.Processor
}

type RunningProcessors []*RunningProcessor
