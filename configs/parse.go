package configs

import "C"
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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"reflect"
	"time"
)

type Config struct {
	Global      *GlobalConfig               `json:"agent"`
	Inputs      []*models.RunningInput      `json:"inputs"`
	Outputs     []*models.RunningOutput     `json:"outputs"`
	Controllers []*models.RunningController `json:"controllers"`
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
	config := make(map[string]map[string]interface{})
	e := json.Unmarshal(content, &config)
	if e != nil {
		return e
	}

	for cK, cV := range config {
		switch cK {
		case "agent":
			json.Unmarshal([]byte(gjson.GetBytes(content, "agent").String()), c.Global)
		case "inputs", "outputs", "controllers":
			for _, v := range cV {
				vV, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("can't parse config content: %v", v)
				}
				if e := c.LoadPlugin(cK, vV["plugin_name"].(string), vV); e != nil {
					log.Printf("LoadPlugin: %s failed: %v", cK, e)
					return e
				}
			}
		}
	}

	return nil
}

func (c *Config) addInputJson(name string, table map[string]interface{}) error {
	creator, ok := inputs.Inputs[name]
	if !ok {
		return fmt.Errorf("undefined but requested input: %s", name)
	}
	input := creator()
	switch t := input.(type) {
	case parsers.ParserInput:
		t.SetParser(nil) //TODO
	}
	pluginConfig, err := buildInputJson(name, table)
	if err != nil {
		return err
	}

	pointMap := make(map[string]deviceAgent.PointDefine, 0)
	if pluginConfig.PointMapContent != "" {
		yaml.UnmarshalStrict([]byte(pluginConfig.PointMapContent), &pointMap)
	} else if pluginConfig.PointMapPath != "" {
		pMContent, err := ioutil.ReadFile(pluginConfig.PointMapPath)
		if err != nil {
			log.Printf("Can't load point_map file: %s, %s", pluginConfig.PointMapPath, err)
		} else {
			yaml.UnmarshalStrict(pMContent, &pointMap)
		}
	}
	input.SetPointMap(pointMap)

	if err := mapstructure.Decode(table, input); err != nil {
		return err
	}

	rp := models.NewRunningInput(input, pluginConfig)
	c.Inputs = append(c.Inputs, rp)
	return nil
}

func (c *Config) addOutputJson(name string, table map[string]interface{}) error {
	creator, ok := outputs.Outputs[name]
	if !ok {
		return fmt.Errorf("undefined but requested output: %s", name)
	}
	output := creator()
	switch t := output.(type) {
	case serializers.SerializerOutput:
		serializer, err := buildSerializerJson(name, table)
		if err != nil {
			return err
		}
		t.SetSerializer(serializer)
	}
	if err := mapstructure.Decode(table, output); err != nil {
		return err
	}

	ro := models.NewRunningOutput(name, output, c.Global.MetricBatchSize, c.Global.MetricBufferLimit)
	c.Outputs = append(c.Outputs, ro)

	return nil
}

func (c *Config) addControllerJson(name string, table map[string]interface{}) error {
	creator, ok := controllers.Controllers[name]
	if !ok {
		return fmt.Errorf("undefined but requested controller: %s", name)
	}
	controller := creator()
	if err := mapstructure.Decode(table, controller); err != nil {
		return err
	}

	rC := models.NewRunningController(name, controller)
	c.Controllers = append(c.Controllers, rC)
	return nil
}

func (c *Config) LoadPlugin(tp string, name string, table map[string]interface{}) error {
	switch tp {
	case "inputs":
		return c.addInputJson(name, table)
	case "outputs":
		return c.addOutputJson(name, table)
	case "controllers":
		return c.addControllerJson(name, table)
	default:
		return errors.New("unknown plugin: " + name)
	}
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

	if node, ok := table["point_map_path"]; ok {
		if nodeV, ok := node.(string); ok {
			cp.PointMapPath = nodeV
		}
	}
	if node, ok := table["point_map_content"]; ok {
		if nodeV, ok := node.(string); ok {
			cp.PointMapContent = nodeV
		}
	}

	return cp, nil
}

func buildSerializerJson(name string, tbl map[string]interface{}) (serializers.Serializer, error) {
	c := &serializers.Config{TimestampUnits: time.Duration(1 * time.Second), DataFormat: "json"}
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
