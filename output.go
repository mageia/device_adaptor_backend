package device_adaptor

type Output interface {
	Connect() error
	Close() error
	Write(metrics []Metric) error
}

// 支持输出点表的 output
type RichOutput interface {
	Output
	// output 启动时被调用一次，而后点表变更时被调用
	WritePointMap(pointMap PointMap) error
}

type ServiceOutput interface {
	Output
	Start() error
	Stop()
}

//type AggregatingOutput interface {
//	Connect() error
//	Close() error
//	Write(metrics []Metric) error
//	Add(in Metric)
//	Push() []Metric
//	Reset()
//}
