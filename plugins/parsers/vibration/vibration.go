package vibration

import "deviceAdaptor"

type Parser struct {
}

func (p *Parser) Parse(line []byte) ([]deviceAgent.Metric, error) {
	return nil, nil
}
func (p *Parser) ParseLine(line string) (deviceAgent.Metric, error) {
	return nil, nil
}
