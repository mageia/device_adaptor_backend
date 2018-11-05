package configs

import (
	"deviceAdaptor/utils"
	"errors"
	"github.com/json-iterator/go"
	"github.com/pelletier/go-toml"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type FullyConfigOld struct {
	Agent       map[string]interface{} `json:"agent"`
	Controllers map[string][]map[string]interface{} `json:"controllers"`
	Inputs map[string][]map[string]interface{} `json:"inputs"`
	Outputs map[string][]map[string]interface{} `json:"outputs"`
}

var FullyConfig FullyConfigOld

var defaultConfig = `
[global_tags]

[agent]
debug = true
interval = "3s"

[[controllers.http]]
address = ":9999"    # 默认端口9999

[[outputs.redis]]
url_address = "redis://localhost:6379/0"

[[inputs.fake]]
name_override = "fake"
interval = "2s"
point_map = "../configs/point_map_opc.yml"
`

var jsonConfigPath = "../configs/device_adaptor.json"
var programConfigPath = "../configs/device_adaptor.toml"
var AgentSample = map[string]interface{}{
	"debug":             false,
	"interval":          "300s",
	"flush_interval":    "10s",
	"collection_jitter": "10ms",
	"flush_jitter":      "10ms",
}
var InputSample = map[string]map[string]interface{}{
	"_base": {
		"interval":      "3s",
		"point_map":     "",
		"name_override": "",
	},
	"modbus": {
		"name_override": "modbus",
	},
	"fake": {
		"name_override": "fake",
	},
	"s7": {
		"name_override": "s7",
	},
}
var OutputSample = map[string]map[string]interface{}{
	"_base": {
		//"data_format": "json",
	},
	"redis": {
		"url_address": "redis://localhost:6379/0",
	},
	"file": {
		"files": []string{"stdout"},
	},
}
var ControllerSample = map[string]map[string]interface{}{
	"_base": {},
	"http": {
		"address": ":9999",
	},
}

func isKeyChainValid(keyChain string, value interface{}) bool {
	l := strings.Split(keyChain, ".")
	if len(l) < 2 {
		return false
	}

	switch l[0] {
	case "inputs":
		switch l[1] {
		case "fake", "modbus", "s7", "http_listener":
			return true
		}
	case "outputs":
		switch l[1] {
		case "redis", "file":
			return true
		}
	case "controllers":
		switch l[1] {
		case "http":
			return true
		}
	}

	return false
}

func (f FullyConfigOld) SetOrDeleteJson(keyChain string, value interface{}) (e error) {
	if !isKeyChainValid(keyChain, value) {
		return errors.New("invalid key chain grammar")
	}

	configB, _ := json.Marshal(f)
	defer func() {
		json.Unmarshal(configB, &f)
		ioutil.WriteFile(jsonConfigPath, configB, 0644)
		JsonToToml(jsonConfigPath, programConfigPath)
	}()

	switch value.(type) {
	case nil:
		if configB, e = sjson.DeleteBytes(configB, keyChain); e != nil {
			log.Println(e)
			return e
		}
	case []interface{}:
		return errors.New("cannot set array as value")
	default:
		//TODO: name_override 检测是否重复
		if configB, e = sjson.SetBytes(configB, keyChain, value); e != nil {
			log.Println(e)
			return
		}
	}

	return e
}

func initDefaultConfig() FullyConfigOld {
	var _fullyConfig = FullyConfigOld{
		Agent: make(map[string]interface{}, 0),
		Inputs: make(map[string][]map[string]interface{}, 0),
		Outputs: make(map[string][]map[string]interface{}, 0),
		Controllers: make(map[string][]map[string]interface{}, 0),
	}

	defaultConfigTree, _ := toml.Load(defaultConfig)

	for ki, vi := range AgentSample {
		_fullyConfig.Agent[ki] = vi
	}

	for k, v := range defaultConfigTree.ToMap() {
		if k == "agent" {
			for kk, vv := range v.(map[string]interface{}) {
				_fullyConfig.Agent[kk] = vv
			}
			continue
		}

		switch vV := v.(type) {
		case []interface{}:
			for _, vvV := range vV {
				log.Println(k, vvV)
			}
		case map[string]interface{}:
			for kk, vv := range vV {
				switch vv.(type) {
				case []interface{}:
					for _, vvV := range vv.([]interface{}) {
						switch k {
						case "inputs":
							item := vvV.(map[string]interface{})
							for ki, vi := range InputSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range InputSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Inputs[kk] = append(_fullyConfig.Inputs[kk], item)
						case "outputs":
							item := vvV.(map[string]interface{})
							for ki, vi := range OutputSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range OutputSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Outputs[kk] = append(_fullyConfig.Outputs[kk], item)
						case "controllers":
							item := vvV.(map[string]interface{})
							for ki, vi := range ControllerSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range ControllerSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Controllers[kk] = append(_fullyConfig.Controllers[kk], item)
						}
					}
				}
			}
		}
	}

	FullyConfig = _fullyConfig

	configB, _ := json.Marshal(_fullyConfig)
	ioutil.WriteFile(jsonConfigPath, configB, 0644)
	return _fullyConfig
}

func ReLoadConfig() {
	defer JsonToToml(jsonConfigPath, programConfigPath)

	if !utils.IsExists(jsonConfigPath) {
		initDefaultConfig()
	} else {
		c, e := ioutil.ReadFile(jsonConfigPath)
		if e != nil {
			log.Println(e)
			return
		}
		if e := json.Unmarshal(c, &FullyConfig); e != nil {
			log.Println(e)
			return
		}
	}
}

func isJsonConfigValid(c map[string]interface{}) bool {
	//TODO
	return true
}

func jsonConfigToMap(jsonPath string) (map[string]interface{}, error) {
	var result map[string]interface{}

	srcFile, err := os.Open(jsonPath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer srcFile.Close()

	byteValue, _ := ioutil.ReadAll(srcFile)
	if json.Unmarshal(byteValue, &result) != nil {
		log.Println(err)
		return nil, err
	}

	if !isJsonConfigValid(result) {
		return nil, errors.New("invalid json config")
	}

	return result, nil
}

func JsonToToml(srcPath, dstPath string) (err error) {
	var r map[string]interface{}

	r, err = jsonConfigToMap(srcPath)
	if err != nil {
		f, _ := json.Marshal(initDefaultConfig())
		json.Unmarshal(f, &r)
	}

	for k, v := range r {
		switch k {
		case "inputs", "outputs", "controllers":
			for kv, vv := range v.(map[string]interface{}) {
				switch vva := vv.(type) {
				case []interface{}:
					if len(vva) == 0 {
						delete(r[k].(map[string]interface{}), kv)
					}
				}
			}
		}
	}

	t, _ := toml.TreeFromMap(r)
	f, e := os.OpenFile(dstPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if e != nil {
		log.Println(e)
	}
	defer f.Close()

	if _, e := t.WriteTo(f); e != nil {
		log.Println(e)
		return e
	}
	return nil
}
