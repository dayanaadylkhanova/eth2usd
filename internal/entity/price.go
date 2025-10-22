package entity

import "time"

type Price struct {
	Base      string
	Quote     string // "USD"
	Value     string // decimal string
	Stale     bool
	UpdatedAt time.Time
}
