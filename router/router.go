package router

import (
	"deviceAdaptor/agent"
	"deviceAdaptor/configs"
	"deviceAdaptor/internal/points"
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
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"
)

type Login struct {
	Username string `form:"username" json:"username" xml:"username"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func getStatic(fs http.FileSystem, p string) ([]byte, error) {
	pp := path.Join("/dist", p)
	x, e := fs.Open(pp)
	if e != nil {
		return nil, e
	}
	return ioutil.ReadAll(x)
}

func getConfigSample(c *gin.Context) {
	var availableInputs = make(map[string][]configs.ConfigSample)
	var availableOutputs = make(map[string][]configs.ConfigSample)
	var availableControllers = make(map[string][]configs.ConfigSample)

	for ki := range inputs.Inputs {
		availableInputs[ki], _ = configs.GenConfigSampleArray("inputs", ki)
	}
	for ko := range outputs.Outputs {
		availableOutputs[ko], _ = configs.GenConfigSampleArray("outputs", ko)
	}
	for kc := range controllers.Controllers {
		availableControllers[kc], _ = configs.GenConfigSampleArray("controllers", kc)
	}

	var availablePluginName = map[string]interface{}{
		"inputs":      availableInputs,
		"outputs":     availableOutputs,
		"controllers": availableControllers,
	}
	c.JSON(200, availablePluginName)
}
func getCurrentConfig(c *gin.Context) {
	c.JSON(200, configs.MemoryConfig)
}
func getPointMap(c *gin.Context) {
	id := c.Param("id")
	iC, ok := configs.GetInputConfigById(id)
	if !ok {
		c.Error(fmt.Errorf("can't find input plugin by id: %s", id))
		return
	}

	pointArray := make([]points.PointDefine, 0)
	pointMap := make(map[string]points.PointDefine)
	points.SqliteDB.Where("input_name = ?", iC["name_override"]).Find(&pointArray)
	for _, p := range pointArray {
		pointMap[p.PointKey] = p
	}

	b, _ := json.Marshal(pointMap)
	c.JSON(200, gin.H{
		"point_map_content": string(b),
		"point_map_path":    "", //TODO  path
	})
}
func putPointMap(c *gin.Context) {
	id := c.Param("id")
	iC, ok := configs.GetInputConfigById(id)
	if !ok {
		c.Error(fmt.Errorf("can't find input plugin by id: %s", id))
		return
	}

	inputName := iC["name_override"].(string)

	var body struct {
		PointMapPath    string `json:"point_map_path"`
		PointMapContent string `json:"point_map_content"`
	}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	pointMap := make(map[string]points.PointDefine)
	if e := yaml.UnmarshalStrict([]byte(body.PointMapContent), &pointMap); e != nil {
		if e = json.Unmarshal([]byte(body.PointMapContent), &pointMap); e != nil {
			c.Error(fmt.Errorf("point_map_content is neither json nor yaml format: %v", e))
			return
		}
	}
	pointMapKeys := make([]string, len(pointMap))

	i := 0
	for _, v := range pointMap {
		pointMapKeys[i] = v.PointKey
		i += 1
	}

	//points.SqliteDB.Unscoped().Delete(&points.PointDefine{}, "input_name = ? AND address NOT IN ?", inputName, pointMapKeys)
	points.SqliteDB.Unscoped().Where("input_name = ?", inputName).Not("point_key", pointMapKeys).Delete(points.PointDefine{})
	for k, v := range pointMap {
		v.InputName = inputName
		if v.Name == "" {
			v.Name = k
		}
		v.PointKey = k
		points.SqliteDB.Unscoped().Assign(v).FirstOrCreate(&v, "input_name = ? AND point_key = ?", inputName, v.PointKey)
	}

	c.JSON(200, body)
}
func getConfig(c *gin.Context) {
	switch c.Param("pluginType") {
	case "agent":
		c.Data(200, "application/json", []byte(gjson.GetBytes(configs.CurrentConfig, "agent").Raw))
	case "inputs", "outputs", "controllers":
		idL := strings.Split(c.Param("id"), "/")
		id := idL[len(idL)-1]

		k := fmt.Sprintf(`%s.#[id=="%s"]`, c.Param("pluginType"), id)
		c.Data(200, "application/json", []byte(gjson.GetBytes(configs.CurrentConfig, k).Raw))
	default:
		c.Error(fmt.Errorf("can't find config by params: %v", c.Params))
	}
}
func deleteConfig(c *gin.Context) {
	idL := strings.Split(c.Param("id"), "/")
	id := idL[len(idL)-1]
	switch c.Param("pluginType") {
	case "inputs":
		for index, iC := range configs.MemoryConfig.Inputs {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				points.SqliteDB.Where("input_name = ?", iC["name_override"]).Delete(points.PointDefine{})
				configs.MemoryConfig.Inputs = append(configs.MemoryConfig.Inputs[:index], configs.MemoryConfig.Inputs[index+1:]...)
				c.JSON(200, configs.MemoryConfig.Inputs)
				return
			}
		}
		c.Error(fmt.Errorf("can't find [%s] config by id: %s", c.Param("pluginType"), id))
	case "outputs":
		for index, iC := range configs.MemoryConfig.Outputs {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				configs.MemoryConfig.Outputs = append(configs.MemoryConfig.Outputs[:index], configs.MemoryConfig.Outputs[index+1:]...)
				c.JSON(200, configs.MemoryConfig.Outputs)
				return
			}
		}
		c.Error(fmt.Errorf("can't find [%s] config by id: %s", c.Param("pluginType"), id))
	case "controllers":
		for index, iC := range configs.MemoryConfig.Controllers {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				configs.MemoryConfig.Controllers = append(configs.MemoryConfig.Controllers[:index], configs.MemoryConfig.Controllers[index+1:]...)
				c.JSON(200, configs.MemoryConfig.Controllers)
				return
			}
		}
		c.Error(fmt.Errorf("can't find [%s] config by id: %s", c.Param("pluginType"), id))
	default:
		c.Error(fmt.Errorf("invalid pluginType: %v", c.Param("pluginType")))
	}
}
func putConfig(c *gin.Context) {
	pluginType := c.Param("pluginType")
	if pluginType == "agent" {
		b, e := c.GetRawData()
		if e != nil {
			c.Error(e)
			return
		}
		if e := json.Unmarshal(b, configs.MemoryConfig.Agent); e != nil {
			c.Error(fmt.Errorf("update agent config failed: %v", e))
		} else {
			c.JSON(200, configs.MemoryConfig.Agent)
		}
		return
	}

	idL := strings.Split(c.Param("id"), "/")
	id := idL[len(idL)-1]
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		c.Error(err)
		return
	}

	switch pluginType {
	case "inputs":
		for index, iC := range configs.MemoryConfig.Inputs {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				pluginName, ok := iC["plugin_name"].(string)
				if !ok {
					c.Error(fmt.Errorf("unenabled plugin_name: %s", pluginName))
					return
				}
				if _, ok := configs.InputSample[pluginName]; !ok {
					c.Error(fmt.Errorf("unsupported plugin_name: %s", pluginName))
					return
				}
				for k, v := range body {
					switch k {
					case "point_map_path", "point_map_content":
					case "name_override":
						err := points.SqliteDB.Model(&points.PointDefine{}).Where("input_name = ?", configs.MemoryConfig.Inputs[index]["name_override"]).Updates(map[string]interface{}{"input_name": v}).Error
						if err != nil {
							log.Error().Str("input_name", v.(string)).Err(err).Msg("update name_override failed")
						}
						configs.MemoryConfig.Inputs[index][k] = v
					default:
						if _, ok := configs.InputSample[pluginName][k]; !ok {
							if _, ok := configs.InputSample["_base"][k]; !ok {
								continue
							}
						}
						configs.MemoryConfig.Inputs[index][k] = v
					}
				}

				c.JSON(200, configs.MemoryConfig.Inputs[index])
				return
			}
		}
	case "outputs":
		for index, iC := range configs.MemoryConfig.Outputs {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				pluginName, ok := iC["plugin_name"].(string)
				if !ok {
					c.Error(fmt.Errorf("unenabled plugin_name: %s", pluginName))
					return
				}
				if _, ok := configs.OutputSample[pluginName]; !ok {
					c.Error(fmt.Errorf("unsupported plugin_name: %s", pluginName))
					return
				}

				for k, v := range body {
					if _, ok := configs.OutputSample[pluginName][k]; !ok {
						if _, ok := configs.OutputSample["_base"][k]; !ok {
							continue
						}
					}
					configs.MemoryConfig.Outputs[index][k] = v
				}

				c.JSON(200, configs.MemoryConfig.Outputs[index])
				return
			}
		}
	case "controllers":
		for index, iC := range configs.MemoryConfig.Controllers {
			if idI, ok := iC["id"]; ok && idI.(string) == id {
				pluginName, ok := iC["plugin_name"].(string)
				if !ok {
					c.Error(fmt.Errorf("unenabled plugin_name: %s", pluginName))
					return
				}
				if _, ok := configs.ControllerSample[pluginName]; !ok {
					c.Error(fmt.Errorf("unsupported plugin_name: %s", pluginName))
					return
				}

				for k, v := range body {
					if _, ok := configs.ControllerSample[pluginName][k]; !ok {
						if _, ok := configs.ControllerSample["_base"][k]; !ok {
							continue
						}
					}
					configs.MemoryConfig.Controllers[index][k] = v
				}

				c.JSON(200, configs.MemoryConfig.Controllers[index])
				return
			}
		}
	}
	c.Error(fmt.Errorf("can't find config by params: %v", c.Params))
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

	switch pluginType {
	case "inputs":
		for _, v := range configs.MemoryConfig.Inputs {
			if v["name_override"].(string) == s["name_override"].(string) {
				c.Error(fmt.Errorf("input alias: %s already exists", v["name_override"]))
				return
			}
		}
		configs.MemoryConfig.Inputs = append(configs.MemoryConfig.Inputs, s)
	case "outputs":
		configs.MemoryConfig.Outputs = append(configs.MemoryConfig.Outputs, s)
	case "controllers":
		configs.MemoryConfig.Controllers = append(configs.MemoryConfig.Controllers, s)
	default:
		c.Error(fmt.Errorf("unknown pluginType: %s", pluginType))
		return
	}

	c.JSON(200, s)
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
			ExpiresAt: time.Now().Add(time.Minute * 500).Unix(),
		},
	}
	token, e := NewJWT().CreateToken(&claims)
	if e != nil {
		c.Error(e)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "expire": claims.ExpiresAt})
}
func reset(c *gin.Context) {
	var body struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(err)
		return
	}
	if configs.MemoryConfig.User[body.Username]["password"] == body.OldPassword {
		configs.MemoryConfig.User[body.Username]["password"] = body.NewPassword
		c.JSON(200, gin.H{"msg": fmt.Sprintf("reset password of: %s success", body.Username)})
	} else {
		c.Error(fmt.Errorf("reset password of: %s failed", body.Username))
	}
}
func refresh(c *gin.Context) {
	cS, ok := c.Get("claims")
	if !ok {
		c.AbortWithStatusJSON(403, gin.H{"error": "refresh failed, please re-login."})
		return
	}

	if claims, ok := cS.(*CustomClaims); ok {
		claims.StandardClaims.ExpiresAt = time.Now().Add(time.Minute * 500).Unix()
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

		if c.Request.Method != "GET" && (strings.HasPrefix(c.Request.URL.Path, "/plugin") || strings.HasPrefix(c.Request.URL.Path, "/pointMap")) {
			go func() {
				configs.FlushMemoryConfig()
				agent.ReloadSignal <- struct{}{}
			}()
		}
	})

	sFs, e := fs.New()
	if e != nil {
		log.Fatal().Err(e)
	}

	router.GET("/", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "index.html")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "text/html; charset=UTF-8", b)
	})
	router.GET("/inputs", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "inputs/index.html")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})
	router.GET("/outputs", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "outputs/index.html")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Header("Content-type", "text/html; charset=UTF-8")
		ctx.String(200, string(b))
	})
	router.GET("/login", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "login/index.html")
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
		ctx.Data(200, "application/javascript", b)
	})
	router.GET("/image/*filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/image", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "image/png; charset=UTF-8", b)
	})

	auth := router.Group("/auth")
	auth.POST("/login", login)
	auth.POST("/reset", JWTAuthMiddleware, reset)
	auth.POST("/refresh", JWTAuthMiddleware, refresh)

	router.GET("/getConfigSample", JWTAuthMiddleware, getConfigSample)
	router.GET("/getCurrentConfig", JWTAuthMiddleware, getCurrentConfig)
	pG := router.Group("/plugin/:pluginType", JWTAuthMiddleware)
	pG.GET("/*id", getConfig)
	pG.PUT("/*id", putConfig)
	pG.DELETE("/*id", deleteConfig)
	pG.POST("/", postConfig)

	pM := router.Group("/pointMap/", JWTAuthMiddleware)
	pM.GET("/:id", getPointMap)
	pM.PUT("/:id", putPointMap)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
