package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/transport/cli"
	"github.com/dayanaadylkhanova/eth2usd/internal/app"
	"github.com/dayanaadylkhanova/eth2usd/pkg/logger"
)

func main() {
	var (
		rpcURL   string
		registry string
		tokens   string
		account  string
		format   string
		output   string
		timeout  time.Duration
	)
	flag.StringVar(&rpcURL, "rpc-url", "http://localhost:8545", "Ethereum JSON-RPC endpoint")
	flag.StringVar(&registry, "chainlink-registry", "", "Chainlink Feed Registry address (required)")
	flag.StringVar(&tokens, "tokens-file", "", "Path to tokens whitelist JSON (overrides defaults)")
	flag.StringVar(&account, "account", "", "Account address to read balances from (required)")
	flag.StringVar(&format, "format", "text", "Output format: text|json")
	flag.StringVar(&output, "out", "", "Output file (stdout if empty)")
	flag.DurationVar(&timeout, "timeout", 30*time.Second, "Global timeout")
	flag.Parse()

	if registry == "" || account == "" {
		log.Fatalf("--chainlink-registry and --account are required")
	}

	l := logger.New("eth2usd")
	runner := cli.NewCLIRunner(l)

	application := app.New(l, runner)

	ctx, cancel := signalContext(context.Background(), timeout)
	defer cancel()

	cfg := app.RunConfig{
		RPCURL:            rpcURL,
		ChainlinkRegistry: registry,
		TokensFile:        tokens,
		Account:           account,
		Format:            format,
		Output:            output,
	}

	if err := application.Start(ctx, cfg); err != nil {
		log.Fatalf("exit with error: %v", err)
	}
}

func signalContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ch:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
