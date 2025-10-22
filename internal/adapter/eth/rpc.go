package eth

import (
	"context"
	"embed"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Встраиваем ABI ERC-20 из соседней папки.
//
//go:embed abi/erc20.json
var erc20FS embed.FS

// Client wraps go-ethereum client + cached ABIs.
type Client struct {
	Eth      *ethclient.Client
	erc20ABI abi.ABI
}

func NewClient(ctx context.Context, endpoint string) (*Client, error) {
	c, err := ethclient.DialContext(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	abiBytes, err := erc20FS.ReadFile("abi/erc20.json")
	if err != nil {
		c.Close()
		return nil, err
	}
	erc, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		c.Close()
		return nil, err
	}
	return &Client{Eth: c, erc20ABI: erc}, nil
}

func (c *Client) Close() { c.Eth.Close() }

func (c *Client) GetBalance(ctx context.Context, account common.Address) (*big.Int, error) {
	return c.Eth.BalanceAt(ctx, account, nil)
}

// --- ERC20 ---

func (c *Client) ERC20BalanceOf(ctx context.Context, token, account common.Address) (*big.Int, error) {
	data, err := c.erc20ABI.Pack("balanceOf", account)
	if err != nil {
		return nil, err
	}
	msg := ethereum.CallMsg{To: &token, Data: data}
	out, err := c.Eth.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}
	results, err := c.erc20ABI.Unpack("balanceOf", out)
	if err != nil || len(results) != 1 {
		return nil, errors.New("bad balanceOf unpack")
	}
	balance, _ := results[0].(*big.Int)
	if balance == nil {
		return nil, errors.New("nil balance")
	}
	return balance, nil
}

func (c *Client) ERC20Decimals(ctx context.Context, token common.Address) (uint8, error) {
	data, err := c.erc20ABI.Pack("decimals")
	if err != nil {
		return 0, err
	}
	msg := ethereum.CallMsg{To: &token, Data: data}
	out, err := c.Eth.CallContract(ctx, msg, nil)
	if err != nil {
		return 0, err
	}
	results, err := c.erc20ABI.Unpack("decimals", out)
	if err != nil || len(results) != 1 {
		return 0, errors.New("bad decimals unpack")
	}
	switch v := results[0].(type) {
	case uint8:
		return v, nil
	case *big.Int:
		return uint8(v.Uint64()), nil
	default:
		return 0, errors.New("unexpected decimals type")
	}
}

func (c *Client) ERC20Symbol(ctx context.Context, token common.Address) (string, error) {
	data, err := c.erc20ABI.Pack("symbol")
	if err != nil {
		return "", err
	}
	msg := ethereum.CallMsg{To: &token, Data: data}
	out, err := c.Eth.CallContract(ctx, msg, nil)
	if err != nil {
		return "", err
	}
	// Try string first
	if res, err := c.erc20ABI.Unpack("symbol", out); err == nil && len(res) == 1 {
		if s, ok := res[0].(string); ok && s != "" {
			return s, nil
		}
	}
	// Fallback: bytes32 символы
	if len(out) >= 32 {
		trim := make([]byte, 0, 32)
		for _, b := range out[:32] {
			if b == 0x00 {
				break
			}
			trim = append(trim, b)
		}
		if len(trim) > 0 {
			return string(trim), nil
		}
	}
	return "", errors.New("bad symbol unpack")
}
