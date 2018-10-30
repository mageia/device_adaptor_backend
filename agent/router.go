package agent

import (
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"
	"github.com/tidwall/sjson"
	"io/ioutil"
	"log"
	"os"
)

type FullyConfigOld struct {
	Agent       map[string]interface{} `json:"agent"`
	Controllers map[string][]map[string]interface{} `json:"controllers"`
	Inputs map[string][]map[string]interface{} `json:"inputs"`
	Outputs map[string][]map[string]interface{} `json:"outputs"`
}

var tomlConfigPath = "../configs/default.toml"
var jsonConfigPath = "../configs/device_adaptor.json"
var programConfigPath = "../configs/device_adaptor.toml"
var ReloadSignal = make(chan struct{})
var fullyConfig FullyConfigOld
var agentSample = map[string]interface{}{
	"debug":             false,
	"interval":          "300s",
	"flush_interval":    "10s",
	"collection_jitter": "10ms",
	"flush_jitter":      "10ms",
	//"metric_batch_size":   0,
	//"metric_buffer_limit": 0,
}
var inputSample = map[string]map[string]interface{}{
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
var outputSample = map[string]map[string]interface{}{
	"_base": {
		"data_format": "json",
	},
	"redis": {
		"output_channel": "output_test",
		"url_address":    "redis://localhost:6379/0",
	},
	"file": {
		"files": []string{"stdout"},
	},
}
var controllerSample = map[string]map[string]interface{}{
	"_base": {},
	"http": {
		"address": ":9999",
	},
}

func (f FullyConfigOld) setOrDeleteJson(keyChain string, value interface{}) (e error) {
	configB, _ := json.Marshal(f)
	defer func() {
		json.Unmarshal(configB, &f)
		ioutil.WriteFile(jsonConfigPath, configB, 0644)
		jsonToToml()
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
		if configB, e = sjson.SetBytes(configB, keyChain, value); e != nil {
			log.Println(e)
			return
		}
	}

	return e
}

func IsExists(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func initLoadConfig() FullyConfigOld {
	var _fullyConfig = FullyConfigOld{
		Agent: make(map[string]interface{}, 0),
		Inputs: make(map[string][]map[string]interface{}, 0),
		Outputs: make(map[string][]map[string]interface{}, 0),
		Controllers: make(map[string][]map[string]interface{}, 0),
	}

	if !IsExists(tomlConfigPath) {
		return _fullyConfig
	}

	currentConfigTree, _ := toml.LoadFile(tomlConfigPath)

	for ki, vi := range agentSample {
		_fullyConfig.Agent[ki] = vi
	}

	for k, v := range currentConfigTree.ToMap() {
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
							for ki, vi := range inputSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range inputSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Inputs[kk] = append(_fullyConfig.Inputs[kk], item)
						case "outputs":
							item := vvV.(map[string]interface{})
							for ki, vi := range outputSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range outputSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Outputs[kk] = append(_fullyConfig.Outputs[kk], item)
						case "controllers":
							item := vvV.(map[string]interface{})
							for ki, vi := range controllerSample["_base"] {
								item[ki] = vi
							}
							for ki, vi := range controllerSample[kk] {
								item[ki] = vi
							}
							_fullyConfig.Controllers[kk] = append(_fullyConfig.Controllers[kk], item)
						}
					}
				}
			}
		}
	}

	fullyConfig = _fullyConfig

	configB, _ := json.Marshal(_fullyConfig)
	ioutil.WriteFile(jsonConfigPath, configB, 0644)
	return _fullyConfig
}

func jsonToToml() {
	viper.SetConfigFile(jsonConfigPath)
	viper.ReadInConfig()

	viper.SetConfigFile(programConfigPath)
	viper.SetConfigType("toml")
	if e := viper.WriteConfig(); e != nil {
		log.Println(e)
	}
	viper.Reset()
}

func LoadConfig() {
	defer jsonToToml()

	if !IsExists(jsonConfigPath) {
		initLoadConfig()
	} else {
		c, e := ioutil.ReadFile(jsonConfigPath)
		if e != nil {
			log.Println(e)
			return
		}
		if e := json.Unmarshal(c, &fullyConfig); e != nil {
			log.Println(e)
			return
		}
	}
}

func getInitData(c *gin.Context) {
	r := make(map[string]interface{})
	availableInput := make([]string, 0)
	availableOutput := make([]string, 0)

	for k := range inputs.Inputs {
		availableInput = append(availableInput, k)
	}
	for k := range outputs.Outputs {
		availableOutput = append(availableOutput, k)
	}

	r["available_input"] = availableInput
	r["available_output"] = availableOutput
	r["fully_config"] = fullyConfig

	c.JSON(200, r)
}

func updatePlugin(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	for k, v := range body {
		fullyConfig.setOrDeleteJson(k, v)
	}
	ReloadSignal <- struct{}{}
	c.JSON(200, fullyConfig)
}

func InitRouter(debug bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			c.AbortWithStatusJSON(400, c.Errors.JSON())
		}
	})

	router.StaticFile("/", "../agent/dist/index.html")
	router.Static("/_nuxt", "../agent/dist/_nuxt")
	router.Static("/image", "../agent/dist/image")

	router.GET("/getInitData", getInitData)
	router.POST("/updatePlugin", updatePlugin)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
