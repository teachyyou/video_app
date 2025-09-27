package repository

import (
	"awesomeProject/src/app/domain"
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type VideoRepository struct {
	DB     *gorm.DB
	Logger *zap.Logger
}

func NewVideoRepository(db *gorm.DB, logger *zap.Logger) *VideoRepository {
	return &VideoRepository{DB: db, Logger: logger}
}

func (repo *VideoRepository) GetAll(ctx context.Context, pagination domain.Pagination, status domain.ListFilter) (domain.ListPayload[domain.Video], error) {
	var result []domain.Video
	var total int64

	query := repo.DB.WithContext(ctx).Model(&domain.Video{})

	switch status {
	case domain.FilterActive:
		query = query.Where("archived_at IS NULL")
	case domain.FilterArchived:
		query = query.Where("archived_at IS NOT NULL")
	}

	if err := query.Count(&total).Error; err != nil {
		return domain.ListPayload[domain.Video]{}, err
	}

	query = query.Limit(int(pagination.Limit)).Offset(int(pagination.Offset))

	if err := query.Find(&result).Error; err != nil {
		return domain.ListPayload[domain.Video]{}, err
	}

	return domain.ListPayload[domain.Video]{Data: result, TotalCount: total}, nil
}

func (repo *VideoRepository) Insert(ctx context.Context, video *domain.Video) (string, error) {
	if err := repo.DB.WithContext(ctx).Create(&video).Error; err != nil {
		return "", err
	}

	return video.ID, nil
}

func (repo *VideoRepository) GetById(ctx context.Context, id string) (*domain.Video, error) {
	var video domain.Video

	if err := repo.DB.WithContext(ctx).First(&video, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVideoNotFound
		}

		return nil, err
	}

	return &video, nil
}

func (repo *VideoRepository) Archive(ctx context.Context, id string) error {
	archivedAt := time.Now().UTC()

	if err := repo.DB.WithContext(ctx).Model(&domain.Video{}).Where("id = ?", id).Updates(map[string]interface{}{"archived_at": archivedAt}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrVideoNotFound
		}
		return err
	}

	return nil
}

var VideoRepoModule = fx.Module("video-repository", fx.Provide(NewVideoRepository))
