package http

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/plugins/controllers"
	"deviceAdaptor/plugins/inputs"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
)

type HTTP struct {
	Address string
	Test    string
}

func (h *HTTP) Set(cmdId string, key string, value interface{}) error {
	h.Test = value.(string)
	return nil
}

func (h *HTTP) CmdHandler(ctx *gin.Context) {
	deviceName := ctx.Param("deviceName")
	creator, ok := inputs.Inputs[deviceName]
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

	input := creator()
	switch iT := input.(type) {
	case deviceAgent.ControllerInput:
		err := iT.Set(cmdId, bodyIn.Key, bodyIn.Value)
		if err != nil {
			ctx.JSON(400, gin.H{"msg": err.Error(), "cmdId": cmdId})
			return
		}
	}
	ctx.JSON(200, gin.H{"msg": "success", "cmdId": cmdId})
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
	router.POST("/set/:deviceName", h.CmdHandler)
	srv := &http.Server{
		Addr:    h.Address,
		Handler: router,
	}
	go func() {
		log.Printf("D! Start CmdServer on: %s\n", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("E! Listen: %s failed: %s\n", srv.Addr, err)
		}
		log.Printf("I! Server: %s closed", srv.Addr)
	}()

	for {
		select {
		case <-ctx.Done():
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("E! CmdServer shutdown: %s", err)
				return err
			}
			return nil
		}
	}
	return nil
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{}
	})
}
