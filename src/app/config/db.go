package config

import (
	"context"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In
	Cfg    *Config
	Logger *zap.Logger
}

func RunMigrations(lc fx.Lifecycle, p Params) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Logger.Info("running migrations")

			m, err := migrate.New("file://resources/migrations", p.Cfg.DB.DSN())

			if err != nil {
				p.Logger.Fatal("failed to run migrations", zap.Error(err))
				return err

			}
			defer m.Close()

			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				return err
			}
			p.Logger.Info("migrations applied")
			return nil
		},
	})
}

func New(params Params) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(params.Cfg.DB.DSN()), &gorm.Config{})

	if err != nil {
		params.Logger.Info(err.Error())
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		params.Logger.Error("failed to get sql.DB from gorm", zap.Error(err))
		return nil, err
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		params.Logger.Error("db ping failed", zap.Error(err))
		_ = sqlDB.Close()
		return nil, err
	}

	params.Logger.Info("db connected")
	return db, nil
}

var DbModule = fx.Module("db", fx.Provide(New), fx.Invoke(RunMigrations))
