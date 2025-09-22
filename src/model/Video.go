package model

import (
	"database/sql"
	"time"
)

type Video struct {
	ID        string
	Filename  string
	Path      string
	SizeBytes int64
	DurationS sql.NullInt32
	CreatedAt time.Time
}
