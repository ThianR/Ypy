package api

import (
	"encoding/csv"
	"strings"
)

type StockRow struct {
	CompanyID                               int64
	Item, Location, Reason, Quantity, Value string
}
type CashRow struct {
	CompanyID                      int64
	Currency, Method, Kind, Amount string
}

// ToStockCSV exports already-authorized rows. CompanyID is required so callers
// no pueda producir accidentalmente un informe con registros fuera de ámbito.
func ToStockCSV(rows []StockRow) string {
	clean := make([]StockRow, 0, len(rows))
	for _, r := range rows {
		if r.CompanyID <= 0 || strings.TrimSpace(r.Item) == "" || strings.TrimSpace(r.Quantity) == "" || !decimalAmount.MatchString(r.Quantity) {
			return ""
		}
		if r.Value != "" && !decimalAmount.MatchString(r.Value) {
			return ""
		}
		r.Item, r.Location, r.Reason = strings.TrimSpace(r.Item), strings.TrimSpace(r.Location), strings.TrimSpace(r.Reason)
		clean = append(clean, r)
	}
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"company_id", "item", "location", "reason", "quantity", "value"})
	for _, r := range clean {
		_ = w.Write([]string{fmtInt(r.CompanyID), r.Item, r.Location, r.Reason, r.Quantity, r.Value})
	}
	w.Flush()
	return b.String()
}

func ToCashCSV(rows []CashRow) string {
	clean := make([]CashRow, 0, len(rows))
	for _, r := range rows {
		if r.CompanyID <= 0 || strings.TrimSpace(r.Currency) == "" || strings.TrimSpace(r.Method) == "" || strings.TrimSpace(r.Kind) == "" || !decimalAmount.MatchString(r.Amount) {
			return ""
		}
		r.Currency, r.Method, r.Kind = strings.TrimSpace(r.Currency), strings.TrimSpace(r.Method), strings.TrimSpace(r.Kind)
		clean = append(clean, r)
	}
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"company_id", "currency", "method", "kind", "amount"})
	for _, r := range clean {
		_ = w.Write([]string{fmtInt(r.CompanyID), r.Currency, r.Method, r.Kind, r.Amount})
	}
	w.Flush()
	return b.String()
}

func fmtInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
