package service

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// FormatAmount converts integer 'raw' with decimals to decimal string with up to 'precision' fractional digits.
func FormatAmount(raw *big.Int, decimals int, precision int) string {
	if raw == nil {
		return "0"
	}
	if decimals == 0 {
		return raw.String()
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	r := new(big.Rat).SetFrac(raw, scale)
	return FormatRat(r, precision)
}

// FormatRat prints big.Rat with fixed max precision, trimming trailing zeros.
func FormatRat(r *big.Rat, precision int) string {
	if r == nil {
		return "0"
	}
	// scale by 10^precision and round
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	scaled := new(big.Rat).Mul(r, new(big.Rat).SetInt(scale))
	i := new(big.Int)
	scaledFloat := new(big.Float).SetRat(scaled)
	scaledFloat.SetMode(big.ToNearestAway)
	scaledFloat.Int(i) // rounded

	intPart := new(big.Int).Quo(i, scale)
	fracPart := new(big.Int).Mod(i, scale)

	if precision == 0 {
		return intPart.String()
	}

	s := fracPart.Text(10)
	// left pad
	if len(s) < precision {
		s = strings.Repeat("0", precision-len(s)) + s
	}
	// trim trailing zeros
	s = strings.TrimRight(s, "0")
	if s == "" {
		return intPart.String()
	}
	return intPart.String() + "." + s
}

// MulDecimalStrings multiplies two decimal strings (a,b) and returns decimal string with 'precision' digits after dot.
func MulDecimalStrings(a, b string, precision int) string {
	af, _, _ := big.ParseFloat(a, 10, 256, big.ToNearestAway)
	bf, _, _ := big.ParseFloat(b, 10, 256, big.ToNearestAway)
	res := new(big.Float).Mul(af, bf)
	// round to precision
	pow := math.Pow10(precision)
	f64, _ := res.Float64()
	f64 = math.Round(f64*pow) / pow
	return strings.TrimRight(strings.TrimRight(fmtFloat(f64, precision), "0"), ".")
}

func fmtFloat(f float64, prec int) string {
	// 'f' format with fixed prec
	return strconv.FormatFloat(f, 'f', prec, 64)
}
