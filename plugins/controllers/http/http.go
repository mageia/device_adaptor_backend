package http

import (
	"context"
	"device_adaptor"
	"device_adaptor/internal/points"
	"device_adaptor/plugins/controllers"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Command struct {
	cmdId string
	input device_agent.ControllerInput
	kv    map[string]interface{}
}

type HTTP struct {
	Address string
	Server  *http.Server
	Inputs  map[string]device_agent.ControllerInput
	ASync   bool
	chanCmd chan *Command
}

func (h *HTTP) SyncExecute(c *Command) error {
	return c.input.SetValue(c.kv)
}

func (h *HTTP) Name() string {
	return "http"
}

func (h *HTTP) RegisterInput(name string, input device_agent.ControllerInput) {
	h.Inputs[name] = input
}

func (h *HTTP) Start(ctx context.Context) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	gin.New()
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		c.AbortWithStatusJSON(400, c.Errors.JSON())
	})
	router.GET("/point_meta", h.getPointMapHandler)
	router.POST("/point_value/:deviceName", h.setPointValueHandler)

	if h.Address == "" {
		h.Address = ":9999"
	}
	srv := &http.Server{
		Addr:    h.Address,
		Handler: router,
	}
	h.Server = srv

	go func() {
		log.Info().Str("plugin", h.Name()).Str("address", srv.Addr).Msg("http controller start success")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Str("address", srv.Addr).Msg("http controller listen failed")
		}
		log.Info().Str("plugin", h.Name()).Str("address", srv.Addr).Msg("http controller close success")
	}()

	go h.Stop(ctx)

	if h.ASync {
		go func() {
			for {
				select {
				case c := <-h.chanCmd:
					h.SyncExecute(c)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	return nil
}

func (h *HTTP) Stop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if err := h.Server.Shutdown(ctx); err != nil {
				log.Info().Msgf("E! CmdServer shutdown: %s", err)
				return err
			}
			return nil
		}
	}
}

func (h *HTTP) getPointMapHandler(ctx *gin.Context) {
	var getBody struct {
		Inputs []string `form:"inputs"`
	}
	if err := ctx.ShouldBindQuery(&getBody); err != nil {
		ctx.Error(err)
		return
	}

	if len(getBody.Inputs) == 0 {
		ctx.JSON(200, nil)
		return
	}

	r := make(map[string]map[string]points.PointDefine)
	for _, iV := range h.Inputs {
		for _, iS := range getBody.Inputs {
			if iS == iV.Name() || iS == iV.OriginName() {
				//r[iV.Name()] = iV.RetrievePointMap(nil)
				pointMap := iV.RetrievePointMap(nil)
				for k, p := range pointMap {
					for kp, v := range p.Extra {
						switch eV := v.(type) {
						case string:
							eVV := make(map[string]interface{})
							if e := jsoniter.Unmarshal([]byte(eV), &eVV); e != nil {
								pointMap[kp].Extra[k] = eV
							} else {
								pointMap[kp].Extra[k] = eVV
							}
						default:
							pointMap[kp].Extra[k] = eV
						}
					}
				}
				r[iV.Name()] = pointMap
				break
			}
		}
	}
	ctx.JSON(200, r)
}

func (h *HTTP) setPointValueHandler(ctx *gin.Context) {
	var setBody map[string]interface{}
	if err := ctx.ShouldBindBodyWith(&setBody, binding.JSON); err != nil {
		ctx.Error(err)
		return
	} else if len(setBody) == 0 {
		ctx.Error(fmt.Errorf("empty key-value pairs"))
		return
	}

	deviceName := ctx.Param("deviceName")
	for _, iV := range h.Inputs {
		if deviceName == iV.Name() || deviceName == iV.OriginName() {
			c := &Command{
				input: iV,
				kv:    setBody,
				cmdId: uuid.New().String(),
			}
			if h.ASync {
				h.chanCmd <- c
				ctx.JSON(200, gin.H{"cmd_id": c.cmdId})
				return
			}

			if e := h.SyncExecute(c); e != nil {
				ctx.JSON(400, gin.H{"cmd_id": c.cmdId, "msg": e.Error()})
				return
			}
			ctx.JSON(200, gin.H{"cmd_id": c.cmdId})
			return
		}
	}

	ctx.Error(fmt.Errorf("unknown or unregistered input: %s", deviceName))
}

func init() {
	controllers.Add("http", func() device_agent.Controller {
		return &HTTP{
			Inputs:  make(map[string]device_agent.ControllerInput),
			chanCmd: make(chan *Command, 1000),
		}
	})
}
