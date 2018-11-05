package configs

import (
	"deviceAdaptor"
	"deviceAdaptor/internal"
	"deviceAdaptor/internal/models"
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	"deviceAdaptor/plugins/parsers"
	"deviceAdaptor/plugins/serializers"
	"deviceAdaptor/utils"
	"fmt"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"time"
)

type Config struct {
	Tags        map[string]string
	Global      *GlobalConfig
	Inputs      []*models.RunningInput
	Outputs     []*models.RunningOutput
	Controllers []*models.RunningController
}

func NewConfig() *Config {
	c := &Config{
		Global: &GlobalConfig{
			Interval:      internal.Duration{Duration: 10 * time.Second},
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
		},
		Controllers: make([]*models.RunningController, 0),
		Tags:        make(map[string]string),
		Inputs:      make([]*models.RunningInput, 0),
		Outputs:     make([]*models.RunningOutput, 0),
	}
	return c
}

type GlobalConfig struct {
	Debug             bool
	Interval          internal.Duration
	FlushInterval     internal.Duration
	CollectionJitter  internal.Duration
	FlushJitter       internal.Duration
	MetricBatchSize   int
	MetricBufferLimit int
}

func getDefaultConfigPath() (string, error) {

	return "../configs/device_adaptor.toml", nil
}

func parseFile(p string) (*ast.Table, error) {
	contents, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return toml.Parse(contents)
}

func (c *Config) LoadConfigJson(content []byte) error {
	var config = make(map[string]interface{})
	if e := json.Unmarshal(content, &config); e != nil {
		return e
	}

	return nil
}

func (c *Config) LoadConfig(path string) error {
	var err error
	if path == "" {
		if path, err = getDefaultConfigPath(); err != nil {
			return err
		}
	}
	tbl, err := parseFile(path)
	if err != nil {
		return fmt.Errorf("error parsing %s, %s", path, err)
	}

	for _, tableName := range []string{"tags", "global_tags"} {
		if val, ok := tbl.Fields[tableName]; ok {
			subTable, ok := val.(*ast.Table)
			if !ok {
				return fmt.Errorf("%s: invalid configuration", path)
			}
			if err = toml.UnmarshalTable(subTable, c.Tags); err != nil {
				log.Printf("E! Could not parse [global_tags] config\n")
				return fmt.Errorf("error parsing %s, %s", path, err)
			}
		}
	}

	if val, ok := tbl.Fields["agent"]; ok {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid configuration", path)
		}
		if err = toml.UnmarshalTable(subTable, c.Global); err != nil {
			log.Printf("E! Could not parse [agent] config\n")
			return fmt.Errorf("error parsing %s, %s", path, err)
		}
	}

	for name, val := range tbl.Fields {
		subTable, ok := val.(*ast.Table)
		if !ok {
			return fmt.Errorf("%s: invalid configuration", path)
		}
		switch name {
		case "agent", "global_tags", "tags":
		case "controllers":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					if err = c.addController(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("error parsing %s, %s", path, err)
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addController(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %s", path, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s, file: %s", pluginName, path)
				}
			}

		case "outputs":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					if err = c.addOutput(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("error parsing %s, %s", path, err)
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addOutput(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %s", path, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s, file: %s", pluginName, path)
				}
			}
		case "inputs":
			for pluginName, pluginVal := range subTable.Fields {
				switch pluginSubTable := pluginVal.(type) {
				case *ast.Table:
					if err = c.addInput(pluginName, pluginSubTable); err != nil {
						return fmt.Errorf("error parsing %s, %s", path, err)
					}
				case []*ast.Table:
					for _, t := range pluginSubTable {
						if err = c.addInput(pluginName, t); err != nil {
							return fmt.Errorf("error parsing %s, %s", path, err)
						}
					}
				default:
					return fmt.Errorf("unsupported config format: %s, file: %s", pluginName, path)
				}
			}
		default:
			if err = c.addInput(name, subTable); err != nil {
				return fmt.Errorf("error parsing %s, %s", path, err)
			}
		}
	}

	return nil
}

func (c *Config) addInput(name string, table *ast.Table) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("undefined but requested input: %s", name)
	}
	input := creator()

	switch t := input.(type) {
	case parsers.ParserInput:
		pM, err := buildParserMap(name, table)
		if err != nil {
			return err
		}
		t.SetParser(pM)
	}

	pluginConfig, err := buildInput(name, table)
	if err != nil {
		return err
	}

	pointMap := make(map[string]deviceAgent.PointDefine, 0)
	if pluginConfig.PointMapPath != "" {
		if pMContent, err := ioutil.ReadFile(pluginConfig.PointMapPath); err != nil {
			log.Printf("Can't load point_map file: %s, %s", pluginConfig.PointMapPath, err)
		} else {
			yaml.UnmarshalStrict(pMContent, &pointMap)
		}
	}
	input.SetPointMap(pointMap)

	if err := toml.UnmarshalTable(table, input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) addOutput(name string, table *ast.Table) error {
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("undefined but requested output: %s", name)
	}
	output := creator()
	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializer(name, table)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}
	if err := toml.UnmarshalTable(table, output); err != nil {
		return err
	}
	ro := models.NewRunningOutput(name, output, c.Global.MetricBatchSize, c.Global.MetricBufferLimit)
	c.Outputs = append(c.Outputs, ro)

	return nil
}

