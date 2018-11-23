package deviceAgent

type Processor interface {
	Apply(in ...Metric) []Metric
}
