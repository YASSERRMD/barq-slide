// Package adapters provides data ingestors that convert external file formats
// into barq-slides chart-ready JSON structures.
package adapters

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// StatProfile summarises the numeric distribution of a column.
type StatProfile struct {
	Column string  `json:"column"`
	Count  int     `json:"count"`
	Sum    float64 `json:"sum"`
	Mean   float64 `json:"mean"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

// XLSXSheet holds parsed data from a single worksheet.
type XLSXSheet struct {
	Name    string            `json:"name"`
	Headers []string          `json:"headers"`
	Rows    [][]string        `json:"rows"`
	Stats   []StatProfile     `json:"stats,omitempty"`
}

// IngestXLSX reads an Excel workbook from path and returns a JSON-encoded
// summary suitable for injecting into chart_data_json asset fields.
func IngestXLSX(path string) ([]byte, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("xlsx: open %q: %w", path, err)
	}
	defer f.Close()

	var sheets []XLSXSheet
	for _, name := range f.GetSheetList() {
		rows, err := f.GetRows(name)
		if err != nil || len(rows) == 0 {
			continue
		}

		sheet := XLSXSheet{Name: name}
		sheet.Headers = rows[0]
		if len(rows) > 1 {
			sheet.Rows = rows[1:]
		}
		sheet.Stats = profileColumns(sheet.Headers, sheet.Rows)
		sheets = append(sheets, sheet)
	}

	return json.Marshal(sheets)
}

// IngestXLSXBytes reads an Excel workbook from raw bytes.
func IngestXLSXBytes(data []byte) ([]byte, error) {
	f, err := excelize.OpenReader(strings.NewReader(""))
	_ = f
	// excelize requires a ReadSeeker; use a temp file approach via OpenReader workaround.
	// Use excelize.OpenBinary if available, otherwise write to temp:
	fr, err := excelize.OpenReader(newByteReader(data))
	if err != nil {
		return nil, fmt.Errorf("xlsx: open bytes: %w", err)
	}
	defer fr.Close()

	var sheets []XLSXSheet
	for _, name := range fr.GetSheetList() {
		rows, err2 := fr.GetRows(name)
		if err2 != nil || len(rows) == 0 {
			continue
		}
		sheet := XLSXSheet{Name: name}
		sheet.Headers = rows[0]
		if len(rows) > 1 {
			sheet.Rows = rows[1:]
		}
		sheet.Stats = profileColumns(sheet.Headers, sheet.Rows)
		sheets = append(sheets, sheet)
	}
	return json.Marshal(sheets)
}

// profileColumns computes numeric statistics for each header column.
func profileColumns(headers []string, rows [][]string) []StatProfile {
	profiles := make([]StatProfile, len(headers))
	for i, h := range headers {
		profiles[i] = StatProfile{Column: h, Min: 0, Max: 0}
	}

	for _, row := range rows {
		for i := range headers {
			if i >= len(row) {
				continue
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(row[i]), 64)
			if err != nil {
				continue
			}
			p := &profiles[i]
			p.Sum += v
			p.Count++
			if p.Count == 1 || v < p.Min {
				p.Min = v
			}
			if p.Count == 1 || v > p.Max {
				p.Max = v
			}
		}
	}

	for i := range profiles {
		if profiles[i].Count > 0 {
			profiles[i].Mean = profiles[i].Sum / float64(profiles[i].Count)
		}
	}

	// Filter out non-numeric columns.
	numeric := profiles[:0]
	for _, p := range profiles {
		if p.Count > 0 {
			numeric = append(numeric, p)
		}
	}
	return numeric
}

// byteReader wraps a []byte to implement io.ReadSeeker.
type byteReader struct {
	data []byte
	pos  int
}

func newByteReader(data []byte) *byteReader { return &byteReader{data: data} }

func (b *byteReader) Read(p []byte) (int, error) {
	if b.pos >= len(b.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, b.data[b.pos:])
	b.pos += n
	return n, nil
}

func (b *byteReader) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(b.pos) + offset
	case 2:
		abs = int64(len(b.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence %d", whence)
	}
	if abs < 0 {
		return 0, fmt.Errorf("negative position")
	}
	b.pos = int(abs)
	return abs, nil
}
