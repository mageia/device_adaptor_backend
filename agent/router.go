package agent

import (
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"strings"
	"time"
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

func updateMap(target map[string]interface{}, keyChain string, value interface{}) {
	t := target
	l := strings.Split(keyChain, ".")
	for i, k := range l {
		s := strings.Index(k, "[")
		e := strings.Index(k, "]")
		index := 0
		if s > 0 && e > s {
			index, _ = strconv.Atoi(k[s+1 : e])
			k = k[:s]
		}

		if _, ok := t[k]; !ok {
			if i == len(l)-1 {
				t[k] = value
				break
			}
			t[k] = make(map[string]interface{})
		}

		switch tV := t[k].(type) {
		case map[string]interface{}:
			t = tV
		case []interface{}:
			if index == len(tV) {
				tV = append(tV, map[string]interface{}{})
				t[k] = tV
			}
			if index < len(tV) {
				if tVV, ok := tV[index].(map[string]interface{}); ok {
					t = tVV
				}
			}
		default:
			t[k] = value
			break
		}
	}
}

func Update(c *gin.Context) {
	//viper.SetConfigName("device_adaptor")
	////viper.AddConfigPath("../configs")
	//if e := viper.ReadInConfig(); e != nil {
	//	c.JSON(400, "viper.ReadInConfig Error")
	//	return
	//}
	viper.Set("global_tags.test", time.Now().String())
	viper.WriteConfig()

	c.JSON(200, viper.GetString("global_tags.test"))
	ReloadSignal <- struct{}{}
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

func createPlugin(c *gin.Context) {
	switch c.Param("pluginName") {
	case "inputs":
		c.JSON(200, "create input")
	case "outputs":
		c.JSON(200, "create output")
	case "controllers":
		c.JSON(200, "create controller")
	default:
		c.Error(errors.New("unknown plugin name"))
	}
}
func removePlugin(c *gin.Context) {

}
func updatePlugin(c *gin.Context) {
	pluginId := c.Param("id")
	if pluginId == "" {
		c.Error(errors.New("invalid plugin type"))
		return
	}
	//var body struct {
	//	Key   string      `json:"key" binding:"required"`
	//	Value interface{} `json:"value" binding:"required"`
	//}
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	switch c.Param("pluginName") {
	case "input":
		for k, v := range body {
			updateMap(currentConfig, k, v)
		}
		flushFullyConfig()
		c.JSON(200, currentConfig)
	case "output":
		c.JSON(200, "create output")
	case "controller":
		c.JSON(200, "create controller")
	default:
		c.Error(errors.New("unknown plugin name"))
	}
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

	router.GET("/initData", getInitData)

	plugin := router.Group("/plugin/:pluginName")
	plugin.POST("/", createPlugin)
	plugin.PUT("/:id", updatePlugin)
	plugin.DELETE("/:id", removePlugin)

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
