package parsers

import (
	"device_adaptor/plugins/parsers/csv"
	"device_adaptor/plugins/parsers/kj66"
	"device_adaptor/plugins/parsers/vibration"
	"fmt"
)

type ParserInput interface {
	SetParser(parsers map[string]Parser)
}

type Parser interface {
	Parse([]byte) (interface{}, error)
	ParseCmd(string, []byte) (interface{}, error)
}

type ParserBlob struct{}

type Config struct {
	DataFormat string
	MetricName string

	//csv
	CSVHeaderRowCount    int
	CSVSkipRows          int
	CSVSkipColumns       int
	CSVDelimiter         string
	CSVComment           string
	CSVTrimSpace         bool
	CSVColumnNames       []string
	CSVTagColumns        []string
	CSVMeasurementColumn string
	CSVTimestampColumn   string
	CSVTimestampFormat   string
	DefaultTags          map[string]string
}

func NewParser(config *Config) (Parser, error) {
	var err error
	var parser Parser
	switch config.DataFormat {
	case "csv":
		parser, err = newCSVParser(config.MetricName,
			config.CSVHeaderRowCount,
			config.CSVSkipRows,
			config.CSVSkipColumns,
			config.CSVDelimiter,
			config.CSVComment,
			config.CSVTrimSpace,
			config.CSVColumnNames,
			config.CSVTagColumns,
			config.CSVMeasurementColumn,
			config.CSVTimestampColumn,
			config.CSVTimestampFormat,
			config.DefaultTags)
	default:
		err = fmt.Errorf("unsupported data format: %s", config.DataFormat)
	}
	return parser, err
}

func newCSVParser(metricName string,
	headerRowCount int,
	skipRows int,
	skipColumns int,
	delimiter string,
	comment string,
	trimSpace bool,
	dataColumns []string,
	tagColumns []string,
	nameColumn string,
	timestampColumn string,
	timestampFormat string,
	defaultTags map[string]string) (Parser, error) {
	if headerRowCount == 0 && len(dataColumns) == 0 {
		return nil, fmt.Errorf("there must be a header if `csv_data_columns` is not specified")
	}

	if delimiter != "" {
		runeStr := []rune(delimiter)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("delimiter must be a single character, got: %s", delimiter)
		}
		delimiter = fmt.Sprintf("%v", runeStr[0])
	}
	if comment != "" {
		runeStr := []rune(comment)
		if len(runeStr) > 1 {
			return nil, fmt.Errorf("comment must be a single character, got: %s", comment)
		}
		comment = fmt.Sprintf("%v", runeStr[0])
	}
	parser := &csv.Parser{
		MetricName:        metricName,
		HeaderRowCount:    headerRowCount,
		SkipRows:          skipRows,
		SkipColumns:       skipColumns,
		Delimiter:         delimiter,
		Comment:           comment,
		TrimSpace:         trimSpace,
		ColumnNames:       dataColumns,
		TagColumns:        tagColumns,
		MeasurementColumn: nameColumn,
		TimestampColumn:   timestampColumn,
		TimestampFormat:   timestampFormat,
		//DefaultTags:       defaultTags,
	}
	return parser, nil
}

func (*ParserBlob) BuildParserCsv(tbl map[string]interface{}) (Parser, error) {
	return &csv.Parser{}, nil
}

func (*ParserBlob) BuildParserVibration(tbl map[string]interface{}) (Parser, error) {
	return &vibration.Parser{}, nil
}

func (*ParserBlob) BuildParserKj66(tbl map[string]interface{}) (Parser, error) {
	return &kj66.Parser{}, nil
}

