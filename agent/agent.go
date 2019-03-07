package agent

import (
	"context"
	"device_adaptor"
	"device_adaptor/configs"
	"device_adaptor/internal"
	"device_adaptor/internal/models"
	"device_adaptor/internal/points"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"runtime"
	"sync"
	"time"
)

var A *Agent

var Signal = make(chan interface{}, 1)

// 整体重载信号
type ReloadSignal struct {
}

// 点表热更新信号
type PointDefineUpdateSignal struct {
	Input device_agent.Input
}

type Agent struct {
	Ctx     context.Context
	Cancel  context.CancelFunc
	Config  *configs.Config
	Version string
	Name    string
}

func NewAgent() (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := configs.NewConfig()

	e := c.LoadConfigJson(configs.GetConfigContent())
	if e != nil {
		log.Error().Err(e)
	}

	if !c.Global.Debug {
		log.Logger = log.Level(zerolog.InfoLevel)
	} else {
		log.Logger = log.Level(zerolog.DebugLevel)
	}

	a := &Agent{
		Name:    "device_adaptor",
		Version: "v1.0.0",
		Ctx:     ctx,
		Cancel:  cancel,
		Config:  c,
	}
	return a, nil
}

func (a *Agent) Reload() {
	A.Cancel()
	log.Info().Msg("I! Reloading main program ... ")

	go func() {
		A, _ = NewAgent()
		A.Run()
	}()
}

// 点表更新回调。有两种触发机会：
//   1. 程序启动或整体 reload 时，必须假设点表更新过了
//   2. 当 agent 明确收到 PointDefineUpdateSignal 信号时
func (a *Agent) OnPointDefineUpdate(input device_agent.Input) {
	log.Info().Str("input", input.Name()).Msg("Point definition is updated")

	needOutput := false
	for _, ro := range a.Config.Outputs {
		if ro.SupportsWritePointDefine() {
			needOutput = true
			break
		}
	}

	if needOutput {
		// 取当前点表
		pointArray := make([]points.PointDefine, 0)
		pointMap := make(map[string]points.PointDefine)
		points.SqliteDB.Where("input_name = ?", input.Name()).Find(&pointArray)
		for _, p := range pointArray {
			pointMap[p.PointKey] = p
		}

		m := device_agent.PointMap{Time: time.Now(), InputName: input.Name(), Points: pointMap}

		for _, ro := range a.Config.Outputs {
			ro.WritePointDefine(m)
		}
	}
}

func (a *Agent) Close() error {
	var err error
	for _, o := range a.Config.Outputs {
		err = o.Output.Close()
		log.Info().Msgf("Successfully closed output: %s", o.Name)
	}
	for _, input := range a.Config.Inputs {
		switch p := input.Input.(type) {
		case device_agent.InteractiveInput:
			p.Stop()
			log.Info().Msgf("Successfully closed input: %s", p.Name())
		}
	}

	return err
}

func panicRecover(input *models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Info().Msgf("FATAL: Input [%s] panicked: %s, Stack:\n%s", input.Name(), err, trace)
	}
}

func gatherWithTimeout(ctx context.Context, input *models.RunningInput, acc device_agent.Accumulator, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		//start := time.Now()
		done <- input.Input.Gather(acc)
		//elapsed := time.Since(start)
		//log.Debug().Msg(time.Since(start).String())
	}()

	for {
		select {
		case err := <-done:
			if err != nil {
				acc.AddError(err)
			}
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			acc.AddError(fmt.Errorf("took longer to collect than collection interval (%s)", timeout))
		}
	}
}

func (a *Agent) gatherer(input *models.RunningInput, interval time.Duration, metricC chan device_agent.Metric) {
	defer panicRecover(input)

	acc := NewAccumulator(input, metricC)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		internal.RandomSleep(a.Config.Global.CollectionJitter.Duration, a.Ctx)
		gatherWithTimeout(a.Ctx, input, acc, interval)
		select {
		case <-a.Ctx.Done():
			return
		case <-ticker.C:
			continue
		}
	}
}

