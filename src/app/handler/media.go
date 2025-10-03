package handler

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MediaHandler struct {
	cfg     *config.Config
	service *service.VideoService
	logger  *zap.Logger
}

func NewMediaHandler(config *config.Config, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		cfg:    config,
		logger: logger,
	}
}

func (h *MediaHandler) Register(router gin.IRouter) {
	fs := gin.Dir(h.cfg.Conv.ConvDir, false)

	router.Use(func(c *gin.Context) {
		c.Header("cache-control", "no-cache")
		c.Next()
	})
	router.StaticFS(h.cfg.Http.PublicMediaUrl, fs)

}

var MediaModule = fx.Module("media-handler", fx.Provide(NewMediaHandler))
