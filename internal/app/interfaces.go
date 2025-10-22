package app

import "context"

//go:generate mockgen -source=interfaces.go -destination=./interfaces_mock.go -package=app

type Runner interface {
	Run(ctx context.Context, cfg RunConfig) error
}

type RunConfig struct {
	RPCURL            string
	ChainlinkRegistry string
	TokensFile        string
	Account           string
	Format            string // "text" or "json"
	Output            string // file path or "" for stdout
}
