package configs

import (
	"deviceAdaptor/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"runtime"
	"sort"
	"time"
)

type MemoryConfigType struct {
	User        map[string]map[string]interface{} `json:"-"`
	Agent       *GlobalConfig                     `json:"agent"`
	Inputs      []map[string]interface{}          `json:"inputs"`
	Outputs     []map[string]interface{}          `json:"outputs"`
	Controllers []map[string]interface{}          `json:"controllers"`
}
type ConfigSample struct {
	Key     string
	Label   string
	Default interface{}
	Type    string
	Order   int
}
type ConfigSampleArray []ConfigSample

var CurrentConfig []byte
var MemoryConfig MemoryConfigType
var defaultConfigJson = `
{
  "user":{
	"admin": {
      "password": "admin"
	}
  },
  "agent": {
    "debug": true,
    "interval": "3s",
    "flush_interval": "10s",
    "collection_jitter": "10ms",
    "flush_jitter": "10ms",
    "metric_batch_size": 0,
    "metric_buffer_limit": 0
  },
  "inputs": [
    {
      "id": "9f54b4b2-efb1-4f01-ad27-6b6da8d676a0",
      "created_at": 1541487113347,
      "interval": "3s",
      "name_override": "fake",
      "plugin_name": "fake",
    }
  ],
  "outputs": [
    {
      "id": "62b3328f-b1cf-4a9f-a456-e33e3a8c22c6",
      "created_at": 1541487113347,
      "plugin_name": "redis",
      "url_address": "redis://localhost:6379/0"
    }
  ],
  "controllers": [
    {
      "id": "d9019649-f2f7-431b-ac16-449ef7ca8ba1",
      "address": ":9999",
      "created_at": 1541487113347,
      "plugin_name": "http"
    }
  ]
}
`
var jsonConfigPath = "device_adaptor.json"

var InputSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at": ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", -100},
	},
	"opc": {
		"plugin_name":     ConfigSample{"plugin_name", "插件名称", "opc", "select", 1},
		"name_override":   ConfigSample{"name_override", "数据源名称", "opc", "input", 2},
		"address":         ConfigSample{"address", "数据源地址", "10.211.55.4:2048", "input", 3},
		"opc_server_name": ConfigSample{"opc_server_name", "OPC名称", "Kepware.KepServerEx.V5", "input", 4},
		"interval":        ConfigSample{"interval", "采集周期", "5s", "combine", 20},
		"field_prefix":    ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":    ConfigSample{"field_suffix", "测点前缀", "", "input", 22},
	},
	"modbus": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "modbus", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "modbus", "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "10.211.55.4:502", "input", 3},
		"slave_id":      ConfigSample{"slave_id", "从站地址", 1, "input", 4},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点前缀", "", "input", 22},
	},
	"fake": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "fake", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "fake", "input", 2},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点前缀", "", "input", 22},
	},
	"s7": {
		"plugin_name":   ConfigSample{"plugin_name", "插件名称", "s7", "select", 1},
		"name_override": ConfigSample{"name_override", "数据源名称", "s7", "input", 2},
		"address":       ConfigSample{"address", "数据源地址", "192.168.0.168", "input", 3},
		"rack":          ConfigSample{"rack", "机架号", 0, "input", 4},
		"slot":          ConfigSample{"slot", "槽号", 1, "input", 5},
		"interval":      ConfigSample{"interval", "采集周期", "3s", "combine", 20},
		"field_prefix":  ConfigSample{"field_prefix", "测点前缀", "", "input", 21},
		"field_suffix":  ConfigSample{"field_suffix", "测点前缀", "", "input", 22},
	},
	"http_listener": {
		"plugin_name":    ConfigSample{"plugin_name", "插件名称", "http_listener", "select", 1},
		"listen_address": ConfigSample{"listen_address", "监听地址", "0.0.0.0:19999", "input", 2},
		"max_body_size":  ConfigSample{"max_body_size", "最大消息体大小", 5 * 1024 * 1024, "input", 3},
		"max_line_size":  ConfigSample{"max_line_size", "最大文件行数", 64 * 1024, "input", 4},
		"read_timeout":   ConfigSample{"read_timeout", "读超时时间", "10s", "combine", 5},
		"write_timeout":  ConfigSample{"write_timeout", "写超时时间", "10s", "combine", 6},
		"basic_username": ConfigSample{"basic_username", "认证账户", "", "input", 7},
		"basic_password": ConfigSample{"basic_password", "认证密码", "", "input", 8},
	},
}
var OutputSample = map[string]map[string]ConfigSample{
	"_base": {
		"metric_buffer_limit": ConfigSample{"metric_buffer_limit", "批量上传缓冲区大小", 0, "input", 100},
		"metric_batch_size":   ConfigSample{"metric_batch_size", "测点批量上传数量", 0, "input", 101},
		"created_at":          ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", 0},
	},
	"redis": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "redis", "select", 1},
		"url_address": ConfigSample{"url_address", "地址URL", "redis://localhost:6379/0", "input", 2},
	},
	"file": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "file", "select", 1},
		"files":       ConfigSample{"files", "输出地址", []string{"stdout"}, "multi-input", 2},
	},
}
var ControllerSample = map[string]map[string]ConfigSample{
	"_base": {
		"created_at": ConfigSample{"created_at", "创建时间", time.Now().UnixNano() / 1e6, "none", 0},
	},
	"http": {
		"plugin_name": ConfigSample{"plugin_name", "插件名称", "http", "select", 1},
		"address":     ConfigSample{"address", "监听地址", "0.0.0.0:9999", "input", 2},
	},
}

