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
	"github.com/json-iterator/go"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

type HTTP struct {
	Address      string
	Server       *http.Server
	Inputs       map[string]deviceAgent.ControllerInput
	chanResults  chan result
	chanCommands chan command
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

	//接收控制命令参数并执行
	go func() {
		for {
			select {
			case c := <-h.chanCommands:
				log.Printf("Starting process cmd: %v\n", c)
				result := c.execute()
				h.chanResults <- result
				log.Printf("command processed: %v\n", result)
			case <-ctx.Done():
				return
			}
		}
	}()

	//接收命令执行结果，判断cmdId并执行回调
	go func() {
		for {
			select {
			case r := <-h.chanResults:
				rS, err := jsoniter.MarshalToString(r)
				if err != nil {
					log.Println(err)
					continue
				}
				if r.CallbackUrl == "" {
					continue
				}
				resp, e := http.Post(r.CallbackUrl, "application/json", strings.NewReader(rS))
				if e != nil {
					log.Println(e)
				} else {
					io.Copy(os.Stdout, resp.Body)
				}
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
		Value []string `json:"value" binding:"required"`
		//CallbackUrl string   `json:"callback_url" binding:"required"`
	}
	var setBody struct {
		Value       map[string]interface{} `json:"value" binding:"required"`
		CallbackUrl string                 `json:"callback_url" binding:"required"`
	}

	if err := ctx.ShouldBindBodyWith(&getBody, binding.JSON); cmdType == "GET" && err == nil {
		c := command{
			input:   input,
			cmdType: cmdType,
			cmdId:   cmdId,
			subCmd:  subCmd,
			value:   getBody.Value,
		}
		r := c.execute()
		ctx.JSON(200, r.Msg)
		return
	} else if err := ctx.ShouldBindBodyWith(&setBody, binding.JSON); cmdType == "SET" && err == nil {
		h.chanCommands <- command{
			input:       input,
			cmdType:     cmdType,
			cmdId:       cmdId,
			subCmd:      subCmd,
			value:       setBody.Value,
			callbackUrl: setBody.CallbackUrl,
		}

		log.Println(reflect.TypeOf(setBody.Value))
	} else {
		log.Println(err)
		// TODO 明确错误的类型
		ctx.Error(errors.New("unmatched value format"))
		return
	}
	ctx.JSON(200, gin.H{"msg": "success", "cmdId": cmdId})
}

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			chanCommands: make(chan command, 100),
			chanResults:  make(chan result, 100),
			Inputs:       make(map[string]deviceAgent.ControllerInput),
		}
	})
}
