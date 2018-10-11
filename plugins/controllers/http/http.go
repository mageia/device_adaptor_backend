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

type aboutCmd struct {
	cmdId       string
	input       deviceAgent.ControllerInput
	key         string
	value       interface{}
	success     bool
	msg         string
	callbackUrl string
}

type HTTP struct {
	Address  string
	Server   *http.Server
	Inputs   map[string]deviceAgent.ControllerInput
	cmdEnd   chan aboutCmd
	cmdParam chan aboutCmd
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

	//接收控制命令参数并执行
	go func() {
		for c := range h.cmdParam {
			log.Printf("Start to process: cmd: %v\n", c)
			err := c.input.Set(c.cmdId, c.key, c.value)
			if err != nil {
				h.cmdEnd <- aboutCmd{cmdId: c.cmdId, success: false, msg: err.Error(), callbackUrl: c.callbackUrl}
			} else {
				h.cmdEnd <- aboutCmd{cmdId: c.cmdId, success: true, msg: "success", callbackUrl: c.callbackUrl}
			}
		}
	}()

	//接收命令执行结果，判断cmdId并执行回调
	go func() {
		for c := range h.cmdEnd {
			log.Println(c.callbackUrl, c.cmdId, c.success, c.msg)
			//TODO
		}
	}()

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
		Key         string      `json:"key" binding:"required"`
		Value       interface{} `json:"value" binding:"required"`
		CallbackUrl string      `json:"callback_url" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&bodyIn); err != nil {
		ctx.Error(errors.New("key and value parameters are all required"))
		return
	}
	cmdId := uuid.New().String()

	h.cmdParam <- aboutCmd{input: input, cmdId: cmdId, key: bodyIn.Key, value: bodyIn.Value, callbackUrl: bodyIn.CallbackUrl}

	ctx.JSON(200, gin.H{"msg": "success", "cmdId": cmdId})
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			cmdParam: make(chan aboutCmd, 100),
			cmdEnd:   make(chan aboutCmd, 100),
			Inputs:   make(map[string]deviceAgent.ControllerInput),
		}
	})
}
