package configs

import (
	"device_adaptor/internal"
	"device_adaptor/internal/models"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/controllers"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/outputs"
	"device_adaptor/plugins/parsers"
	"device_adaptor/plugins/serializers"
	"device_adaptor/utils"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"reflect"
	"time"
)

type Config struct {
	Global      *GlobalConfig               `json:"agent"`
	Inputs      []*models.RunningInput      `json:"inputs"`
	Outputs     []*models.RunningOutput     `json:"outputs"`
	Controllers []*models.RunningController `json:"controllers"`
	Processors  []*models.RunningProcessor  `json:"processors"`
}

type AgentConfig struct {
	Agent  GlobalConfig `json:"agent"`
	Inputs map[string][]map[string]interface{} `json:"inputs"`
	Outputs map[string][]map[string]interface{} `json:"outputs"`
	Controllers map[string][]map[string]interface{} `json:"controllers"`
}

func NewConfig() *Config {
	c := &Config{
		Global: &GlobalConfig{
			Interval:      internal.Duration{Duration: 10 * time.Second},
			FlushInterval: internal.Duration{Duration: 10 * time.Second},
		},
		Controllers: make([]*models.RunningController, 0),
		Inputs:      make([]*models.RunningInput, 0),
		Outputs:     make([]*models.RunningOutput, 0),
	}
	return c
}

type GlobalConfig struct {
	Debug             bool              `json:"debug"`
	Interval          internal.Duration `json:"interval"`
	FlushInterval     internal.Duration `json:"flush_interval"`
	CollectionJitter  internal.Duration `json:"collection_jitter"`
	FlushJitter       internal.Duration `json:"flush_jitter"`
	MetricBatchSize   int               `json:"metric_batch_size"`
	MetricBufferLimit int               `json:"metric_buffer_limit"`
}

func (c *Config) LoadConfigJson(content []byte) error {
	jsoniter.Unmarshal([]byte(gjson.GetBytes(content, "agent").String()), c.Global)

	for _, inputConfig := range gjson.GetBytes(content, "inputs").Array() {
		c.addInputBytes([]byte(inputConfig.Raw))
	}
	for _, outputConfig := range gjson.GetBytes(content, "outputs").Array() {
		c.addOutputBytes([]byte(outputConfig.Raw))
	}
	for _, controllerConfig := range gjson.GetBytes(content, "controllers").Array() {
		c.addControllersBytes([]byte(controllerConfig.Raw))
	}

	return nil
}

func (c *Config) addInputBytes(table []byte) error {
	pluginConf := make(map[string]interface{})
	jsoniter.Unmarshal(table, &pluginConf)

	pluginName := gjson.GetBytes(table, "plugin_name").String()
	creator, ok := inputs.Inputs[pluginName]
	if !ok {
		return fmt.Errorf("undefined but requested input: %s", pluginName)
	}
	input := creator()
	switch t := input.(type) {
	case parsers.ParserInput:
		parserMap, err := buildParserMapJson(pluginName, pluginConf)
		if err != nil {
			return err
		}
		t.SetParser(parserMap)
	}

	inputConfig, err := buildInputJson(pluginName, pluginConf)
	if err != nil {
		return err
	}

	//set point map to per input plugin
	nameOverride := gjson.GetBytes(table, "name_override").String()
	if nameOverride == "" {
		nameOverride = pluginName
	}
	pointMap := make(map[string]points.PointDefine)
	pointArray := make([]points.PointDefine, 0)
	points.SqliteDB.Where("input_name = ?", nameOverride).Find(&pointArray)
	for _, v := range pointArray {
		pointMap[v.PointKey] = v
	}
	input.SetPointMap(pointMap)

	if err := jsoniter.Unmarshal(table, &input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, inputConfig)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) addOutputBytes(table []byte) error {
	pluginConf := make(map[string]interface{})
	jsoniter.Unmarshal(table, &pluginConf)
	pluginName := gjson.GetBytes(table, "plugin_name").String()

	creator, ok := outputs.Outputs[pluginName]
	if !ok {
		return fmt.Errorf("undefined but requested output: %s", pluginName)
	}
	output := creator()
	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializerJson(pluginName, pluginConf)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}

	jsoniter.Unmarshal(table, &output)
	ro := models.NewRunningOutput(pluginName, output, c.Global.MetricBatchSize, c.Global.MetricBufferLimit)
	c.Outputs = append(c.Outputs, ro)

	return nil
}

