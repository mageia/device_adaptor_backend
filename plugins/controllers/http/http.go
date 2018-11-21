package http

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/internal/points"
	"deviceAdaptor/plugins/controllers"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Command struct {
	cmdId string
	input deviceAgent.ControllerInput
	kv    map[string]interface{}
}

type HTTP struct {
	Address string
	Server  *http.Server
	Inputs  map[string]deviceAgent.ControllerInput
	chanCmd chan *Command
}

func (h *HTTP) SyncExecute(c *Command) error {
	if e := c.input.SetValue(c.kv); e != nil {
		return e
	}

	return nil
}

func (h *HTTP) Name() string {
	return "http"
}

func (h *HTTP) RegisterInput(name string, input deviceAgent.ControllerInput) {
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
		log.Info().Msgf("Successfully connected to controller: %s, address: [%s]", h.Name(), srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Info().Msgf("E! Listen: %s failed: %s", srv.Addr, err)
		}
		log.Info().Msgf("Successfully closed controller: %s, address: [%s]", h.Name(), srv.Addr)
	}()

	go h.Stop(ctx)

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

	r := make(map[string]map[string]points.PointDefine)
	if len(getBody.Inputs) == 0 {
		for _, iV := range h.Inputs {
			r[iV.Name()] = iV.RetrievePointMap(nil)
		}
		ctx.JSON(200, r)
		return
	}

	for _, iV := range h.Inputs {
		for _, iS := range getBody.Inputs {
			if iS == iV.Name() || iS == iV.OriginName() {
				r[iV.Name()] = iV.RetrievePointMap(nil)
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
			h.chanCmd <- c
			ctx.JSON(200, gin.H{"cmd_id": c.cmdId})
			return
		}
	}

	ctx.Error(fmt.Errorf("unknown or unregistered input: %s", deviceName))
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			Inputs:  make(map[string]deviceAgent.ControllerInput),
			chanCmd: make(chan *Command, 1000),
		}
	})
}
