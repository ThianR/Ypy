package api

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// TotalsPDF produce un informe PDF pequeño y válido sin runtimes externos.
func TotalsPDF(totals []Total) ([]byte, error) {
	rows := append([]Total(nil), totals...)
	for _, t := range rows {
		if t.Currency == "" || t.Method == "" || !decimalAmount.MatchString(t.Amount) {
			return nil, ErrInvalidSale
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Currency == rows[j].Currency {
			return rows[i].Method < rows[j].Method
		}
		return rows[i].Currency < rows[j].Currency
	})
	text := "BT /F1 12 Tf 50 790 Td (currency | method | amount) Tj"
	for _, r := range rows {
		text += fmt.Sprintf(" 0 -20 Td (%s | %s | %s) Tj", pdfText(r.Currency), pdfText(r.Method), pdfText(r.Amount))
	}
	text += " ET"
	objects := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>", "<< /Type /Font /Subtype /Type1 /BaseFont /Courier >>", fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(text), text)}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for i, obj := range objects {
		offsets = append(offsets, b.Len())
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets[1:] {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xref)
	return b.Bytes(), nil
}

func pdfText(value string) string {
	return strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)").Replace(value)
}
