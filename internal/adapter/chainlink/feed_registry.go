package chainlink

import (
	"embed"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Встраиваем ABI Feed Registry.
//
//go:embed abi/feed_registry.json
var feedRegistryFS embed.FS

type FeedRegistry struct {
	addr common.Address
	abi  abi.ABI
}

func NewFeedRegistry(addr string) (*FeedRegistry, error) {
	if !common.IsHexAddress(addr) {
		return nil, errors.New("invalid feed registry address")
	}
	abiBytes, err := feedRegistryFS.ReadFile("abi/feed_registry.json")
	if err != nil {
		return nil, err
	}
	a, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, err
	}
	return &FeedRegistry{addr: common.HexToAddress(addr), abi: a}, nil
}

func (r *FeedRegistry) Address() common.Address { return r.addr }

// DecodeLatestRoundData: возвращает цену (answer) и updatedAt.
func (r *FeedRegistry) DecodeLatestRoundData(out []byte) (*big.Int, time.Time, error) {
	// latestRoundData returns: (roundId, answer, startedAt, updatedAt, answeredInRound)
	res, err := r.abi.Unpack("latestRoundData", out)
	if err != nil || len(res) != 5 {
		return nil, time.Time{}, errors.New("unpack latestRoundData")
	}
	answer, _ := res[1].(*big.Int)
	updated, _ := res[3].(*big.Int)
	if answer == nil || updated == nil {
		return nil, time.Time{}, errors.New("nil values")
	}
	return answer, time.Unix(updated.Int64(), 0), nil
}

func (r *FeedRegistry) PackLatestRoundData(base, quote common.Address) ([]byte, error) {
	return r.abi.Pack("latestRoundData", base, quote)
}

func (r *FeedRegistry) PackDecimals(base, quote common.Address) ([]byte, error) {
	return r.abi.Pack("decimals", base, quote)
}

func (r *FeedRegistry) UnpackDecimals(out []byte) (uint8, error) {
	res, err := r.abi.Unpack("decimals", out)
	if err != nil || len(res) != 1 {
		return 0, errors.New("unpack decimals")
	}
	switch v := res[0].(type) {
	case uint8:
		return v, nil
	case *big.Int:
		return uint8(v.Uint64()), nil
	default:
		return 0, errors.New("unexpected decimals type")
	}
}
