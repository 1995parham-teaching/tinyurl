package url

import (
	"database/sql"
	"time"
)

type URL struct {
	Key    string `gorm:"primaryKey"`
	URL    string
	Visits uint64
	Expire sql.NullTime

	CreatedAt time.Time
	UpdatedAt time.Time
}
