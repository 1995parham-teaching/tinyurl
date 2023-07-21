package model

import "time"

type URL struct {
	Key    string
	URL    string
	Expire *time.Time
}
