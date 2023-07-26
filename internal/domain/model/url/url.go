package url

import (
	"time"
)

type URL struct {
	Key    string `gorm:"primaryKey"`
	URL    string
	Visits uint64
	Expire *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