func (a *Agent) flush() {
	var wg sync.WaitGroup
	wg.Add(len(a.Config.Outputs))
	for _, o := range a.Config.Outputs {
		go func(output *models.RunningOutput) {
			defer wg.Done()
			if err := output.WriteCached(); err != nil {
				log.Info().Msgf("Error writing to output [%s]: %s", output.Name, err.Error())
			}
		}(o)
	}
}

func (a *Agent) flusher(inMetricC chan device_agent.Metric, outMetricC chan device_agent.Metric) error {
	var wg sync.WaitGroup

	// 从input channel 读数据并传给 output channel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.Ctx.Done():
				return
			case metric := <-inMetricC:
				metrics := []device_agent.Metric{metric}
				for _, metric := range metrics {
					log.Debug().Str("input", metric.Name()).Int("fields_count", len(metric.Fields())).Msg("Agent.flusher")
					outMetricC <- metric
				}
			}
		}
	}()

	// 从output channel 读数据传给 各个output组件
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.Ctx.Done():
				return
			case metric := <-outMetricC:
				for i, o := range a.Config.Outputs {
					if i == len(a.Config.Outputs)-1 {
						o.AddMetric(metric)
					} else {
						o.AddMetric(metric.Copy())
					}
				}
			}
		}
	}()

	// randomSleep??
	ticker := time.NewTicker(a.Config.Global.FlushInterval.Duration)
	semaphore := make(chan struct{})
	for {
		select {
		case <-a.Ctx.Done():
			a.flush()
			return nil
		case <-ticker.C:
			go func() {
				select {
				case semaphore <- struct{}{}:
					internal.RandomSleep(a.Config.Global.FlushJitter.Duration, a.Ctx)
					a.flush()
					<-semaphore
				default:
					log.Info().Msg("I! skipping a scheduled flush")
				}
			}()
		}
	}
}

func (a *Agent) Run() error {
	var wg sync.WaitGroup
	// input channel
	metricC := make(chan device_agent.Metric, 100)
	// output channel
	outMetricC := make(chan device_agent.Metric, 100)

	//flusher
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(metricC, outMetricC); err != nil {
			log.Error().Msgf("Flusher routine failed, exiting: %s", err.Error())
		}
	}()

	//controller
	for _, controller := range a.Config.Controllers {
		switch p := controller.Controller.(type) {
		case device_agent.Controller:
			if err := p.Start(a.Ctx); err != nil {
				log.Error().Msgf("Starting controller: %s failed, exiting\n%s", controller.Name, err.Error())
				return err
			}
		}
	}

	for _, o := range a.Config.Outputs {
		switch ot := o.Output.(type) {
		case device_agent.ServiceOutput:
			if err := ot.Start(); err != nil {
				log.Error().Err(err).Str("plugin", o.Name).Msg("ServiceOutput Start failed")
				return err
			}
		}
		err := o.Output.Connect()
		if err != nil {
			log.Error().Err(err).Str("plugin", o.Name).Msg("Output Connect failed, retrying in 3s")
			time.Sleep(3 * time.Second)
			err = o.Output.Connect()
			if err != nil {
				log.Error().Err(err).Str("plugin", o.Name).Msg("Cancel connect after retry")
				continue
			}
		}
		log.Info().Str("plugin", o.Name).Msg("output start success")
	}

	wg.Add(len(a.Config.Inputs))
	for _, input := range a.Config.Inputs {
		switch p := input.Input.(type) {
		case device_agent.InteractiveInput:
			if err := p.Start(); err != nil {
				log.Error().Err(err).Str("plugin", input.Name()).Msg("InteractiveInput start failed")
				break
			}
			log.Info().Str("plugin", input.Name()).Msg("InteractiveInput start success")

			switch pC := p.(type) {
			case device_agent.ControllerInput:
				for _, c := range a.Config.Controllers {
					c.Controller.RegisterInput(pC.Name(), pC)
				}
			}
		}

		// 启动时点表有可能已变更，需通知点表更新
		a.OnPointDefineUpdate(input.Input)

		inter := a.Config.Global.Interval.Duration
		if input.Config.Interval != 0 {
			inter = input.Config.Interval
		}
		go func(in *models.RunningInput, interval time.Duration) {
			defer wg.Done()
			a.gatherer(in, interval, metricC)
		}(input, inter)
	}

	wg.Wait()
	a.Close()

	return nil
}
