package deviceAgent

type PointDefine struct {
	Label     string
	Name      string
	Unit      string
	IsAnalog  bool
	Parameter float64
	Option    map[string]string
	Control   map[string]string
}

type Input interface {
	Name() string
	Gather(Accumulator) error
	SetPointMap(map[string]PointDefine)
	FlushPointMap(Accumulator) error
}

type ServiceInput interface {
	Name() string
	Gather(Accumulator) error
	SetPointMap(map[string]PointDefine)
	FlushPointMap(Accumulator) error

	Start(Accumulator) error
	Stop() error
}
