package csv

import (
	"bytes"
	"device_adaptor"
	"device_adaptor/metric"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	MetricName        string   `json:"metric_name"`
	HeaderRowCount    int      `json:"header_row_count"`
	SkipRows          int      `json:"skip_rows"`
	SkipColumns       int      `json:"skip_columns"`
	Delimiter         string   `json:"delimiter"`
	Comment           string   `json:"comment"`
	TrimSpace         bool     `json:"trim_space"`
	ColumnNames       []string `json:"column_names"`
	TagColumns        []string `json:"tag_columns"`
	MeasurementColumn string   `json:"measurement_column"`
	TimestampColumn   string   `json:"timestamp_column"`
	TimestampFormat   string   `json:"timestamp_format"`
	//DefaultTags       map[string]string
}

func (p *Parser) ParseCmd(string, []byte) (interface{}, error) {
	panic("implement me")
}

func (p *Parser) Parse([]byte) (interface{}, error) {
	return nil, nil
}

func (p *Parser) initReader(r *bytes.Reader) (*csv.Reader, error) {
	csvReader := csv.NewReader(r)
	csvReader.FieldsPerRecord = -1
	if p.Delimiter != "" {
		csvReader.Comma = []rune(p.Delimiter)[0]
	}
	if p.Comment != "" {
		csvReader.Comment = []rune(p.Comment)[0]
	}
	return csvReader, nil
}

func (p *Parser) parseRecord(record []string) (device_adaptor.Metric, error) {
	recordFields := make(map[string]interface{})
	tags := make(map[string]string)

	record = record[p.SkipColumns:]

outer:
	for i, fieldName := range p.ColumnNames {
		if i < len(record) {
			value := record[i]
			if p.TrimSpace {
				value = strings.Trim(value, " ")
			}
			for _, tagName := range p.TagColumns {
				if tagName == fieldName {
					tags[tagName] = value
					continue outer
				}
			}

			if iValue, err := strconv.ParseInt(value, 10, 64); err == nil {
				recordFields[fieldName] = iValue
			} else if fValue, err := strconv.ParseFloat(value, 64); err == nil {
				recordFields[fieldName] = fValue
			} else if bValue, err := strconv.ParseBool(value); err == nil {
				recordFields[fieldName] = bValue
			} else {
				recordFields[fieldName] = value
			}
		}
	}

	//for k, v := range p.DefaultTags {
	//	tags[k] = v
	//}

	measurementName := p.MetricName
	if recordFields[p.MeasurementColumn] != nil {
		measurementName = fmt.Sprintf("%v", recordFields[p.MeasurementColumn])
	}

	metricTime := time.Now()
	if p.TimestampColumn != "" {
		if recordFields[p.TimestampColumn] == nil {
			return nil, fmt.Errorf("timestamp column: %v could not be found", p.TimestampColumn)
		}
		tStr := fmt.Sprintf("%v", recordFields[p.TimestampColumn])
		if p.TimestampFormat == "" {
			return nil, fmt.Errorf("timestamp format must by specified")
		}
		var err error
		metricTime, err = time.Parse(p.TimestampFormat, tStr)
		if err != nil {
			return nil, err
		}
	}

	m, err := metric.New(measurementName, tags, recordFields, device_adaptor.QualityGood, metricTime, device_adaptor.Untyped)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (p *Parser) Parse2(line []byte) ([]device_adaptor.Metric, error) {
	r := bytes.NewReader([]byte(line))
	csvReader, err := p.initReader(r)
	if err != nil {
		return nil, err
	}
	for i := 0; i < p.SkipRows; i++ {
		csvReader.Read()
	}
	headerNames := make([]string, 0)
	if len(p.ColumnNames) == 0 {
		for i := 0; i < p.HeaderRowCount; i++ {
			header, err := csvReader.Read()
			if err != nil {
				return nil, err
			}
			for i := range header {
				name := header[i]
				if p.TrimSpace {
					name = strings.Trim(name, " ")
				}
				if len(headerNames) <= i {
					headerNames = append(headerNames, name)
				} else {
					headerNames[i] = headerNames[i] + name
				}
			}
		}
		p.ColumnNames = headerNames[p.SkipColumns:]
	} else {
		for i := 0; i < p.HeaderRowCount; i++ {
			csvReader.Read()
		}
	}

	table, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}
	metrics := make([]device_adaptor.Metric, 0)
	for _, record := range table {
		m, err := p.parseRecord(record)
		if err != nil {
			return metrics, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (p *Parser) ParseLine(line string) (device_adaptor.Metric, error) {
	r := bytes.NewReader([]byte(line))
	csvReader, err := p.initReader(r)
	if err != nil {
		return nil, err
	}
	if len(p.ColumnNames) == 0 {
		return nil, fmt.Errorf("[parsers.csv] data columns must be specified")
	}
	record, err := csvReader.Read()
	if err != nil {
		return nil, err
	}
	m, err := p.parseRecord(record)
	if err != nil {
		return nil, err

	}
	return m, nil
}

//func (p *Parser) SetDefaultTags(tags map[string]string) {
//	p.DefaultTags = tags
//}
