package device_adaptor

type Processor interface {
	Apply(in ...Metric) []Metric
}
