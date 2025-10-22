package service

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/chainlink"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/eth"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/tokens"
	"github.com/dayanaadylkhanova/eth2usd/pkg/logger"
)

// Valuator coordinates on-chain reads and pricing to produce valuation rows.
type Valuator struct {
	log  *logger.Logger
	eth  *eth.Client
	feed *chainlink.FeedRegistry
}

func NewValuator(log *logger.Logger, ethc *eth.Client, feed *chainlink.FeedRegistry) *Valuator {
	return &Valuator{log: log, eth: ethc, feed: feed}
}

type ValuationRow struct {
	Symbol string
	Amount string // human amount
	USD    string // human usd
	Source string // "chainlink" | "chainlink:stale" | "error"
	Err    string // optional error message for the row
}

type ValuationResult struct {
	Rows     []ValuationRow
	TotalUSD string
}

// ValueOne reads the balance for a token, fetches its USD price via Chainlink Feed Registry,
// and returns a formatted valuation row.
func (v *Valuator) ValueOne(ctx context.Context, account string, t tokens.Token) (ValuationRow, error) {
	// Validate account
	if !common.IsHexAddress(account) {
		return ValuationRow{}, errors.New("invalid --account address")
	}
	acc := common.HexToAddress(account)

	// 1) Read on-chain balance + metadata
	var (
		raw      *big.Int
		sym      string
		decimals uint8
	)

	if t.Address == chainlink.ETHPseudoAddress {
		// Native ETH
		bal, err := v.eth.GetBalance(ctx, acc)
		if err != nil {
			return ValuationRow{}, err
		}
		raw = bal
		sym = "ETH"
		decimals = 18
	} else {
		// ERC-20
		if !common.IsHexAddress(t.Address) {
			return ValuationRow{}, errors.New("token address is not hex: " + t.Address)
		}
		addr := common.HexToAddress(t.Address)

		bal, err := v.eth.ERC20BalanceOf(ctx, addr, acc)
		if err != nil {
			return ValuationRow{}, err
		}
		raw = bal

		dec, err := v.eth.ERC20Decimals(ctx, addr)
		if err != nil {
			return ValuationRow{}, err
		}
		decimals = dec

		if t.Symbol != "" {
			sym = t.Symbol
		} else if s, err := v.eth.ERC20Symbol(ctx, addr); err == nil && s != "" {
			sym = s
		} else {
			sym = "TKN"
		}
	}

	// Pre-format human-readable amount (even if price is missing we can return this)
	amountHuman := FormatAmount(raw, int(decimals), 6)

	// 2) Price via Chainlink Feed Registry (base, quote=USD)
	var base common.Address
	if t.Address == chainlink.ETHPseudoAddress {
		// Native token placeholder used by Chainlink registry on mainnet
		base = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
	} else {
		base = common.HexToAddress(t.Address)
	}
	quote := common.HexToAddress("0x0000000000000000000000000000000000000348") // USD

	// decimals(base, quote)
	decData, err := v.feed.PackDecimals(base, quote)
	if err != nil {
		return ValuationRow{}, err
	}
	decOut, err := v.eth.Eth.CallContract(ctx, ethereum.CallMsg{To: ptr(v.feed.Address()), Data: decData}, nil)
	if err != nil {
		return ValuationRow{}, err
	}
	priceDecimals, err := v.feed.UnpackDecimals(decOut)
	if err != nil {
		return ValuationRow{}, err
	}

	// latestRoundData(base, quote)
	ld, err := v.feed.PackLatestRoundData(base, quote)
	if err != nil {
		return ValuationRow{}, err
	}
	ldOut, err := v.eth.Eth.CallContract(ctx, ethereum.CallMsg{To: ptr(v.feed.Address()), Data: ld}, nil)
	if err != nil {
		return ValuationRow{}, err
	}
	answer, updatedAt, err := v.feed.DecodeLatestRoundData(ldOut)
	if err != nil {
		return ValuationRow{}, err
	}

	// 3) Validate price and staleness
	if answer == nil || answer.Sign() <= 0 {
		// No price available
		return ValuationRow{
			Symbol: sym,
			Amount: amountHuman,
			USD:    "0",
			Source: "chainlink",
			Err:    ErrNoPrice.Error(),
		}, nil
	}

	stale := time.Since(updatedAt) > 24*time.Hour

	// 4) Compute USD = amount * price
	priceHuman := FormatAmount(answer, int(priceDecimals), 8)
	usd := MulDecimalStrings(amountHuman, priceHuman, 2)

	source := "chainlink"
	if stale {
		source = "chainlink:stale"
	}

	rowErr := ""
	if stale {
		rowErr = ErrStalePrice.Error()
	}

	return ValuationRow{
		Symbol: sym,
		Amount: amountHuman,
		USD:    usd,
		Source: source,
		Err:    rowErr,
	}, nil
}

func ptr(a common.Address) *common.Address { return &a }
