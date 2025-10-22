package cli

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/chainlink"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/eth"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/tokens"
	"github.com/dayanaadylkhanova/eth2usd/internal/app"
	"github.com/dayanaadylkhanova/eth2usd/internal/service"
	"github.com/dayanaadylkhanova/eth2usd/pkg/logger"
)

type CLIRunner struct {
	log *logger.Logger
}

func NewCLIRunner(log *logger.Logger) *CLIRunner { return &CLIRunner{log: log} }

func (r *CLIRunner) Run(ctx context.Context, cfg app.RunConfig) error {
	// deps
	ethc, err := eth.NewClient(ctx, cfg.RPCURL)
	if err != nil {
		return err
	}
	defer ethc.Close()

	feed, err := chainlink.NewFeedRegistry(cfg.ChainlinkRegistry)
	if err != nil {
		return err
	}

	// tokens
	toks, err := tokens.Load(cfg.TokensFile)
	if err != nil {
		return err
	}
	if len(toks) == 0 {
		return fmt.Errorf("no tokens to process")
	}

	valuator := service.NewValuator(r.log, ethc, feed)

	// evaluate
	res := service.ValuationResult{Rows: make([]service.ValuationRow, 0, len(toks))}
	for _, t := range toks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		row, err := valuator.ValueOne(ctx, cfg.Account, t)
		if err != nil {
			r.log.Errorf("token %s: %v", t.Symbol, err)
			// keep going, add an error row
			res.Rows = append(res.Rows, service.ValuationRow{
				Symbol: t.Symbol,
				Amount: "0",
				USD:    "0",
				Source: "error",
				Err:    err.Error(),
			})
			continue
		}
		res.Rows = append(res.Rows, row)
	}

	// totals
	var totalUSD = new(big.Rat)
	for _, row := range res.Rows {
		if row.Err != "" {
			continue
		}
		v, ok := new(big.Rat).SetString(row.USD)
		if ok {
			totalUSD.Add(totalUSD, v)
		}
	}
	res.TotalUSD = service.FormatRat(totalUSD, 2)

	// output
	var out string
	switch cfg.Format {
	case "json":
		out, err = service.FormatJSON(res)
	default:
		out, err = service.FormatText(res)
	}
	if err != nil {
		return err
	}

	if cfg.Output == "" {
		fmt.Println(out)
		return nil
	}
	return os.WriteFile(cfg.Output, []byte(out), 0o644)
}

// _ = time to avoid unused import on some toolchains
var _ = time.Second
