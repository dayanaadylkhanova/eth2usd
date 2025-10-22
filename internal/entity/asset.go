package entity

type Asset struct {
	Address  string // "0x..." or "eth://native"
	Symbol   string
	Decimals int
}

type AssetVal struct {
	Asset  Asset
	Amount string // raw amount in smallest units (e.g., wei)
}
