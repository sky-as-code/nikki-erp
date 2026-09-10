package tabular

import (
	"io"

	"github.com/xuri/excelize/v2"
	"go.bryk.io/pkg/errors"
)

// readXlsx reads the first worksheet. Rows are streamed so a large workbook does not have to
// be materialised beyond the row cap before the limit is reported.
func readXlsx(r io.Reader, opts Options) (*Table, error) {
	book, err := excelize.OpenReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "open xlsx")
	}
	defer book.Close()

	sheets := book.GetSheetList()
	if len(sheets) == 0 {
		return nil, ErrNoHeader
	}
	rows, err := book.Rows(sheets[0])
	if err != nil {
		return nil, errors.Wrap(err, "read xlsx sheet")
	}
	defer rows.Close()

	raw, err := collectXlsxRows(rows, opts)
	if err != nil {
		return nil, err
	}
	return build(raw, opts)
}

func collectXlsxRows(rows *excelize.Rows, opts Options) ([][]string, error) {
	raw := make([][]string, 0, 64)
	// +1 for the header, +1 to detect the overflow instead of silently truncating.
	limit := 0
	if opts.MaxRows > 0 {
		limit = opts.MaxRows + 2
	}
	for rows.Next() {
		cells, err := rows.Columns()
		if err != nil {
			return nil, errors.Wrap(err, "read xlsx row")
		}
		raw = append(raw, cells)
		if limit > 0 && len(raw) > limit {
			break
		}
	}
	return raw, rows.Error()
}
