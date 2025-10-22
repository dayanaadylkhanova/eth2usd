package tokens

import (
	"encoding/json"
	"os"
)

type Token struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals,omitempty"`
}

func Load(path string) ([]Token, error) {
	if path == "" {
		return DefaultList, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []Token
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

var DefaultList = []Token{
	{Address: "eth://native", Symbol: "ETH", Decimals: 18},
}
