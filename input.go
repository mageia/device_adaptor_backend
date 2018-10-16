package deviceAgent

type PointDefine struct {
	Label     string            `json:"label"`
	Name      string            `json:"name"`
	Desc      string            `json:"desc"`
	Unit      string            `json:"unit"`
	IsAnalog  bool              `json:"is_analog"`
	Parameter float64           `json:"parameter,omitempty"`
	Option    map[string]string `json:"option,omitempty"`
	Control   map[string]string `json:"control,omitempty"`
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
