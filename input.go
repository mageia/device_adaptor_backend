package deviceAgent

type PointDefine struct {
	Name      string                 `json:"name"`
	Unit      string                 `json:"unit"`
	Parameter float64                `json:"parameter,omitempty"`
	Option    map[string]string      `json:"option,omitempty"`
	Control   map[string]string      `json:"control,omitempty"`
	Tags      []string               `json:"tags,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

type Quality int

const (
	_ Quality = iota
	QualityGood
	QualityBad
	QualityUnknown
)

type Input interface {
	Name() string
	Gather(Accumulator) error
	SelfCheck() Quality
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
