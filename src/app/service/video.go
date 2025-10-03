package service

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/domain"
	"awesomeProject/src/app/repository"
	"awesomeProject/src/util"
	"context"
	"database/sql"
	"io"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type VideoService struct {
	Config     *config.Config
	Repository *repository.VideoRepository
	HlsService *ConversionService
	log        *zap.Logger
}

func NewVideoService(repo *repository.VideoRepository, config *config.Config, convService *ConversionService, log *zap.Logger) *VideoService {
	return &VideoService{
		Config:     config,
		Repository: repo,
		HlsService: convService,
		log:        log,
	}
}

func (service *VideoService) GetAllVideos(ctx context.Context, pagination domain.Pagination, status domain.ListFilter) (domain.ListPayload[domain.Video], error) {
	payload, err := service.Repository.GetAll(ctx, pagination, status)

	return payload, err
}

func (service *VideoService) GetVideo(ctx context.Context, id string) (*domain.Video, error) {
	video, err := service.Repository.GetById(ctx, id)

	return video, err
}

func (service *VideoService) UpdateVideoTitle(ctx context.Context, id string, title string) (*domain.Video, error) {
	video, err := service.Repository.UpdateById(ctx, id, map[string]interface{}{"filename": title + ".mp4"})
	return video, err
}

func (service *VideoService) Save(ctx context.Context, header *multipart.FileHeader) (string, error) {

	slug, err := util.RandomSlug(service.Config.Data.SlugLength)

	if err != nil {
		log.Println("error generating slug", err)
		return "", err
	}
	datePath := time.Now().Format("2006/01/02")
	dirPath := filepath.Join(service.Config.Data.RawDir, datePath, slug)

	//Нужно создать директорию под видос, если такой ещё нет
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Println("error creating directories", err)
		return "", err
	}

	//slug будет названием директории, а сам файл назовем source.mp4
	destPath := filepath.Join(dirPath, "source.mp4")

	file, err := header.Open()
	if err != nil {
		log.Println("error opening file", err)
		return "", err
	}
	defer file.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		log.Println("error creating file", err)
		return "", err
	}

	defer dest.Close()
	if _, err := io.Copy(dest, file); err != nil {
		log.Println("error writing to file", err)
		return "", err
	}

	duration, err := util.ProbeDuration(ctx, destPath)
	var durationField sql.NullInt32

	if err == nil {
		seconds := int32(math.Round(duration.Seconds()))
		durationField = sql.NullInt32{Int32: seconds, Valid: true}
	} else {
		service.log.Error("ffprobe for duration failed", zap.Error(err), zap.String("slug", slug))
		return "", err

	}

	_, err = service.Repository.Insert(ctx, &domain.Video{
		Filename:  header.Filename,
		Slug:      slug,
		SizeBytes: header.Size,
		DurationS: durationField,
	})
	if err != nil {
		return "", err
	}

	service.HlsService.Enqueue(slug)
	log.Println("enqueued worker for video", zap.String("slug", slug))

	return destPath, nil
}

func (service *VideoService) Archive(ctx context.Context, id string) error {
	video, err := service.Repository.GetById(ctx, id)

	if err != nil {
		log.Println("error getting file from db: ", err)

		return err
	}

	if video.ArchivedAt != nil {
		err := domain.ErrAlreadyArchived
		log.Println(err.Error())
		return err
	}

	if video.IsProcessing() {
		err := domain.ErrVideoIsProcessing
		log.Println(err.Error())
		return err
	}

	datePath := time.Now().Format("2006/01/02")

	//Достаем из raw/.../slug/source.mp4
	oldPath := filepath.Join(service.Config.Data.RawDir, datePath, video.Slug, "source.mp4")

	originalFile, err := os.Open(oldPath)
	if err != nil {
		log.Println("error opening the file: ", err)
		return err
	}
	defer originalFile.Close()

	//Кладем в archive/slug

	ext := strings.ToLower(filepath.Ext(video.Filename))
	newName := video.Slug + ext
	newPath := filepath.Join(service.Config.Data.ArchiveDir, newName)

	newFile, err := os.Create(newPath)
	if err != nil {
		log.Println("we here!")
		log.Println(newPath)
		log.Println("error creating new file: ", err)
		return err
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, originalFile)
	if err != nil {
		log.Println("error copying file: ", err)
		return err
	}

	originalFile.Close()

	oldDir := filepath.Join(service.Config.Data.RawDir, datePath, video.Slug)

	if err := os.RemoveAll(oldDir); err != nil {
		log.Println("error removing old dir:", err)
		return err
	}

	err = service.Repository.Archive(ctx, id)

	if err != nil {
		log.Println("error archiving the file: ", err)
		return err
	}
	return nil
}

var VideoModule = fx.Module("video-service", fx.Provide(NewVideoService))
