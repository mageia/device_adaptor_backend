package agent

import (
	"deviceAdaptor/configs"
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/gin-gonic/gin/json"
	"github.com/tidwall/gjson"
)

var ReloadSignal = make(chan struct{}, 1)

func getInitData(c *gin.Context) {
	r := make(map[string]interface{})

	availableInput := make(map[string]map[string]interface{}, 0)
	availableOutput := make(map[string]map[string]interface{}, 0)
	availableController := make(map[string]map[string]interface{}, 0)
	for k := range inputs.Inputs {
		if _, ok := configs.InputSample[k]; ok {
			availableInput[k] = make(map[string]interface{})
			for kk, vv := range configs.InputSample["_base"] {
				availableInput[k][kk] = vv
			}
			for kk, vv := range configs.InputSample[k] {
				availableInput[k][kk] = vv
			}
		}
	}
	for k := range outputs.Outputs {
		if _, ok := configs.OutputSample[k]; ok {
			availableOutput[k] = make(map[string]interface{})
			for kk, vv := range configs.OutputSample["_base"] {
				availableOutput[k][kk] = vv
			}
			for kk, vv := range configs.OutputSample[k] {
				availableOutput[k][kk] = vv
			}
		}
	}
	for k := range controllers.Controllers {
		if _, ok := configs.ControllerSample[k]; ok {
			availableController[k] = make(map[string]interface{})
			for kk, vv := range configs.ControllerSample["_base"] {
				availableController[k][kk] = vv
			}
			for kk, vv := range configs.ControllerSample[k] {
				availableController[k][kk] = vv
			}
		}
	}

	r["availableInput"] = availableInput
	r["availableOutput"] = availableOutput
	r["availableController"] = availableController
	r["fully_config"] = configs.FullyConfig
	c.JSON(200, r)
}

func getTest(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}
	r := make(map[string]interface{})
	configB, _ := json.Marshal(configs.FullyConfig)
	for k := range body {
		r[k] = gjson.GetBytes(configB, k).Value()
	}
	c.JSON(200, r)
}

func updatePlugin(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	for k, v := range body {
		if e := configs.FullyConfig.SetOrDeleteJson(k, v); e != nil {
			c.JSON(400, gin.H{"error": e.Error()})
			return
		}
	}
	c.JSON(200, configs.FullyConfig)
}

func InitRouter(debug bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			c.AbortWithStatusJSON(400, c.Errors.JSON())
		}

		if c.Request.URL.Path == "/updatePlugin" {
			go func() {
				ReloadSignal <- struct{}{}
			}()
		}
	})

	router.StaticFile("/", "../frontend/dist/index.html")
	router.Static("/_nuxt", "../frontend/dist/_nuxt")
	router.Static("/image", "../frontend/dist/image")

	router.GET("/getInitData", getInitData)
	router.POST("/getTest", getTest)
	router.POST("/updatePlugin", updatePlugin)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
