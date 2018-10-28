package agent

import (
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"strings"
)

type FullyConfigOld struct {
	Agent       map[string]interface{} `json:"agent"`
	Controllers map[string][]map[string]interface{} `json:"controllers"`
	Inputs map[string][]map[string]interface{} `json:"inputs"`
	Outputs map[string][]map[string]interface{} `json:"outputs"`
}

var ReloadSignal = make(chan struct{})
var fullyConfig FullyConfigOld
var currentConfig = make(map[string]interface{})
var agentSample = map[string]interface{}{
	"debug":               false,
	"interval":            "300s",
	"flush_interval":      "10s",
	"collection_jitter":   "10ms",
	"flush_jitter":        "10ms",
	"metric_batch_size":   0,
	"metric_buffer_limit": 0,
}
var inputSample = map[string]map[string]interface{}{
	"_base": {
		"_base_test":    "11111111111111",
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

func isValidKeyChain(keyChain string) bool {
	l := strings.Split(keyChain, ".")
	if len(l) < 2 {
		if len(l) == 1 && l[0] == "global_tags" {
			return true
		}
		return false
	}
	l1 := l[1]
	s := strings.Index(l1, "[")
	e := strings.Index(l1, "]")
	if s*e < 0 {
		return false
	}
	if s > 0 && e > s {
		l1 = l[1][:s]
	}

	switch l[0] {
	case "agent", "global_tags":
		return true
	case "controllers":
		if _, ok := controllers.Controllers[l1]; ok {
			return true
		}
	case "inputs":
		if _, ok := inputs.Inputs[l1]; ok {
			return true
		}
	case "outputs":
		if _, ok := outputs.Outputs[l1]; ok {
			return true
		}
	}
	return false
}

func updateMap(t map[string]interface{}, keyChain string, value interface{}) bool {
	if !isValidKeyChain(keyChain) {
		return false
	}

	l := strings.Split(keyChain, ".")
	for i, k := range l {
		s := strings.Index(k, "[")
		e := strings.Index(k, "]")

		index := 0
		if s > 0 && e > s {
			ii, e := strconv.Atoi(k[s+1 : e])
			if e != nil || i < 0 {
				return false
			}
			index = ii
			k = k[:s]
		}

		if _, ok := t[k]; !ok {
			if i == len(l)-1 {
				t[k] = value
				return true
			}
			t[k] = make(map[string]interface{})
		}

		switch tV := t[k].(type) {
		case map[string]interface{}:
			t = tV
		case []interface{}:
			if index < 0 {
				return false
			}
			if index == len(tV) {
				tV = append(tV, map[string]interface{}{})
				t[k] = tV
			}
			if index < len(tV) {
				if tVV, ok := tV[index].(map[string]interface{}); ok {
					t = tVV
				}
			}
			if index > len(tV) {
				return false
			}
		default:
			t[k] = value
			return true
		}
	}
	return true
}

func checkIndex(k string) (string, int, bool) {
	index := -1
	s := strings.Index(k, "[")
	e := strings.Index(k, "]")
	if s*e < 0 {
		return k, index, false
	}
	if s > 0 && e > s {
		ii, e := strconv.Atoi(k[s+1 : e])
		if e != nil || ii < 0 {
			return k, index, false
		}
		return k[:s], ii, true
	}
	return k, index, true
}

func updateMapNew(t map[string]interface{}, keyChain string, value interface{}) bool {
	if !isValidKeyChain(keyChain) {
		return false
	}

	l := strings.Split(keyChain, ".")
	for i, k := range l {
		k, index, ok := checkIndex(k)
		if !ok {
			return false
		}

		switch tV := t[k].(type) {
		case map[string]interface{}:
			if i == len(l)-1 { //End, 循环的出口1
				switch value.(type) {
				case nil:
					delete(t, k)
				default:
					t[k] = value
				}
				return true
			}
			t = tV //next
		case []interface{}:
			log.Println(index, len(tV))

			if index == len(tV) && i == len(l)-1 { //len(tV) == index: 扩展一个
				switch value.(type) {
				case nil:
				default:
					t[k] = append(t[k].([]interface{}), value)
				}
				return true
			} else if len(tV) > index && index >= 0 { //remove one
				if i == len(l)-1 {
					switch value.(type) {
					case []interface{}:
						return false
					case nil:
						t[k] = append(tV[:index], tV[index+1:]...)
					default:
						t[k].([]interface{})[index] = value
					}
					return true
				} else {
					switch tVV := tV[index].(type) {
					case map[string]interface{}:
						t = tVV
					default: //TODO: 数组嵌套层数受限
						tV[index] = make(map[string]interface{})
						t = tV[index].(map[string]interface{})
					}
				}
			} else if index == -1 && (len(tV) == 0 || i == len(l)-1) {
				switch value.(type) {
				case nil:
				default:
					t[k] = value
				}
				return true
			} else {
				return false
			}
		case string, bool, int:
			switch value.(type) {
			case nil:
				delete(t, k)
			default:
				t[k] = value
			}
			return true
		default:
			if index < 0 {
				switch value.(type) {
				case nil:
					delete(t, k)
				default:
					t[k] = value
				}
				return true
			} else if index == 0 {
				switch value.(type) {
				case nil:
					delete(t, k)
				default:
					t[k] = []interface{}{value}
				}
				return true
			}
		}
	}

	return true
}

func removeMapByKey(t map[string]interface{}, keyChain string) bool {
	return false
}

func toMapSlice(v interface{}) []map[string]interface{} {
	switch vV := v.(type) {
	case map[string]interface{}:
		return []map[string]interface{}{vV}
	case []interface{}:
		r := make([]map[string]interface{}, 0)
		for _, vVItem := range vV {
			switch vV1 := vVItem.(type) {
			case map[string]interface{}:
				r = append(r, vV1)
			}
		}
		return r
	}
	return []map[string]interface{}{}
}

func flushFullyConfig() FullyConfigOld {
	var _fullyConfig = FullyConfigOld{
		Agent: make(map[string]interface{}, 0),
		Inputs: make(map[string][]map[string]interface{}, 0),
		Outputs: make(map[string][]map[string]interface{}, 0),
		Controllers: make(map[string][]map[string]interface{}, 0),
	}

	configItem := make(map[string]interface{})
	for ki, vi := range agentSample {
		configItem[ki] = vi
	}
	for k, v := range viper.GetStringMap("agent") {
		configItem[k] = v
	}
	_fullyConfig.Agent = configItem

	for k := range inputs.Inputs {
		configs := make([]map[string]interface{}, 0)
		for _, item := range toMapSlice(viper.GetStringMap("inputs")[k]) {
			configItem := make(map[string]interface{})
			for ki, vi := range inputSample[k] {
				configItem[ki] = vi
			}
			for ki, vi := range inputSample["_base"] {
				configItem[ki] = vi
			}
			for ki, vi := range item {
				configItem[ki] = vi
			}
			configs = append(configs, configItem)
		}
		_fullyConfig.Inputs[k] = configs
	}
	for k := range outputs.Outputs {
		configs := make([]map[string]interface{}, 0)
		for _, item := range toMapSlice(viper.GetStringMap("outputs")[k]) {
			configItem := make(map[string]interface{})
			for ki, vi := range outputSample[k] {
				configItem[ki] = vi
			}
			for ki, vi := range outputSample["_base"] {
				configItem[ki] = vi
			}
			for ki, vi := range item {
				configItem[ki] = vi
			}
			configs = append(configs, configItem)
		}
		_fullyConfig.Outputs[k] = configs
	}
	for k := range controllers.Controllers {
		configs := make([]map[string]interface{}, 0)
		for _, item := range toMapSlice(viper.GetStringMap("controllers")[k]) {
			configItem := make(map[string]interface{})
			for ki, vi := range controllerSample[k] {
				configItem[ki] = vi
			}
			for ki, vi := range controllerSample["_base"] {
				configItem[ki] = vi
			}
			for ki, vi := range item {
				configItem[ki] = vi
			}
			configs = append(configs, configItem)
		}
		_fullyConfig.Controllers[k] = configs
	}

	fullyConfig = _fullyConfig
	return _fullyConfig
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

func removePlugin(c *gin.Context) {
	var body []string
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	for _, k := range body {
		removeMapByKey(currentConfig, k)
	}
	flushFullyConfig()
	c.JSON(200, currentConfig)
}
func updatePlugin(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	for k, v := range body {
		updateMapNew(currentConfig, k, v)
	}
	flushFullyConfig()
	c.JSON(200, currentConfig)
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
	router.POST("/removePlugin", removePlugin)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}

func LoadConfig() {
	viper.SetConfigName("device_adaptor")
	viper.AddConfigPath("../configs")
	if e := viper.ReadInConfig(); e != nil {
		log.Fatalln("viper read config failed")
	}
	fullyConfig = flushFullyConfig()
	currentConfig = viper.AllSettings()
}
