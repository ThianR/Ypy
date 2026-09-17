package api

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"time"
)

type xlsxCell struct {
	Ref   string `xml:"r,attr"`
	Type  string `xml:"t,attr,omitempty"`
	Value string `xml:"v"`
}
type xlsxRow struct {
	Ref   int        `xml:"r,attr"`
	Cells []xlsxCell `xml:"c"`
}
type worksheet struct {
	XMLName   xml.Name `xml:"http://schemas.openxmlformats.org/spreadsheetml/2006/main worksheet"`
	SheetData struct {
		Rows []xlsxRow `xml:"row"`
	} `xml:"sheetData"`
}

// TotalsXLSX creates a deterministic, dependency-free Excel workbook.
func TotalsXLSX(totals []Total) ([]byte, error) {
	clean := make([]Total, 0, len(totals))
	for _, t := range totals {
		if t.Currency == "" || t.Method == "" || !decimalAmount.MatchString(t.Amount) {
			return nil, ErrInvalidSale
		}
		clean = append(clean, t)
	}
	sort.Slice(clean, func(i, j int) bool {
		if clean[i].Currency == clean[j].Currency {
			return clean[i].Method < clean[j].Method
		}
		return clean[i].Currency < clean[j].Currency
	})
	ws := worksheet{}
	ws.SheetData.Rows = append(ws.SheetData.Rows, xlsxRow{Ref: 1, Cells: []xlsxCell{{Ref: "A1", Type: "str", Value: "currency"}, {Ref: "B1", Type: "str", Value: "method"}, {Ref: "C1", Type: "str", Value: "amount"}}})
	for i, t := range clean {
		n := i + 2
		ws.SheetData.Rows = append(ws.SheetData.Rows, xlsxRow{Ref: n, Cells: []xlsxCell{{Ref: fmt.Sprintf("A%d", n), Type: "str", Value: t.Currency}, {Ref: fmt.Sprintf("B%d", n), Type: "str", Value: t.Method}, {Ref: fmt.Sprintf("C%d", n), Value: t.Amount}}})
	}
	data, err := xml.Marshal(ws)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	z := zip.NewWriter(&out)
	files := map[string]string{"[Content_Types].xml": `<?xml version="1.0" encoding="UTF-8"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`, `_rels/.rels`: `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`, `xl/workbook.xml`: `<?xml version="1.0" encoding="UTF-8"?><workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sales" sheetId="1" r:id="rId1"/></sheets></workbook>`, `xl/_rels/workbook.xml.rels`: `<?xml version="1.0" encoding="UTF-8"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`}
	add := func(name, content string) error {
		h := &zip.FileHeader{Name: name, Method: zip.Store}
		h.SetModTime(time.Unix(0, 0).UTC())
		w, e := z.CreateHeader(h)
		if e != nil {
			return e
		}
		_, e = w.Write([]byte(content))
		return e
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels"} {
		if e := add(name, files[name]); e != nil {
			return nil, e
		}
	}
	if err := add("xl/worksheets/sheet1.xml", string(data)); err != nil {
		return nil, err
	}
	if err = z.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
