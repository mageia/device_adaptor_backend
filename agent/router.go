package agent

import (
	"github.com/gin-gonic/gin"
	"github.com/theherk/viper"
	"time"
)

var ReloadSignal = make(chan struct{})

func Index(c *gin.Context) {
	c.HTML(200, "index", gin.H{
		"Title": "Index",
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
		"Global": map[string]interface{}{
			"debug":    true,
			"interval": "3s",
		},
	})
}

func Update(c *gin.Context) {
	viper.SetConfigName("device_adaptor")
	viper.AddConfigPath("../configs")
	if e := viper.ReadInConfig(); e != nil {
		c.JSON(400, "viper.ReadInConfig Error")
		return
	}
	viper.Set("global_tags.test", time.Now().String())
	viper.WriteConfig()

	c.JSON(200, viper.GetString("global_tags.test"))
	ReloadSignal <- struct{}{}
}

func InitRouter() *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			c.AbortWithStatusJSON(400, c.Errors.JSON())
		}
	})
	router.LoadHTMLGlob("../configs/templates/*")
	router.GET("/", Index)
	router.GET("/update", Update)

	//plugins := router.Group("/plugins")
	//plugins.GET("/", Index)
	//plugins.GET("/inputs")
	//plugins.GET("/outputs")
	//plugins.GET("/controllers")
	//plugins.GET("/parsers")
	//plugins.GET("/serializers")
	return router
}
