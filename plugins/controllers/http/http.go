package http

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/plugins/controllers"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"log"
	"net/http"
	"strings"
)

type HTTP struct {
	Address string
	Server  *http.Server
	Inputs  map[string]deviceAgent.ControllerInput
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
	router.POST("/:cmdType/:deviceName/:subCmd", h.cmdHandler)

	if h.Address == "" {
		h.Address = ":9999"
	}
	srv := &http.Server{
		Addr:    h.Address,
		Handler: router,
	}
	h.Server = srv

	go func() {
		log.Printf("D! Start CmdServer on: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("E! Listen: %s failed: %s\n", srv.Addr, err)
		}
		log.Printf("I! Server: %s closed", srv.Addr)
	}()

	go h.Stop(ctx)

	return nil
}

func (h *HTTP) Stop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			if err := h.Server.Shutdown(ctx); err != nil {
				log.Printf("E! CmdServer shutdown: %s", err)
				return err
			}
			return nil
		}
	}
}

func (h *HTTP) cmdHandler(ctx *gin.Context) {
	deviceName := ctx.Param("deviceName")
	input, ok := h.Inputs[deviceName]
	if !ok {
		ctx.Error(fmt.Errorf("undefined but requested input: %s", deviceName))
		return
	}
	cmdType := strings.ToUpper(ctx.Param("cmdType"))
	subCmd := ctx.Param("subCmd")
	cmdId := uuid.New().String()

	var getBody struct {
		Keys []string `json:"keys" binding:"required"`
	}
	if cmdType != "GET" {
		ctx.Error(errors.New("invalid command type"))
		return
	}

	if err := ctx.ShouldBindBodyWith(&getBody, binding.JSON); err != nil {
		ctx.Error(err)
		return
	}
	r, err := command{
		input:   input,
		cmdType: cmdType,
		cmdId:   cmdId,
		subCmd:  subCmd,
		keys:    getBody.Keys,
	}.execute()
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(200, r)
	return
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			Inputs: make(map[string]deviceAgent.ControllerInput),
		}
	})
}
