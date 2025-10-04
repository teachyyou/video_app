package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

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

type CacheConfig struct {
	DefaultExpiration time.Duration
	DefaultFrequency  time.Duration
}

type HttpConfig struct {
	PublicMediaUrl string
}

type DataConfig struct {
	SlugLength int
	DataDir    string
	ArchiveDir string
	RawDir     string
}

type ConversionConfig struct {
	TmpDir   string
	ConvDir  string
	Parallel int
}

type Config struct {
	DB    DBConfig
	Http  HttpConfig
	Data  DataConfig
	Conv  ConversionConfig
	Cache CacheConfig
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
		},
		Conv: ConversionConfig{
			TmpDir:   getEnv("TMP_DIR", "/data/tmp/work"),
			ConvDir:  getEnv("CONV_DIR", "/data/converted"),
			Parallel: getEnvAsInt("PARALLEL_MAX", 4),
		},
		Http: HttpConfig{
			PublicMediaUrl: getEnv("PUBLIC_MEDIA_URL", ""),
		},
		Cache: CacheConfig{
			DefaultExpiration: time.Duration(getEnvAsInt("DEFAULT_CACHE_EXPIRATION_SECS", 60)) * time.Second,
			DefaultFrequency:  time.Duration(getEnvAsInt("DEFAULT_CACHE_CLEAN_FREQUENCY_SECS", 60)) * time.Second,
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
