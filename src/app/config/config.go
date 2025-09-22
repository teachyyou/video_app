package config

import (
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"go.uber.org/fx"
)

type DBConfig struct {
	Port     string
	User     string
	Password string
	Host     string
	Name     string
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		db.User, db.Password, db.Host, db.Port, db.Name,
	)
}

type DataConfig struct {
	DataDir string
}

type Config struct {
	DB   DBConfig
	Data DataConfig
}

func Load() *Config {
	return &Config{
		DB: DBConfig{
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Name:     getEnv("DB_NAME", "name"),
		},
		Data: DataConfig{
			DataDir: getEnv("DATA_DIR", "/test"),
		},
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var Module = fx.Module("config",
	fx.Provide(
		Load,
	),
)
