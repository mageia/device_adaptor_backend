package http

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/plugins/controllers"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
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
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		c.AbortWithStatusJSON(400, c.Errors.JSON())
	})
	router.POST("/set/:deviceName", h.cmdHandler)

	if h.Address == "" {
		h.Address = ":8080"
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

	var bodyIn struct {
		Key   string      `json:"key" binding:"required"`
		Value interface{} `json:"value" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&bodyIn); err != nil {
		ctx.Error(errors.New("key and value parameters are all required"))
		return
	}
	cmdId := uuid.New().String()

	success, err := input.Set(cmdId, bodyIn.Key, bodyIn.Value)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !success {
		ctx.JSON(400, gin.H{"msg": "failed", "cmdId": cmdId})
		return
	}

	ctx.JSON(200, gin.H{"msg": "success", "cmdId": cmdId})
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			Inputs: make(map[string]deviceAgent.ControllerInput),
		}
	})
}
