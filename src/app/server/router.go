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
	MediaHandler *handler.MediaHandler
}

func NewRouter(p RouterParams) *gin.Engine {

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	api.GET("/hello", p.HelloHandler.Hello)

	api.GET("/video", p.VideoHandler.GetVideos)
	api.GET("/video/:video_uuid", p.VideoHandler.GetVideo)
	api.POST("/video", p.VideoHandler.AddVideo)
	api.PATCH("/video/:video_uuid", p.VideoHandler.UpdateVideo)
	api.DELETE("/video/:video_uuid", p.VideoHandler.ArchiveVideo)

	p.MediaHandler.Register(r)
	return r
}
