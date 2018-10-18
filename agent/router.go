package agent

import (
	"github.com/fatih/structs"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"time"
)

var ReloadSignal = make(chan struct{})

func Index(c *gin.Context) {
	c.HTML(200, "test.html", gin.H{
		"Title": "Index",
		"Global": map[string]interface{}{
			"debug":    true,
			"interval": "3s",
		},
		"GlobalConfig": structs.New(A.Config.Global).Map(),

		"Plugins": map[string]interface{}{
			"Controllers": map[string]interface{}{
				"http": map[string]interface{}{
					"enable":  true,
					"address": ":9999",
				},
				"redis": map[string]interface{}{
					"enable": false,
				},
				"websocket": map[string]interface{}{
					"enable": false,
				},
			},
			"Inputs": map[string]interface{}{
				"s7":     map[string]interface{}{},
				"modbus": map[string]interface{}{},
				"ftp":    map[string]interface{}{},
			},
			"Outputs": map[string]interface{}{
				"redis": map[string]interface{}{},
				"file":  map[string]interface{}{},
			},
			"Parsers": map[string]interface{}{
				"csv": map[string]interface{}{},
			},
			"Serializers": map[string]interface{}{
				"json": map[string]interface{}{},
			},
		},
	})
}

func Update(c *gin.Context) {
	viper.SetConfigName("device_adaptor")
	//viper.AddConfigPath("../configs")
	if e := viper.ReadInConfig(); e != nil {
		c.JSON(400, "viper.ReadInConfig Error")
		return
	}
	viper.Set("global_tags.test", time.Now().String())
	viper.WriteConfig()

	c.JSON(200, viper.GetString("global_tags.test"))
	ReloadSignal <- struct{}{}
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

	router.Static("/", "../agent/dist")

	//router.LoadHTMLGlob("../agent/dist/*")
	//router.LoadHTMLFiles("../agent/dist/*")
	//router.GET("/", Index)
	//router.GET("/update", Update)
	//
	//plugins := router.Group("/plugins")
	//plugins.GET("/", Index)
	//plugins.GET("/inputs")
	//plugins.GET("/outputs")
	//plugins.GET("/controllers")
	//plugins.GET("/parsers")
	//plugins.GET("/serializers")
	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
