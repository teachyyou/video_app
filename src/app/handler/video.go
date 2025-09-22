package handler

import (
	"awesomeProject/src/app/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type VideoHandler struct {
	service *service.VideoService
	logger  *zap.Logger
}

func NewVideoHandler(videoService *service.VideoService, logger *zap.Logger) *VideoHandler {
	return &VideoHandler{
		service: videoService,
		logger:  logger,
	}
}

func (h *VideoHandler) AddVideo(ctx *gin.Context) {
	fileHeader, err := ctx.FormFile("video")

	if err != nil {
		ctx.JSON(400, gin.H{"message": err})
		return
	}

	path, err := h.service.Save(ctx.Request.Context(), fileHeader)

	if err != nil {
		ctx.JSON(400, gin.H{"message": err})
		return
	}

	ctx.JSON(200, gin.H{"message": fileHeader.Filename, "path": path})

}

var VideoModule = fx.Module("video-handler", fx.Provide(NewVideoHandler))
