package repository

import (
	"awesomeProject/src/app/cache"
	"awesomeProject/src/app/domain"
	"context"
	"errors"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VideoRepository struct {
	Cache  *cache.VideoCache
	DB     *gorm.DB
	Logger *zap.Logger
}

func NewVideoRepository(db *gorm.DB, logger *zap.Logger, cache *cache.VideoCache) *VideoRepository {
	return &VideoRepository{
		DB:     db,
		Logger: logger,
		Cache:  cache,
	}
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

	query = query.Limit(int(pagination.Limit)).Offset(int(pagination.Offset)).Order("created_at DESC")

	if err := query.Find(&result).Error; err != nil {
		return domain.ListPayload[domain.Video]{}, err
	}

	return domain.ListPayload[domain.Video]{Data: result, TotalCount: total}, nil
}

func (repo *VideoRepository) UpdateById(ctx context.Context, id string, updates map[string]interface{}) (*domain.Video, error) {
	var video domain.Video

	res := repo.DB.WithContext(ctx).
		Model(&video).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(updates)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrVideoNotFound
	}

	repo.Cache.Delete(id)

	return &video, nil
}

func (repo *VideoRepository) GetBySlug(ctx context.Context, slug string) (*domain.Video, error) {
	var video domain.Video

	if err := repo.DB.WithContext(ctx).First(&video, "slug = ?", slug).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVideoNotFound
		}
		return nil, err
	}
	return &video, nil
}

func (repo *VideoRepository) Insert(ctx context.Context, video *domain.Video) (string, error) {
	if err := repo.DB.WithContext(ctx).Create(&video).Error; err != nil {
		return "", err
	}

	return video.ID, nil
}

func (repo *VideoRepository) SetProcessing(ctx context.Context, id string, time time.Time) error {
	res := repo.DB.WithContext(ctx).
		Model(&domain.Video{}).
		Where("id = ?", id).
		Update("status", domain.StatusProcessing).
		Update("processing_started_at", time)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrVideoNotFound
	}
	repo.Cache.Delete(id)
	return nil
}

func (repo *VideoRepository) SetReady(ctx context.Context, id string, time time.Time) error {
	updates := map[string]any{
		"status":       string(domain.StatusComplete),
		"hls_ready_at": time,
	}

	res := repo.DB.WithContext(ctx).
		Model(&domain.Video{}).
		Where("id = ?", id).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrVideoNotFound
	}
	repo.Cache.Delete(id)
	return nil
}

func (repo *VideoRepository) SetInterrupted(ctx context.Context, id string, reason error) error {
	updates := map[string]any{
		"status":         string(domain.StatusInterrupted),
		"failure_reason": reason.Error(),
		"retry_attempt":  gorm.Expr("retry_attempt + 1"),
	}

	res := repo.DB.WithContext(ctx).
		Model(&domain.Video{}).
		Where("id = ?", id).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrVideoNotFound
	}
	repo.Cache.Delete(id)
	return nil
}

func (repo *VideoRepository) GetById(ctx context.Context, id string) (*domain.Video, error) {

	if cachedVideo, ok := repo.Cache.Get(id); ok {
		repo.Logger.Info("got video from cache", zap.String("slug", cachedVideo.Slug))
		return cachedVideo, nil
	}

	var video domain.Video
	if err := repo.DB.WithContext(ctx).First(&video, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVideoNotFound
		}
		repo.Cache.Delete(id)
		return nil, err
	}
	repo.Cache.Set(id, &video, 0)
	repo.Logger.Info("added video to cache", zap.String("slug", video.Slug))

	return &video, nil

}

func (repo *VideoRepository) Archive(ctx context.Context, id string) error {
	archivedAt := time.Now().UTC()

	if err := repo.DB.WithContext(ctx).Model(&domain.Video{}).Where("id = ?", id).Updates(map[string]interface{}{"archived_at": archivedAt, "status": string(domain.StatusArchived)}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrVideoNotFound
		}
		return err
	}
	repo.Cache.Delete(id)
	return nil
}

var VideoRepoModule = fx.Module("video-repository", fx.Provide(NewVideoRepository))
