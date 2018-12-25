package models

import (
	"device_adaptor"
	"device_adaptor/internal/buffer"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type RunningOutput struct {
	Name              string
	Output            deviceAgent.Output
	MetricBufferLimit int `json:"metric_buffer_limit"`
	MetricBatchSize   int `json:"metric_batch_size"`

	metrics     *buffer.Buffer
	failMetrics *buffer.Buffer

	writeMutex sync.Mutex
}

func NewRunningOutput(name string, output deviceAgent.Output, batchSize int, bufferLimit int) *RunningOutput {
	if bufferLimit == 0 {
		bufferLimit = 1
	}
	if batchSize == 0 {
		batchSize = 1
	}
	ro := &RunningOutput{
		Name:              name,
		Output:            output,
		metrics:           buffer.NewBuffer(batchSize),
		failMetrics:       buffer.NewBuffer(bufferLimit),
		MetricBufferLimit: bufferLimit,
		MetricBatchSize:   batchSize,
	}
	return ro
}

func (ro *RunningOutput) AddMetric(m deviceAgent.Metric) {
	if m == nil {
		return
	}
	ro.metrics.Add(m)
	if ro.metrics.Len() >= ro.MetricBatchSize {
		batch := ro.metrics.Batch(ro.MetricBatchSize)
		err := ro.Write(batch)
		if err != nil {
			ro.failMetrics.Add(batch...)
			log.Printf("E! Error writing to output [%s]: %v", ro.Name, err)
		}
	}
}

func (ro *RunningOutput) WriteCached() error {
	return nil
}

func (ro *RunningOutput) Write(metrics []deviceAgent.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	ro.writeMutex.Lock()
	defer ro.writeMutex.Unlock()
	start := time.Now()
	err := ro.Output.Write(metrics)
	elapsed := time.Since(start)
	if err == nil {
		log.Debug().Str("output", ro.Name).Int("wrote_count", len(metrics)).Dur("time_since", elapsed).Msg("RunningOutput.Write")
	}
	return nil
}
