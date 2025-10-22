package app

import "context"

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
