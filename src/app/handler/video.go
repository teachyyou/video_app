package handler

import (
	"awesomeProject/src/app/domain"
	"awesomeProject/src/app/service"
	"awesomeProject/src/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *VideoHandler) GetVideos(ctx *gin.Context) {
	var pgn = util.ParsePagination(ctx)

	status, ok := domain.ParseListFilter(ctx.Query("status"))

	if !ok {
		ctx.JSON(400, gin.H{"message": "incorrect status"})
		return
	}

	payload, err := h.service.GetAllVideos(ctx.Request.Context(), pgn, status)

	if err != nil {
		ctx.JSON(400, gin.H{"message": err})
		return
	}

	ctx.JSON(200, payload)

}

func (h *VideoHandler) GetVideo(ctx *gin.Context) {

	id, err := uuid.Parse(ctx.Param("video_uuid"))

	if err != nil {
		ctx.JSON(util.HttpResponseFromError(domain.ErrIncorrectUuid))
		return
	}

	video, err := h.service.GetVideo(ctx.Request.Context(), id.String())

	if err != nil {
		h.logger.Info("error getting video by id")
		ctx.JSON(util.HttpResponseFromError(err))
		return
	}

	ctx.JSON(200, video)

}

func (h *VideoHandler) ArchiveVideo(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("video_uuid"))

	if err != nil {
		ctx.JSON(util.HttpResponseFromError(domain.ErrIncorrectUuid))
		return
	}

	if err := h.service.Archive(ctx.Request.Context(), id.String()); err != nil {
		h.logger.Info("error archiving video", zap.Error(err))
		ctx.JSON(util.HttpResponseFromError(err))
		return
	}

	ctx.JSON(204, gin.H{})

}

var VideoModule = fx.Module("video-handler", fx.Provide(NewVideoHandler))
