package configs

import (
	"device_adaptor/utils"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

func GetInputConfigById(id string) (map[string]interface{}, bool) {
	for _, v := range MemoryConfig.Inputs {
		if vN, ok := v["name_override"]; ok && vN == id {
			return v, true
		}
		if vN, ok := v["id"]; ok && vN == id {
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
	r["created_at"] = time.Now().UnixNano() / 1e6
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
	log.Debug().Interface("MemoryConfig", MemoryConfig).Msg("FlushMemoryConfig")
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
		if runtime.GOARCH == "arm" {
			jsonConfigPath = "./device_adaptor.json"
		} else {
			jsonConfigPath = "/var/device_adaptor/device_adaptor.json"
		}
	}
}
