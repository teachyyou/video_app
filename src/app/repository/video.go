package repository

import (
	"awesomeProject/src/model"
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type VideoRepository struct {
	DB     *sql.DB
	Logger *zap.Logger
}

func NewVideoRepository(db *sql.DB, logger *zap.Logger) *VideoRepository {
	return &VideoRepository{DB: db, Logger: logger}
}

func (repo *VideoRepository) Insert(ctx context.Context, video *model.Video) (string, error) {
	const q = `
        INSERT INTO videos (filename, path, size_bytes, duration_s)
        VALUES ($1, $2, $3, $4)
        RETURNING id;
    `
	var id string
	var dur sql.NullInt32
	if video.DurationS.Valid {
		dur = sql.NullInt32{Int32: video.DurationS.Int32, Valid: true}
	}
	if err := repo.DB.QueryRowContext(ctx, q, video.Filename, video.Path, video.SizeBytes, dur).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

var VideoRepoModule = fx.Module("video-repository", fx.Provide(NewVideoRepository))
