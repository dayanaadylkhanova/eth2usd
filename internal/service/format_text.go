package service

import (
	"fmt"
	"strings"
)

// FormatText TODO поправить игнор ошибки
func FormatText(r ValuationResult) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "ASSET\tAMOUNT\tUSD\tSOURCE\tERROR\n")
	for _, row := range r.Rows {
		fmt.Fprintf(&b, "%s\t%s\t%s\t%s\t%s\n", row.Symbol, row.Amount, row.USD, row.Source, row.Err)
	}
	fmt.Fprintf(&b, "\nTOTAL USD: %s\n", r.TotalUSD)
	return b.String(), nil
}
