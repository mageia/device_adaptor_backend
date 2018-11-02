package agent

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/configs"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var A *Agent

type Agent struct {
	Ctx          context.Context
	Cancel       context.CancelFunc
	ConfigServer *http.Server
	Config       *configs.Config
	Version      string
	Name         string
}

func NewAgent() (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := configs.NewConfig()
	if err := c.LoadConfig(""); err != nil {
		return nil, err
	}

	a := &Agent{
		Name:    "deviceAdaptor",
		Version: "v1.0.0",
		Ctx:     ctx,
		Cancel:  cancel,
		Config:  c,
		ConfigServer: &http.Server{
			Addr:    ":8080",
			Handler: InitRouter(c.Global.Debug),
		},
	}
	return a, nil
}

func (a *Agent) Reload() {
	A.Cancel()
	go func() {
		A, _ = NewAgent()
		A.Run()
	}()
}

func (a *Agent) Close() error {
	var err error
	for _, o := range a.Config.Outputs {
		err = o.Output.Close()
		log.Printf("D! Successfully closed output: %s\n", o.Name)
	}

	a.ConfigServer.Shutdown(a.Ctx)
	return err
}

func panicRecover(input *models.RunningInput) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("E! FATAL: Input [%s] panicked: %s, Stack:\n%s\n", input.Name(), err, trace)
	}
}

func gatherWithTimeout(ctx context.Context, input *models.RunningInput, acc deviceAgent.Accumulator, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()

	done := make(chan error)
	go func() {
		done <- input.Input.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			if err != nil {
				acc.AddError(err)
			}
			return
		case <-ticker.C:
			err := fmt.Errorf("took longer to collect than collection interval (%s)", timeout)
			acc.AddError(err)
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) gatherer(input *models.RunningInput, interval time.Duration, metricC chan deviceAgent.Metric) {
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
				log.Printf("E! Error writing to output [%s]: %s\n", output.Name, err.Error())
			}
		}(o)
	}
}

func (a *Agent) flusher(metricC chan deviceAgent.Metric, outMetricC chan deviceAgent.Metric) error {
	var wg sync.WaitGroup
	wg.Add(1)

	// 从input channel 读数据并传给 output channel
	go func() {
		defer wg.Done()
		for {
			select {
			case <-a.Ctx.Done():
				return
			case metric := <-metricC:
				metrics := []deviceAgent.Metric{metric}
				for _, metric := range metrics {
					outMetricC <- metric
				}
			}
		}
	}()

	wg.Add(1)

	// 从output channel 读数据传给 各个output组件
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
					log.Println("I! skipping a scheduled flush")
				}
			}()
		}
	}

	return nil
}

func (a *Agent) Run() error {
	var wg sync.WaitGroup
	// input channel
	metricC := make(chan deviceAgent.Metric, 100)
	// output channel
	outMetricC := make(chan deviceAgent.Metric, 100)

	//ConfigServer
	go func() {
		gin.SetMode(gin.ReleaseMode)
		a.ConfigServer.ListenAndServe()
	}()

	//flusher
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(metricC, outMetricC); err != nil {
			log.Printf("E! Flusher roution failed, exiting: %s\n", err.Error())
		}
	}()

	//controller
	for _, controller := range a.Config.Controllers {
		switch p := controller.Controller.(type) {
		case deviceAgent.Controller:
			if err := p.Start(a.Ctx); err != nil {
				log.Printf("E! starting controller: %s failed, exiting\n%s\n", controller.Name, err.Error())
				return err
			}
		}
	}

	for _, o := range a.Config.Outputs {
		switch ot := o.Output.(type) {
		case deviceAgent.ServiceOutput:
			if err := ot.Start(); err != nil {
				log.Printf("E! Service for output %s failed to start, exiting\n%s\n",
					o.Name, err.Error())
				return err
			}
		}
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

	wg.Add(len(a.Config.Inputs))
	for _, input := range a.Config.Inputs {
		switch p := input.Input.(type) {
		case deviceAgent.ServiceInput:
			if err := p.Start(); err != nil {
				log.Printf("E! Service for input %s failed to start:\n%s\n", input.Name(), err.Error())
				break
			} else {
				defer p.Stop()
			}
			log.Printf("D! Successfully connected to input: %s\n", p.Name())

			switch pC := p.(type) {
			case deviceAgent.ControllerInput:
				for _, c := range a.Config.Controllers {
					c.Controller.RegisterInput(pC.Name(), pC)
				}
			}
		}

		inter := a.Config.Global.Interval.Duration
		if input.Config.Interval != 0 {
			inter = input.Config.Interval
		}
		go func(in *models.RunningInput, interval time.Duration) {
			defer wg.Done()
			a.gatherer(in, interval, metricC)
		}(input, inter)
	}

	//debug all stat
	wg.Add(1)
	go func() {
		defer wg.Done()
		//log.Println(A.Config.Global)
		//
		//for range time.Tick(time.Second * 1) {
		//	log.Println(models.GlobalMetricsGathered.Name(), models.GlobalMetricsGathered.FieldName(), models.GlobalMetricsGathered.Get())
		//	log.Println(NErrors.Name(), NErrors.FieldName(), NErrors.Get())
		//	log.Println(MetricFieldsCount.Name(), MetricFieldsCount.FieldName(), MetricFieldsCount.Get())
		//
		//	for _, i := range a.Config.Inputs {
		//		log.Println(i.MetricsGathered.Name(), i.MetricsGathered.FieldName(), i.MetricsGathered.Get())
		//	}
		//}
	}()

	LoadConfig()

	wg.Wait()
	a.Close()

	return nil
}