func (c *Config) addControllersBytes(table []byte) error {
	pluginConf := make(map[string]interface{})
	jsoniter.Unmarshal(table, &pluginConf)
	pluginName := gjson.GetBytes(table, "plugin_name").String()

	creator, ok := controllers.Controllers[pluginName]
	if !ok {
		return fmt.Errorf("undefined but requested controller: %s", pluginName)
	}
	controller := creator()
	jsoniter.Unmarshal(table, &controller)

	rC := models.NewRunningController(pluginName, controller)
	c.Controllers = append(c.Controllers, rC)

	return nil
}

func buildInputJson(name string, table map[string]interface{}) (*models.InputConfig, error) {
	cp := &models.InputConfig{Name: name}
	if node, ok := table["interval"]; ok {
		if nodeV, ok := node.(string); ok {
			dur, err := time.ParseDuration(nodeV)
			if err != nil {
				return nil, err
			}
			cp.Interval = dur
		}
	}
	return cp, nil
}

func buildSerializerJson(name string, tbl map[string]interface{}) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second), DataFormat: "json"}
	return serializers.NewSerializer(c)
}

func buildParserByName(name string, table map[string]interface{}) (parsers.Parser, error) {
	param := []reflect.Value{reflect.ValueOf(&parsers.ParserBlob{}), reflect.ValueOf(table)}
	funcName := "BuildParser" + utils.UcFirst(name)

	if m, ok := reflect.TypeOf(&parsers.ParserBlob{}).MethodByName(funcName); !ok {
		return nil, fmt.Errorf("[%s]: invalid parser configuration: %s", utils.GetLineNo(), name)
	} else {
		result := m.Func.Call(param)
		if result[1].Interface() != nil {
			return nil, fmt.Errorf("[%s] %s: invalid parser configuration", utils.GetLineNo(), name)
		}
		return result[0].Interface().(parsers.Parser), nil
	}

	return nil, nil
}
func buildParserMapJson(name string, table map[string]interface{}) (map[string]parsers.Parser, error) {
	r := make(map[string]parsers.Parser)
	if val, ok := table["parser"]; ok {
		if pV, ok := val.(map[string]interface{}); ok {
			for k, v := range pV {
				if vV, ok := v.(map[string]interface{}); ok {
					p, e := buildParserByName(k, vV)
					if e == nil {
						r[k] = p
					}
				}
			}
		}
	}
	return r, nil
}

//func buildParserMap(name string, tbl *ast.Table) (map[string]parsers.Parser, error) {
//	r := make(map[string]parsers.Parser)
//
//	if val, ok := tbl.Fields["parser"]; ok {
//		blobMap := make(map[string]interface{})
//		switch pV := val.(type) {
//		case *ast.Table:
//			for k, v := range pV.Fields {
//				switch vT := v.(type) {
//				case []*ast.Table:
//					for _, vP := range vT {
//						if e := toml.UnmarshalTable(vP, blobMap); e != nil {
//							return nil, e
//						}
//					}
//				case *ast.Table:
//					if e := toml.UnmarshalTable(vT, blobMap); e != nil {
//						return nil, e
//					}
//				default:
//					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
//				}
//
//				if _, ok := blobMap["parser_name"]; !ok {
//					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
//				}
//
//				if _, ok := blobMap["parser_name"]; !ok {
//					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
//				}
//				if _, ok := blobMap["parser_name"].(string); !ok {
//					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
//				}
//
//				param := []reflect.Value{reflect.ValueOf(&parsers.ParserBlob{}), reflect.ValueOf(blobMap)}
//				funcName := "BuildParser" + utils.UcFirst(blobMap["parser_name"].(string))
//				if m, ok := reflect.TypeOf(&parsers.ParserBlob{}).MethodByName(funcName); !ok {
//					return nil, fmt.Errorf("[%s] %s: invalid parser configuration: %s", utils.GetLineNo(), name, k)
//				} else {
//					result := m.Func.Call(param)
//					if result[1].Interface() != nil {
//						return nil, fmt.Errorf("[%s] %s: invalid parser configuration", utils.GetLineNo(), name)
//					}
//					r[k] = result[0].Interface().(parsers.Parser)
//				}
//			}
//		default:
//			return nil, fmt.Errorf("[%s] %s: invalid parser configuration", utils.GetLineNo(), name)
//		}
//	}
//
//	delete(tbl.Fields, "parser")
//
//	return r, nil
//}
