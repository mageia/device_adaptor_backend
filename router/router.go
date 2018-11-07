package router

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/configs"
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"deviceAdaptor/plugins/outputs"
	_ "deviceAdaptor/statik"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/rakyll/statik/fs"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

func getAll(c *gin.Context) {
	var availablePluginName = make(map[string][]string)
	for k := range inputs.Inputs {
		availablePluginName["inputs"] = append(availablePluginName["inputs"], k)
	}
	for k := range outputs.Outputs {
		availablePluginName["outputs"] = append(availablePluginName["outputs"], k)
	}
	for k := range controllers.Controllers {
		availablePluginName["controllers"] = append(availablePluginName["controllers"], k)
	}

	c.JSON(200, gin.H{
		"current_config":        configs.MemoryConfig,
		"available_plugin_name": availablePluginName,
	})
}

func getConfig(c *gin.Context) {
	id := c.Param("id")

	switch c.Param("pluginType") {
	case "agent":
		c.JSON(200, configs.MemoryConfig.Agent)
	case "inputs":
		c.JSON(200, configs.MemoryConfig.Inputs[id])
	case "outputs":
		c.JSON(200, configs.MemoryConfig.Outputs[id])
	case "controllers":
		c.JSON(200, configs.MemoryConfig.Controllers[id])
	default:
		c.Error(fmt.Errorf("can't find config by params: %v", c.Params))
	}
}

func deleteConfig(c *gin.Context) {
	id := c.Param("id")

	switch c.Param("pluginType") {
	case "inputs":
		delete(configs.MemoryConfig.Inputs, id)
		c.JSON(200, gin.H{"msg": "OK"})
	case "outputs":
		delete(configs.MemoryConfig.Outputs, id)
		c.JSON(200, gin.H{"msg": "OK"})
	case "controllers":
		delete(configs.MemoryConfig.Controllers, id)
		c.JSON(200, gin.H{"msg": "OK"})
	default:
		c.Error(fmt.Errorf("invalid pluginType: %v", c.Param("pluginType")))
	}
}

func putConfig(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}
	id := c.Param("id")

	switch c.Param("pluginType") {
	case "agent":
		b, _ := json.Marshal(body)
		if e := json.Unmarshal(b, configs.MemoryConfig.Agent); e != nil {
			c.Error(fmt.Errorf("update agent config failed: %v", e))
		} else {
			c.JSON(200, configs.MemoryConfig.Agent)
		}

	case "inputs":
		if _, ok := configs.MemoryConfig.Inputs[id]; ok {
			for k, v := range body {
				if _, ok := configs.MemoryConfig.Inputs[id][k]; ok {
					configs.MemoryConfig.Inputs[id][k] = v
				}
			}
		}
		c.JSON(200, configs.MemoryConfig.Inputs[id])
	case "outputs":
		if _, ok := configs.MemoryConfig.Outputs[id]; ok {
			for k, v := range body {
				if _, ok := configs.MemoryConfig.Outputs[id][k]; ok {
					configs.MemoryConfig.Outputs[id][k] = v
				}
			}
		}
		c.JSON(200, configs.MemoryConfig.Inputs[id])
	case "controllers":
		if _, ok := configs.MemoryConfig.Controllers[id]; ok {
			for k, v := range body {
				if _, ok := configs.MemoryConfig.Controllers[id][k]; ok {
					configs.MemoryConfig.Controllers[id][k] = v
				}
			}
		}
		c.JSON(200, configs.MemoryConfig.Controllers[id])
	default:
		c.Error(fmt.Errorf("can't find config by params: %v", c.Params))
	}
	configs.FlushMemoryConfig()
}

func postConfig(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	pluginType := c.Param("pluginType")
	pluginName, ok := body["plugin_name"]
	if !ok {
		c.Error(fmt.Errorf("plugin_name must by specified"))
		return
	}

	s, e := configs.GenConfigSample(pluginType, pluginName.(string))
	if e != nil {
		c.Error(e)
		return
	}
	for k := range s {
		if v, ok := body[k]; ok {
			s[k] = v
		}
	}

	id := s["id"].(string)
	delete(s, "id")

	switch pluginType {
	case "inputs":
		for _, v := range configs.MemoryConfig.Inputs {
			if v["name_override"].(string) == s["name_override"].(string) {
				c.Error(fmt.Errorf("input alias: %s already exists", v["name_override"]))
				return
			}
		}
		configs.MemoryConfig.Inputs[id] = s
	case "outputs":
		configs.MemoryConfig.Outputs[id] = s
	case "controllers":
		configs.MemoryConfig.Controllers[id] = s
	default:
		c.Error(fmt.Errorf("unknown pluginType: %s", pluginType))
		return
	}

	c.JSON(200, gin.H{id: s})
}

func getStatic(fs http.FileSystem, p string) ([]byte, error) {
	pp := path.Join("/dist", p)
	x, e := fs.Open(pp)
	if e != nil {
		return nil, e
	}
	return ioutil.ReadAll(x)
}

type Login struct {
	Username string `form:"username" json:"username" xml:"username"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func login(c *gin.Context) {
	var body Login
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(err)
		return
	}

	if body.Username != "admin" || body.Password != "admin" {
		c.Error(errors.New("unauthorized"))
		return
	}

	claims := CustomClaims{
		body.Username,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 20).Unix(),
		},
	}
	token, e := NewJWT().CreateToken(&claims)
	if e != nil {
		c.Error(e)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "expire": claims.ExpiresAt})
}

func refresh(c *gin.Context) {
	cS, ok := c.Get("claims")
	if !ok {
		c.AbortWithStatusJSON(403, gin.H{"error": "refresh failed, please re-login."})
		return
	}

	if claims, ok := cS.(*CustomClaims); ok {
		claims.StandardClaims.ExpiresAt = time.Now().Add(time.Minute * 5).Unix()
		token, e := NewJWT().CreateToken(claims)
		if e != nil {
			c.Error(e)
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
	c.AbortWithStatusJSON(403, gin.H{"error": "refresh failed, please re-login"})
}

func InitRouter(debug bool) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) != 0 {
			c.AbortWithStatusJSON(400, c.Errors.JSON())
			return
		}

		if strings.HasPrefix(c.Request.URL.Path, "/plugin") && c.Request.Method != "GET" {
			go func() {
				configs.FlushMemoryConfig()
				agent.ReloadSignal <- struct{}{}
			}()
		}
	})

	sFs, e := fs.New()
	if e != nil {
		log.Fatalln(e)
	}

	router.GET("/", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "index.html")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})
	router.GET("/favicon.ico", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "favicon.ico")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})

	router.GET("/_nuxt/*filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/_nuxt", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})
	router.GET("/image/*filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/image", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})

	auth := router.Group("/auth")
	auth.POST("/login", login)
	auth.POST("/refresh", JWTAuthMiddleware, refresh)

	router.GET("/getInitData", JWTAuthMiddleware, getAll)
	pG := router.Group("/plugin/:pluginType", JWTAuthMiddleware)
	pG.GET("/:id", getConfig)
	pG.PUT("/:id", putConfig)
	pG.DELETE("/:id", deleteConfig)
	pG.POST("/", postConfig)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
