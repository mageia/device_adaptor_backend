package interfaces

type Output interface {
	Connect() error
	Close() error
	Description() error
	SampleConfig() string
	Write(metrics []Metric) error
}
