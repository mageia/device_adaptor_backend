package deviceAgent

type PointDefine struct {
	Label     string            `json:"label"`
	Name      string            `json:"name"`
	Unit      string            `json:"unit"`
	IsAnalog  bool              `json:"is_analog"`
	Parameter float64           `json:"parameter"`
	Option    map[string]string `json:"option"`
	Control   map[string]string `json:"control"`
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

	Start() error
	Stop() error
}
