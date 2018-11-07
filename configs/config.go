package configs

import (
	"deviceAdaptor/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"time"
)

type MemoryConfigType struct {
	Agent       *GlobalConfig                     `json:"agent"`
	Inputs      map[string]map[string]interface{} `json:"inputs"`
	Outputs     map[string]map[string]interface{} `json:"outputs"`
	Controllers map[string]map[string]interface{} `json:"controllers"`
}

var MemoryConfig MemoryConfigType
var defaultConfigJson = `
{
  "agent": {
    "collection_jitter": "10ms",
    "debug": true,
    "flush_interval": "10s",
    "flush_jitter": "10ms",
    "interval": "3s"
  },
  "controllers": {
    "d9019649-f2f7-431b-ac16-449ef7ca8ba1": {
	  "created_at": 1541487113347,
      "plugin_name": "http",
      "address": ":9999"
    }
  },
  "inputs": {
    "9f54b4b2-efb1-4f01-ad27-6b6da8d676a0": {
	  "created_at": 1541487113347,
      "plugin_name": "fake",
      "interval": "3s",
      "name_override": "fake",
      "point_map": ""
    }
  },
  "outputs": {
    "62b3328f-b1cf-4a9f-a456-e33e3a8c22c6": {
      "created_at": 1541487113347,
      "plugin_name": "redis",
      "url_address": "redis://localhost:6379/0"
    }
  }
}
`
var jsonConfigPath = "../configs/device_adaptor.json"
var InputSample = map[string]map[string]interface{}{
	"_base": {
		"interval":      "3s",
		"point_map":     "",
		"name_override": "",
		"created_at":    time.Now().UnixNano() / 1e6,
	},
	"modbus": {
		"name_override": "modbus",
		"plugin_name":   "modbus",
	},
	"fake": {
		"name_override": "fake",
		"plugin_name":   "fake",
	},
	"s7": {
		"name_override": "s7",
		"plugin_name":   "s7",
	},
}
var OutputSample = map[string]map[string]interface{}{
	"_base": {
		"created_at": time.Now().UnixNano() / 1e6,
	},
	"redis": {
		"plugin_name": "redis",
		"url_address": "redis://localhost:6379/0",
	},
	"file": {
		"plugin_name": "file",
		"files":       []string{"stdout"},
	},
}
var ControllerSample = map[string]map[string]interface{}{
	"_base": {
		"created_at": time.Now().UnixNano() / 1e6,
	},
	"http": {
		"plugin_name": "http",
		"address":     ":9999",
	},
}

func GenConfigSample(pluginType, pluginName string) (map[string]interface{}, error) {
	r := make(map[string]interface{})

	switch pluginType {
	case "inputs":
		if _, ok := InputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range InputSample["_base"] {
			r[k] = v
		}
		for k, v := range InputSample[pluginName] {
			r[k] = v
		}
	case "outputs":
		if _, ok := OutputSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range OutputSample["_base"] {
			r[k] = v
		}
		for k, v := range OutputSample[pluginName] {
			r[k] = v
		}
	case "controllers":
		if _, ok := ControllerSample[pluginName]; !ok {
			return nil, fmt.Errorf("unknown pluginName: %s", pluginName)
		}
		for k, v := range ControllerSample["_base"] {
			r[k] = v
		}
		for k, v := range ControllerSample[pluginName] {
			r[k] = v
		}
	}

	r["id"] = uuid.New().String()
	return r, nil
}

func FlushMemoryConfig() {
	b, _ := json.Marshal(MemoryConfig)
	ioutil.WriteFile(jsonConfigPath, b, 0644)
}

func GetConfigContent() []byte {
	var r = []byte(defaultConfigJson)
	if utils.IsExists(jsonConfigPath) {
		r, _ = ioutil.ReadFile(jsonConfigPath)
	}

	c := make(map[string]interface{})
	json.Unmarshal(r, &c)

	mapstructure.Decode(c, &MemoryConfig)
	json.Unmarshal([]byte(gjson.GetBytes(r, "agent").String()), &MemoryConfig.Agent)
	ioutil.WriteFile(jsonConfigPath, r, 0644)

	return r
}
