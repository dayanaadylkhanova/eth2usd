package service

import "errors"

var (
	ErrStalePrice = errors.New("stale price")
	ErrNoPrice    = errors.New("no price available")
)
