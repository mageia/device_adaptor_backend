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
	Output            device_adaptor.Output
	MetricBufferLimit int `json:"metric_buffer_limit"`
	MetricBatchSize   int `json:"metric_batch_size"`

	metrics     *buffer.Buffer
	failMetrics *buffer.Buffer

	writeMutex sync.Mutex
}

func NewRunningOutput(name string, output device_adaptor.Output, batchSize int, bufferLimit int) *RunningOutput {
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

func (ro *RunningOutput) AddMetric(m device_adaptor.Metric) {
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

func (ro *RunningOutput) Write(metrics []device_adaptor.Metric) error {
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

// 判断当前 output 是否支持点表输出
func (ro *RunningOutput) SupportsWritePointDefine() bool {
	_, ok := ro.Output.(device_adaptor.RichOutput)
	return ok
}

// 输出指定 input 的点表
func (ro *RunningOutput) WritePointDefine(pointMap device_adaptor.PointMap) {
	if o, ok := ro.Output.(device_adaptor.RichOutput); ok {
		ro.writeMutex.Lock()
		defer ro.writeMutex.Unlock()
		start := time.Now()
		err := o.WritePointMap(pointMap)
		elapsed := time.Since(start)
		if err == nil {
			log.Debug().Str("output", ro.Name).Int("wrote_count", len(pointMap.Points)).Dur("time_since", elapsed).Msg("RunningOutput.WritePoints")
		}
	}
}
