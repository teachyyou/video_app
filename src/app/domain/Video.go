package domain

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrAlreadyArchived = errors.New("video is already archived")
	ErrIncorrectUuid   = errors.New("incorrect uuid format")
	ErrVideoNotFound   = errors.New("video is not found")
)

const (
	DefaultLimit = 50
	MaxLimit     = 200
)

type Video struct {
	ID        string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Filename  string
	Slug      string
	SizeBytes int64
	DurationS sql.NullInt32
	CreatedAt time.Time `gorm:"not null;default:now()"`

	Status              string `gorm:"type:text;not null;default:uploaded"`
	RetryAttempt        int    `gorm:"not null;default:0"`
	FailureReason       *string
	ProcessingStartedAt *time.Time
	HLSReadyAt          *time.Time

	ArchivedAt *time.Time
}

type Pagination struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}

type ListPayload[T any] struct {
	Data       []T   `json:"data"`
	TotalCount int64 `json:"totalCount"`
}

type VideoStatus string

const (
	StatusUploaded    VideoStatus = "uploaded"
	StatusProcessing  VideoStatus = "processing"
	StatusComplete    VideoStatus = "complete"
	StatusInterrupted VideoStatus = "interrupted"
	StatusArchived    VideoStatus = "archived"
)

type ListFilter string

const (
	FilterAll      ListFilter = "all"
	FilterActive   ListFilter = "active"   // archived_at IS NULL
	FilterArchived ListFilter = "archived" // archived_at IS NOT NULL
)

func (p *Pagination) Normalize() {
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
}

func ParseListFilter(s string) (ListFilter, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "all":
		return FilterAll, true
	case "active":
		return FilterActive, true
	case "archived":
		return FilterArchived, true
	default:
		return FilterAll, false
	}
}
