package server

import (
	"awesomeProject/src/app/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type RouterParams struct {
	fx.In

	Logger       *zap.Logger
	HelloHandler *handler.HelloHandler
	VideoHandler *handler.VideoHandler
}

func NewRouter(p RouterParams) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	api.GET("/hello", p.HelloHandler.Hello)
	api.POST("/video", p.VideoHandler.AddVideo)

	return r
}
