package deviceAgent

type Output interface {
	Connect() error
	Close() error
	Description() string
	SampleConfig() string
	Write(metrics []Metric) error
}

type ServiceOutput interface {
	Connect() error
	Close() error
	Description() string
	SampleConfig() string
	Write(metrics []Metric) error
	Start() error
	Stop()
}

//type AggregatingOutput interface {
//	Connect() error
//	Close() error
//	Description() string
//	SampleConfig() string
//	Write(metrics []Metric) error
//	Add(in Metric)
//	Push() []Metric
//	Reset()
//}
