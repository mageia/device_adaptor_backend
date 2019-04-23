package kj66

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
)

type Parser struct {
}

type XDWsData struct {
	CmdId   int           `json:"CMD_ID"`
	Content []interface{} `json:"CONTENT"`
}

func (p *Parser) Parse(m []byte) (interface{}, error) {
	var contentData XDWsData
	if e := json.Unmarshal(m, &contentData); e != nil {
		log.Error().Err(e).Msg("Unmarshal Message")
	}

	log.Debug().Interface("CmdId", contentData.CmdId).Interface("content", contentData.Content).Msg("CmdId")

	//fields := make(map[string]interface{})
	//for _, content := range contentData.Content {
		//for k, v := range content.(map[string]interface{}) {
			//log.Debug().Str("key", k, ).Interface("value", v).Msg("content")
		//}
	//}
	return nil, nil
}
func (p *Parser) ParseCmd(cmd string, line []byte) (interface{}, error) { return nil, nil }
