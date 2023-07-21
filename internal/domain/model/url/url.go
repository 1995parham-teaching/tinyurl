package url

import "time"

type URL struct {
	Key    string
	URL    string
	Expire *time.Time
}