func (c ConfigSampleArray) Len() int {
	return len(c)
}
func (c ConfigSampleArray) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c ConfigSampleArray) Less(i, j int) bool {
	return c[i].Order < c[j].Order
}

func GetInputConfigById(id string) (map[string]interface{}, bool) {
	for _, v := range MemoryConfig.Inputs {
		if v["id"] == id && v["name_override"] != "" {
			return v, true
		}
	}
	return nil, false
}

func GenConfigSample(pluginType, pluginName string, exclude ...string) (map[string]interface{}, error) {
	r := make(map[string]interface{})

	switch pluginType {
	case "inputs":
		if _, ok := InputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range InputSample["_base"] {
			r[k] = v.Default
		}
		for k, v := range InputSample[pluginName] {
			r[k] = v.Default
		}
	case "outputs":
		if _, ok := OutputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range OutputSample["_base"] {
			r[k] = v.Default
		}
		for k, v := range OutputSample[pluginName] {
			r[k] = v.Default
		}
	case "controllers":
		if _, ok := ControllerSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range ControllerSample["_base"] {
			r[k] = v.Default
		}
		for k, v := range ControllerSample[pluginName] {
			r[k] = v.Default
		}
	}
	r["id"] = uuid.New().String()
	for _, k := range exclude {
		delete(r, k)
	}
	return r, nil
}

func GenConfigSampleArray(pluginType, pluginName string) (ConfigSampleArray, error) {
	targetArray := make(ConfigSampleArray, 0)

	switch pluginType {
	case "inputs":
		if _, ok := InputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}

		for _, v := range InputSample["_base"] {
			targetArray = append(targetArray, v)
		}

		for _, v := range InputSample[pluginName] {
			targetArray = append(targetArray, v)
		}
	case "outputs":
		if _, ok := OutputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}

		for _, v := range OutputSample["_base"] {
			targetArray = append(targetArray, v)
		}

		for _, v := range OutputSample[pluginName] {
			targetArray = append(targetArray, v)
		}

	case "controllers":
		if _, ok := ControllerSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}

		for _, v := range ControllerSample["_base"] {
			targetArray = append(targetArray, v)
		}

		for _, v := range ControllerSample[pluginName] {
			targetArray = append(targetArray, v)
		}
	}

	sort.Sort(targetArray)
	return targetArray, nil
}

func FlushMemoryConfig() {
	CurrentConfig, _ = json.Marshal(MemoryConfig)
	ioutil.WriteFile(jsonConfigPath, CurrentConfig, 0644)
}

func GetConfigContent() []byte {
	CurrentConfig = []byte(defaultConfigJson)
	if utils.IsExists(jsonConfigPath) {
		CurrentConfig, _ = ioutil.ReadFile(jsonConfigPath)
	}

	result := gjson.GetManyBytes(CurrentConfig, "agent", "inputs", "outputs", "controllers")
	json.Unmarshal([]byte(result[0].Raw), &MemoryConfig.Agent)
	json.Unmarshal([]byte(result[1].Raw), &MemoryConfig.Inputs)
	json.Unmarshal([]byte(result[2].Raw), &MemoryConfig.Outputs)
	json.Unmarshal([]byte(result[3].Raw), &MemoryConfig.Controllers)

	ioutil.WriteFile(jsonConfigPath, CurrentConfig, 0644)

	return CurrentConfig
}

func init() {
	if runtime.GOOS == "linux" {
		jsonConfigPath = "/var/run/deviceAdaptor/device_adaptor.json"
	}
}
