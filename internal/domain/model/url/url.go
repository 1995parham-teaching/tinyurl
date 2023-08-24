package url

import (
	"database/sql"
	"time"
)

type URL struct {
	Key    string       `gorm:"primaryKey"`
	URL    string       `gorm:"index"`
	Visits uint64       `gorm:"check:,visits >= 0"`
	Expire sql.NullTime `gorm:"check:,expire > created_at"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
