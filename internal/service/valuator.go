package service

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/chainlink"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/eth"
	"github.com/dayanaadylkhanova/eth2usd/internal/adapter/tokens"
	"github.com/dayanaadylkhanova/eth2usd/pkg/logger"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

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
	Source string // "chainlink"
	Err    string // optional error
}

type ValuationResult struct {
	Rows     []ValuationRow
	TotalUSD string
}

// ValueOne reads on-chain balance, fetches USD price, returns formatted row.
func (v *Valuator) ValueOne(ctx context.Context, account string, t tokens.Token) (ValuationRow, error) {
	var acc common.Address
	if !common.IsHexAddress(account) {
		return ValuationRow{}, errors.New("invalid --account address")
	}
	acc = common.HexToAddress(account)

	var raw *big.Int
	var sym string
	var decimals uint8

	if t.Address == chainlink.ETHPseudoAddress {
		bal, err := v.eth.GetBalance(ctx, acc)
		if err != nil {
			return ValuationRow{}, err
		}
		raw = bal
		sym = "ETH"
		decimals = 18
	} else {
		if !common.IsHexAddress(t.Address) {
			return ValuationRow{}, errors.New("token address is not hex: " + t.Address)
		}
		addr := common.HexToAddress(t.Address)
		// balance
		bal, err := v.eth.ERC20BalanceOf(ctx, addr, acc)
		if err != nil {
			return ValuationRow{}, err
		}
		raw = bal
		// decimals
		dec, err := v.eth.ERC20Decimals(ctx, addr)
		if err != nil {
			return ValuationRow{}, err
		}
		decimals = dec
		// symbol
		if t.Symbol != "" {
			sym = t.Symbol
		} else if s, err := v.eth.ERC20Symbol(ctx, addr); err == nil && s != "" {
			sym = s
		} else {
			sym = "TKN"
		}
	}

	// price via feed registry
	var base common.Address
	if t.Address == chainlink.ETHPseudoAddress {
		// Chainlink uses 0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE as native?
		// However in Feed Registry, base is token address (ETH = 0xEeeee... on mainnet).
		base = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
	} else {
		base = common.HexToAddress(t.Address)
	}
	quote := common.HexToAddress("0x0000000000000000000000000000000000000348") // USD = 0x0348 per registry docs

	// decimals(base,quote)
	decData, err := v.feed.PackDecimals(base, quote)
	if err != nil {
		return ValuationRow{}, err
	}
	decOut, err := v.eth.Eth.CallContract(ctx, ethereum.CallMsg{To: addrPtr(v.feed.Address()), Data: decData}, nil)
	if err != nil {
		return ValuationRow{}, err
	}
	priceDecimals, err := v.feed.UnpackDecimals(decOut)
	if err != nil {
		return ValuationRow{}, err
	}

	// latestRoundData(base,quote)
	ld, err := v.feed.PackLatestRoundData(base, quote)
	if err != nil {
		return ValuationRow{}, err
	}
	ldOut, err := v.eth.Eth.CallContract(ctx, ethereum.CallMsg{To: addrPtr(v.feed.Address()), Data: ld}, nil)
	if err != nil {
		return ValuationRow{}, err
	}
	answer, updatedAt, err := v.feed.DecodeLatestRoundData(ldOut)
	if err != nil {
		return ValuationRow{}, err
	}
	stale := time.Since(updatedAt) > 24*time.Hour // simple staleness check

	// compute: amountHuman * priceHuman
	amountHuman := FormatAmount(raw, int(decimals), 6)
	priceHuman := FormatAmount(answer, int(priceDecimals), 8)

	usd := MulDecimalStrings(amountHuman, priceHuman, 2)

	source := "chainlink"
	if stale {
		source = "chainlink:stale"
	}

	return ValuationRow{
		Symbol: sym,
		Amount: amountHuman,
		USD:    usd,
		Source: source,
	}, nil
}

func addrPtr(a common.Address) *common.Address { return &a }

// stringsNewReader is a tiny helper to avoid importing bytes in multiple files.
func stringsNewReader(s string) *stringsReader { return &stringsReader{s: s} }

type stringsReader struct {
	s string
	i int
}

func (r *stringsReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, ioEOF{}
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

type ioEOF struct{}

func (ioEOF) Error() string   { return "EOF" }
func (ioEOF) Timeout() bool   { return false }
func (ioEOF) Temporary() bool { return false }
