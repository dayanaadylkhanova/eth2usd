package app

import (
	"context"

	"github.com/dayanaadylkhanova/eth2usd/pkg/logger"
)

type App struct {
	log    *logger.Logger
	runner Runner
}

func New(log *logger.Logger, runner Runner) *App {
	return &App{log: log, runner: runner}
}

func (a *App) Start(ctx context.Context, cfg RunConfig) error {
	a.log.Infof("starting rpc=%s acct=%s fmt=%s", cfg.RPCURL, cfg.Account, cfg.Format)
	defer a.log.Infof("stopped")
	return a.runner.Run(ctx, cfg)
}
