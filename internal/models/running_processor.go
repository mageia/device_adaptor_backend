package models

import (
	"device_adaptor"
	"sync"
)

type RunningProcessor struct {
	Name string
	sync.Mutex
	Processor device_adaptor.Processor
}

type RunningProcessors []*RunningProcessor
