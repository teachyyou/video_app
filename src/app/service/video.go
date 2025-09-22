package service

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/repository"
	"awesomeProject/src/model"
	"context"
	"database/sql"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type VideoService struct {
	Config     *config.Config
	Repository *repository.VideoRepository
	log        *zap.Logger
}

func NewVideoService(repo *repository.VideoRepository, config *config.Config, log *zap.Logger) *VideoService {
	return &VideoService{
		Config:     config,
		Repository: repo,
		log:        log,
	}
}

func (service *VideoService) Save(ctx context.Context, header *multipart.FileHeader) (string, error) {

	path := filepath.Join(service.Config.Data.DataDir, header.Filename)

	file, err := header.Open()
	if err != nil {
		log.Println("error opening file", err)
		return "", err
	}
	defer file.Close()

	dest, err := os.Create(path)
	if err != nil {
		log.Println("error creating file", err)
		return "", err
	}

	defer dest.Close()
	if _, err := io.Copy(dest, file); err != nil {
		log.Println("error writing to file", err)
		return "", err
	}

	_, err = service.Repository.Insert(ctx, &model.Video{
		Filename:  header.Filename,
		Path:      path,
		SizeBytes: header.Size,
		DurationS: sql.NullInt32{Int32: 42, Valid: true},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}

var VideoModule = fx.Module("video-service", fx.Provide(NewVideoService))
