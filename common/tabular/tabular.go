// Package tabular reads the first sheet of an xlsx workbook or a csv document into a header
// row plus string rows. It is the reading half of resource import: cells are handed over as
// text and typed by the caller against the target schema.
package tabular

import (
	"io"
	"strings"

	"go.bryk.io/pkg/errors"
)

const (
	ExtXlsx = "xlsx"
	ExtCsv  = "csv"
)

var ErrTooManyRows = errors.New("tabular: row limit exceeded")
var ErrUnsupportedExtension = errors.New("tabular: unsupported file extension")
var ErrNoHeader = errors.New("tabular: file has no header row")

// Table is a parsed file. Rows are data rows only; every row is padded or truncated to the
// header width so callers may index by column position safely.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Options bounds a read. MaxRows counts data rows, the header excluded; zero means unbounded.
type Options struct {
	MaxRows int
}

// Read parses r according to ext ("xlsx" or "csv", case-insensitive, dot optional).
func Read(r io.Reader, ext string, opts Options) (*Table, error) {
	switch strings.ToLower(strings.TrimPrefix(strings.TrimSpace(ext), ".")) {
	case ExtXlsx:
		return readXlsx(r, opts)
	case ExtCsv:
		return readCsv(r, opts)
	default:
		return nil, errors.Wrap(ErrUnsupportedExtension, ext)
	}
}

// build turns raw rows into a Table: the first non-empty row is the header, trailing empty
// rows are dropped, cells are trimmed and every row is normalised to the header width.
func build(raw [][]string, opts Options) (*Table, error) {
	headerIdx := firstNonEmptyRow(raw)
	if headerIdx < 0 {
		return nil, ErrNoHeader
	}
	headers := normalizeHeaders(raw[headerIdx])
	rows := make([][]string, 0, len(raw)-headerIdx-1)
	for _, row := range raw[headerIdx+1:] {
		if isEmptyRow(row) {
			continue
		}
		rows = append(rows, fitRow(row, len(headers)))
	}
	if opts.MaxRows > 0 && len(rows) > opts.MaxRows {
		return nil, ErrTooManyRows
	}
	return &Table{Headers: headers, Rows: rows}, nil
}

func firstNonEmptyRow(raw [][]string) int {
	for i, row := range raw {
		if !isEmptyRow(row) {
			return i
		}
	}
	return -1
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// normalizeHeaders trims and collapses inner whitespace so "Tên  sản phẩm " equals "Tên sản phẩm".
func normalizeHeaders(row []string) []string {
	headers := make([]string, len(row))
	for i, cell := range row {
		headers[i] = strings.Join(strings.Fields(cell), " ")
	}
	return trimTrailingEmpty(headers)
}

func trimTrailingEmpty(cells []string) []string {
	end := len(cells)
	for end > 0 && cells[end-1] == "" {
		end--
	}
	return cells[:end]
}

func fitRow(row []string, width int) []string {
	fitted := make([]string, width)
	for i := 0; i < width && i < len(row); i++ {
		fitted[i] = strings.TrimSpace(row[i])
	}
	return fitted
}
