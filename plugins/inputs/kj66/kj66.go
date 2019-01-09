package kj66

import (
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/inputs"
	"github.com/howeyc/crc16"
	"github.com/rs/zerolog/log"
)

type KJ66 struct {
	Address string `json:"address"`
	Version string `json:"version"`

	connected    bool
	quality      device_agent.Quality
	originName   string
	FieldPrefix  string `json:"field_prefix"`
	FieldSuffix  string `json:"field_suffix"`
	NameOverride string `json:"name_override"`
}

func (k *KJ66) SelfCheck() device_agent.Quality {
	return device_agent.QualityGood
}

func (k *KJ66) Name() string {
	if k.NameOverride != "" {
		return k.NameOverride
	}
	return k.originName
}

func (k *KJ66) Gather(device_agent.Accumulator) error {
	//24 FF 00 00 00 07 08 06 06 06 06 06 06 80 80 80 80 80 80 80 80 00 00 0F 00 01 0A 63 63 63 63 63 63 00 14
	//'0x24, 0xFF, 0x00, 0x00, 0x00, 0x07, 0x08, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00, 0x0F, 0x00, 0x01, 0x0A, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x00, 0x14'
	log.Debug().Uint16("crc", crc16.Checksum([]byte{0x24, 0xFF, 0x00, 0x00, 0x00, 0x07, 0x08, 0x06, 0x06, 0x06, 0x06, 0x06, 0x06, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80,
	0x80, 0x80, 0x00, 0x00, 0x0F, 0x00, 0x01, 0x0A, 0x63, 0x63, 0x63, 0x63, 0x63, 0x63, 0x00}, crc16.CCITTFalseTable)).Msg("CRC")
	return nil
}

func (k *KJ66) SetPointMap(map[string]points.PointDefine) {

}

func (k *KJ66) Start() error {
	return nil
}

func (k *KJ66) Stop() {

}

func init() {
	inputs.Add("kj66", func() device_agent.Input {
		return &KJ66{}
	})
}
