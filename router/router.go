package router

import (
	"device_adaptor/agent"
	"device_adaptor/alarm"
	"device_adaptor/configs"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/controllers"
	"device_adaptor/plugins/inputs"
	"device_adaptor/plugins/outputs"
	_ "device_adaptor/statik"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/json-iterator/go"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/tidwall/gjson"
	"gopkg.in/olahol/melody.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

type Login struct {
	Username string `form:"username" json:"username" xml:"username"  binding:"required"`
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
}

func getStatic(fs http.FileSystem, p string) ([]byte, error) {
	pp := path.Join("/", p)
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
		_ = c.Error(fmt.Errorf("can't find input plugin by id: %s", id))
		return
	}

	pointArray := make([]points.PointDefine, 0)
	pointMap := make(map[string]points.PointDefine)
	points.SqliteDB.Where("input_name = ?", iC["name_override"]).Find(&pointArray)
	for _, p := range pointArray {
		pointMap[p.PointKey] = p
	}

	c.JSON(200, gin.H{
		"point_map_content": pointMap,
		"point_map_path":    "", //TODO  path
	})
}
func probePointMap(c *gin.Context) {
	c.JSON(200, gin.H{

	})
}
func putPointMap(c *gin.Context) {
	id := c.Param("id")
	iC, ok := configs.GetInputConfigById(id)
	if !ok {
		_ = c.Error(fmt.Errorf("can't find input plugin by id: %s", id))
		return
	}

	inputName := iC["name_override"].(string)

	var body struct {
		PointMapPath    string `json:"point_map_path"`
		PointMapContent string `json:"point_map_content"`
	}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}

	pointMap := make(map[string]points.PointDefine)
	if e := yaml.UnmarshalStrict([]byte(body.PointMapContent), &pointMap); e != nil {
		if e = jsoniter.Unmarshal([]byte(body.PointMapContent), &pointMap); e != nil {
			_ = c.Error(fmt.Errorf("point_map_content is neither json nor yaml format: %v", e))
			return
		}
	}
	pointMapKeys := make([]string, len(pointMap))

	i := 0
	for _, v := range pointMap {
		pointMapKeys[i] = v.PointKey
		i += 1
	}

	timeS := time.Now()
	begin := points.SqliteDB.Begin()
	if r := begin.Where("input_name = ?", inputName).Delete(points.PointDefine{}); r.Error != nil {
		r.Rollback()
		log.Error().Err(r.Error).Msg("Delete")
		_ = c.Error(r.Error)
		return
	}

	for k, v := range pointMap {
		v.InputName = inputName
		if v.Name == "" {
			v.Name = k
		}
		v.PointKey = k
		if r := begin.Assign(v).FirstOrCreate(&v, "input_name = ? AND point_key = ?", inputName, v.PointKey); r.Error != nil {
			r.Rollback()
			log.Error().Err(r.Error).Msg("FirstOrCreate")
			c.Error(r.Error)
			return
		}
	}

	if r := begin.Commit(); r.Error != nil {
		r.Rollback()
		log.Error().Err(r.Error).Msg("UpdatePointMap")
		c.Error(r.Error)
		return
	}
	log.Info().Str("TimeSince", time.Since(timeS).String()).Msg("UpdatePointMap")

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
		if e := jsoniter.Unmarshal(b, configs.MemoryConfig.Agent); e != nil {
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

func getStatInfo() map[string]interface{} {
	n, _ := host.Info()
	v, _ := mem.VirtualMemory()
	d, _ := disk.Usage("/")
	bootTime, _ := host.BootTime()
	cc, _ := cpu.Percent(time.Second, false)

	nv, _ := net.IOCounters(true)
	network := make(map[string]interface{}, 0)
	for _, nC := range nv {
		if nC.Name == "en0" {
			network["send"] = fmt.Sprintf("%d MB", nv[0].BytesSent/1024/1024)
			network["recv"] = fmt.Sprintf("%d MB", nv[0].BytesRecv/1024/1024)
		}
	}

	return map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"cpu":       fmt.Sprintf("%.2f%%", cc[0]),
		"disk":      fmt.Sprintf("%.2f%%", d.UsedPercent),
		"boot_time": time.Duration(1e9 * (time.Now().Unix() - int64(bootTime))).String(),
		"os":        fmt.Sprintf("%v(%v) %v", n.Platform, n.PlatformFamily, n.PlatformVersion),
		"network":   network,
		"memory": map[string]interface{}{
			"used":         fmt.Sprintf("%d MB", v.Used/1024/1024),
			"used_percent": fmt.Sprintf("%.2f%%", v.UsedPercent),
		},
	}
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

		if c.Request.Method != "GET" && (strings.HasPrefix(c.Request.URL.Path, "/interface/plugin") ||
			strings.HasPrefix(c.Request.URL.Path, "/interface/pointMap")) {
			go func() {
				configs.FlushMemoryConfig()
				agent.Signal <- agent.ReloadSignal{}
			}()
		}
	})

	sFs, e := fs.New()
	if e != nil {
		log.Fatal().Err(e)
	}

	//real time websocket interface for frontend
	m := melody.New()
	counter := 0
	lock := new(sync.Mutex)
	sessionMap := make(map[string]map[*melody.Session]int)

	//stat, alarm and real time info
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for {
			select {
			case <-ticker.C:
				b, _ := jsoniter.Marshal(getStatInfo())
				sL := make([]*melody.Session, 0)
				for s := range sessionMap["/interface/stat"] {
					sL = append(sL, s)
				}
				m.BroadcastMultiple(b, sL)
			case a := <-alarm.ChanAlarm:
				sL := make([]*melody.Session, 0)
				for s := range sessionMap["/interface/alarm"] {
					sL = append(sL, s)
				}
				b, _ := jsoniter.Marshal(a)
				m.BroadcastMultiple(b, sL)
			case c := <-alarm.ChanRealTime:
				sL := make([]*melody.Session, 0)
				for s := range sessionMap[strings.Join([]string{"/interface/real_time", c.PluginName}, "/")] {
					sL = append(sL, s)
				}
				b, _ := jsoniter.Marshal(c.Metric)
				m.BroadcastMultiple(b, sL)
			}
		}
	}()

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		pathSplit := strings.Split(s.Request.URL.Path, "/")
		log.Debug().Interface("path", pathSplit).Msg("HandleMessage")
		switch pathSplit[2] {
		case "ws":
			m.BroadcastMultiple(msg, []*melody.Session{s})
		}
	})

	m.HandleConnect(func(s *melody.Session) {
		key := s.Request.URL.Path

		lock.Lock()
		defer lock.Unlock()
		if _, ok := sessionMap[key]; !ok {
			sessionMap[key] = make(map[*melody.Session]int)
		}
		sessionMap[key][s] = counter
		counter += 1
		log.Info().Interface("path", s.Request.URL.Path).Msg("WebSocket client connect")
	})
	m.HandleDisconnect(func(s *melody.Session) {
		key := s.Request.URL.Path

		lock.Lock()
		defer lock.Unlock()
		if _, ok := sessionMap[key]; ok {
			delete(sessionMap[key], s)
		}
		log.Info().Interface("path", s.Request.URL.Path).Msg("WebSocket client disconnect")
	})

	//static resources for configure page
	router.GET("/", func(ctx *gin.Context) {
		b, e := getStatic(sFs, "index.html")
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "text/html; charset=UTF-8", b)
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
	router.GET("/css/:filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/css", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "text/css", b)
	})
	router.GET("/js/:filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/js", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "application/javascript", b)
	})
	router.GET("/img/:filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/img", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "image/png; charset=UTF-8", b)
	})
	router.GET("/fonts/:filename", func(ctx *gin.Context) {
		b, e := getStatic(sFs, path.Join("/fonts", ctx.Param("filename")))
		if e != nil {
			ctx.Error(e)
			return
		}
		ctx.Data(200, "application/font-woff", b)
	})

	api := router.Group("/interface")

	api.GET("/ws", func(ctx *gin.Context) { m.HandleRequest(ctx.Writer, ctx.Request) })
	api.GET("/stat", func(ctx *gin.Context) { m.HandleRequest(ctx.Writer, ctx.Request) })
	api.GET("/alarm", func(ctx *gin.Context) { m.HandleRequest(ctx.Writer, ctx.Request) })
	api.GET("/real_time/:input", func(ctx *gin.Context) { m.HandleRequest(ctx.Writer, ctx.Request) })

	api.GET("/getConfigSample", JWTAuthMiddleware, getConfigSample)
	api.GET("/getCurrentConfig", JWTAuthMiddleware, getCurrentConfig)

	auth := api.Group("/auth")
	auth.POST("/login", login)
	auth.POST("/reset", JWTAuthMiddleware, reset)
	auth.POST("/refresh", JWTAuthMiddleware, refresh)

	pG := api.Group("/plugin/:pluginType", JWTAuthMiddleware)
	pG.GET("/*id", getConfig)
	pG.PUT("/*id", putConfig)
	pG.DELETE("/*id", deleteConfig)
	pG.POST("/", postConfig)

	pM := api.Group("/pointMap/", JWTAuthMiddleware)
	pM.GET("/:id", getPointMap)
	pM.PUT("/:id", putPointMap)

	//api.GET("/probePointMap/:id", JWTAuthMiddleware, probePointMap)

	if debug {
		gin.SetMode(gin.DebugMode)
	}
	return router
}
