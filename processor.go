package device_agent

type Processor interface {
	Apply(in ...Metric) []Metric
}
