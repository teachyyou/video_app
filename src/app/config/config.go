package config

import (
	"fmt"
	"os"
	"strconv"

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
	SlugLength int
	DataDir    string
	ArchiveDir string
	RawDir     string
	TmpDir     string
	ConvDir    string
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
			SlugLength: getEnvAsInt("SLUG_LENGTH", 12),
			DataDir:    getEnv("DATA_DIR", "/data"),
			ArchiveDir: getEnv("ARCHIVE_DIR", "/data/archive"),
			RawDir:     getEnv("RAW_DIR", "/data/raw"),
			TmpDir:     getEnv("TMP_DIR", "/data/tmp/work"),
			ConvDir:    getEnv("CONV_DIR", "/data/converted"),
		},
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getEnvAsInt(key string, def int) int {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return def
}

var Module = fx.Module("config",
	fx.Provide(
		Load,
	),
)
