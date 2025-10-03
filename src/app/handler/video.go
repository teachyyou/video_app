package handler

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/domain"
	"awesomeProject/src/app/service"
	"awesomeProject/src/util"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type VideoHandler struct {
	cfg     *config.Config
	service *service.VideoService
	logger  *zap.Logger
}

type PatchVideoRequestPayload struct {
	Filename string `json:"filename" binding:"required"`
}

func NewVideoHandler(config *config.Config, videoService *service.VideoService, logger *zap.Logger) *VideoHandler {
	return &VideoHandler{
		cfg:     config,
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

func (h *VideoHandler) UpdateVideo(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("video_uuid"))

	if err != nil {
		ctx.JSON(util.HttpResponseFromError(domain.ErrIncorrectUuid))
		return
	}

	var req PatchVideoRequestPayload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	name := strings.TrimSpace(req.Filename)
	if name == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "filename must not be empty"})
		return
	}
	if utf8.RuneCountInString(name) > 200 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "filename is too long (max 200 chars)"})
		return
	}

	video, err := h.service.GetVideo(ctx.Request.Context(), id.String())

	if video.ArchivedAt != nil || video.Status == string(domain.StatusArchived) {
		ctx.JSON(400, gin.H{"message": "video is archived"})
		return
	}

	updatedVideo, err := h.service.UpdateVideoTitle(ctx.Request.Context(), video.ID, req.Filename)

	if err != nil {
		h.logger.Error("update title failed", zap.String("id", video.ID), zap.Error(err))
		ctx.JSON(util.HttpResponseFromError(err))
		return
	}

	dto := updatedVideo.ToDto()
	datePath := video.CreatedAt.Format("2006/01/02")

	dto.ConvertedUrl = util.JoinURL(h.cfg.Http.PublicMediaUrl, datePath, video.Slug, "index.m3u8")

	ctx.JSON(200, dto)

}

func (h *VideoHandler) GetVideos(ctx *gin.Context) {
	var pgn = util.ParsePagination(ctx)

	status, ok := domain.ParseListFilter(ctx.Query("status"))

	if !ok {
		ctx.JSON(400, gin.H{"message": "incorrect status"})
		return
	}

	payloadVideos, err := h.service.GetAllVideos(ctx.Request.Context(), pgn, status)

	if err != nil {
		ctx.JSON(400, gin.H{"message": err})
		return
	}

	dtos := make([]domain.VideoDTO, 0, len(payloadVideos.Data))

	for _, video := range payloadVideos.Data {
		videoDTO := video.ToDto()
		datePath := video.CreatedAt.Format("2006/01/02")
		videoDTO.ConvertedUrl = util.JoinURL(h.cfg.Http.PublicMediaUrl, datePath, video.Slug, "index.m3u8")

		dtos = append(dtos, videoDTO)
	}

	payload := domain.ListPayload[domain.VideoDTO]{
		Data:       dtos,
		TotalCount: payloadVideos.TotalCount,
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

	dto := video.ToDto()
	datePath := video.CreatedAt.Format("2006/01/02")

	dto.ConvertedUrl = util.JoinURL(h.cfg.Http.PublicMediaUrl, datePath, video.Slug, "index.m3u8")

	ctx.JSON(200, dto)

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