func (c *Config) addController(name string, table *ast.Table) error {
	creator, ok := controllers.Controllers[name]
	if !ok {
		return fmt.Errorf("undefined but requested controller: %s", name)
	}
	controller := creator()
	if err := toml.UnmarshalTable(table, controller); err != nil {
		return err
	}
	rC := models.NewRunningController(name, controller)
	c.Controllers = append(c.Controllers, rC)

	return nil
}

func buildInput(name string, table *ast.Table) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}

	if node, ok := table.Fields["interval"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				dur, err := time.ParseDuration(str.Value)
				if err != nil {
					return nil, err
				}
				cp.Interval = dur
			}
		}
	}

	if node, ok := table.Fields["point_map"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				cp.PointMapPath = str.Value
			}
		}
	}
	delete(table.Fields, "point_map")
	delete(table.Fields, "interval")
	return cp, nil
}

func buildSerializer(name string, tbl *ast.Table) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second)}

	if node, ok := tbl.Fields["data_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DataFormat = str.Value
			}
		}
	}
	if c.DataFormat == "" {
		c.DataFormat = "json"
	}

	delete(tbl.Fields, "data_format")
	delete(tbl.Fields, "prefix")
	return serializers.NewSerializer(c)
}

func buildParserMap(name string, tbl *ast.Table) (map[string]parsers.Parser, error) {
	r := make(map[string]parsers.Parser)

	if val, ok := tbl.Fields["parser"]; ok {
		blobMap := make(map[string]interface{})
		switch pV := val.(type) {
		case *ast.Table:
			for k, v := range pV.Fields {
				switch vT := v.(type) {
				case []*ast.Table:
					for _, vP := range vT {
						if e := toml.UnmarshalTable(vP, blobMap); e != nil {
							return nil, e
						}
					}
				case *ast.Table:
					if e := toml.UnmarshalTable(vT, blobMap); e != nil {
						return nil, e
					}
				default:
					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
				}

				if _, ok := blobMap["parser_name"]; !ok {
					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
				}

				if _, ok := blobMap["parser_name"]; !ok {
					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
				}
				if _, ok := blobMap["parser_name"].(string); !ok {
					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
				}

				param := []reflect.Value{reflect.ValueOf(&parsers.ParserBlob{}), reflect.ValueOf(blobMap)}
				funcName := "BuildParser" + utils.UcFirst(blobMap["parser_name"].(string))
				if m, ok := reflect.TypeOf(&parsers.ParserBlob{}).MethodByName(funcName); !ok {
					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
				} else {
					result := m.Func.Call(param)
					if result[1].Interface() != nil {
						return nil, fmt.Errorf("[%s] %s: invalid parser configuration", utils.GetLineNo(), name)
					}
					r[k] = result[0].Interface().(parsers.Parser)
				}
			}
		default:
			return nil, fmt.Errorf("[%s] %s: invalid parser configuration", utils.GetLineNo(), name)
		}
	}

	delete(tbl.Fields, "parser")

	return r, nil
}

func buildParser1(name string, tbl *ast.Table) (parsers.Parser, error) {
	c := &parsers.Config{}
	if node, ok := tbl.Fields["data_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.DataFormat = str.Value
			}
		}
	}
	//for csv parser
	if node, ok := tbl.Fields["csv_column_names"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CSVColumnNames = append(c.CSVColumnNames, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_tag_columns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if ary, ok := kv.Value.(*ast.Array); ok {
				for _, elem := range ary.Value {
					if str, ok := elem.(*ast.String); ok {
						c.CSVTagColumns = append(c.CSVTagColumns, str.Value)
					}
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_delimiter"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVDelimiter = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_comment"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVComment = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_measurement_column"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVMeasurementColumn = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_timestamp_column"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVTimestampColumn = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_timestamp_format"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				c.CSVTimestampFormat = str.Value
			}
		}
	}

	if node, ok := tbl.Fields["csv_header_row_count"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				iVal, err := strconv.Atoi(str.Value)
				c.CSVHeaderRowCount = iVal
				if err != nil {
					return nil, fmt.Errorf("parsing to int: %v", err)
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_skip_rows"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				iVal, err := strconv.Atoi(str.Value)
				c.CSVSkipRows = iVal
				if err != nil {
					return nil, fmt.Errorf("error parsing to int: %v", err)
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_skip_columns"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.String); ok {
				iVal, err := strconv.Atoi(str.Value)
				c.CSVSkipColumns = iVal
				if err != nil {
					return nil, fmt.Errorf("parsing to int: %v", err)
				}
			}
		}
	}

	if node, ok := tbl.Fields["csv_trim_space"]; ok {
		if kv, ok := node.(*ast.KeyValue); ok {
			if str, ok := kv.Value.(*ast.Boolean); ok {
				//for config with no quotes
				val, err := strconv.ParseBool(str.Value)
				c.CSVTrimSpace = val
				if err != nil {
					return nil, fmt.Errorf("parsing to bool: %v", err)
				}
			}
		}
	}

	c.MetricName = name
	delete(tbl.Fields, "data_format")
	delete(tbl.Fields, "csv_data_columns")
	delete(tbl.Fields, "csv_tag_columns")
	delete(tbl.Fields, "csv_field_columns")
	delete(tbl.Fields, "csv_name_column")
	delete(tbl.Fields, "csv_timestamp_column")
	delete(tbl.Fields, "csv_timestamp_format")
	delete(tbl.Fields, "csv_delimiter")
	delete(tbl.Fields, "csv_header")
	return parsers.NewParser(c)
}
