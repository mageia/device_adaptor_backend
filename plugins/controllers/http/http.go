package http

import (
	"context"
	"deviceAdaptor"
	"deviceAdaptor/plugins/controllers"
	"github.com/gin-gonic/gin"
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

func (h *HTTP) getPointMapHandler(ctx *gin.Context) {
	var getBody struct {
		Inputs []string `form:"inputs"`
	}
	if err := ctx.ShouldBindQuery(&getBody); err != nil {
		ctx.Error(err)
		return
	}

	r := make(map[string]map[string]deviceAgent.PointDefine)
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

func init() {
	controllers.Add("http", func() deviceAgent.Controller {
		return &HTTP{
			Inputs: make(map[string]deviceAgent.ControllerInput),
		}
	})
}
