package config

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

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

func New(params Params) (*sql.DB, error) {

	db, err := sql.Open("pgx", params.Cfg.DB.DSN())

	if err != nil {
		params.Logger.Info(err.Error())
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		params.Logger.Error("db ping failed", zap.Error(err))
		_ = db.Close()
		return nil, err
	}

	params.Logger.Info("db connected")
	return db, nil
}

var DbModule = fx.Module("db", fx.Provide(New), fx.Invoke(RunMigrations))
