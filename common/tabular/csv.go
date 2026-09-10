package tabular

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"io"
	"strings"

	"go.bryk.io/pkg/errors"
)

var utf8Bom = []byte{0xEF, 0xBB, 0xBF}

// readCsv accepts comma, semicolon and tab delimited text with an optional UTF-8 BOM. The
// delimiter is sniffed from the header line: the candidate that splits it into the most
// columns wins, comma on a tie.
func readCsv(r io.Reader, opts Options) (*Table, error) {
	buffered := bufio.NewReader(r)
	if err := skipBom(buffered); err != nil {
		return nil, err
	}
	headerLine, err := buffered.Peek(buffered.Size())
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return nil, errors.Wrap(err, "peek csv")
	}

	reader := csv.NewReader(buffered)
	reader.Comma = sniffDelimiter(firstLine(headerLine))
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	raw, err := collectCsvRows(reader, opts)
	if err != nil {
		return nil, err
	}
	return build(raw, opts)
}

func skipBom(r *bufio.Reader) error {
	head, err := r.Peek(len(utf8Bom))
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return errors.Wrap(err, "peek csv bom")
	}
	if bytes.Equal(head, utf8Bom) {
		_, err = r.Discard(len(utf8Bom))
	}
	return err
}

func firstLine(buf []byte) string {
	line, _, _ := strings.Cut(string(buf), "\n")
	return line
}

func sniffDelimiter(line string) rune {
	best, bestCount := ',', strings.Count(line, ",")
	for _, candidate := range []rune{';', '\t'} {
		if count := strings.Count(line, string(candidate)); count > bestCount {
			best, bestCount = candidate, count
		}
	}
	return best
}

func collectCsvRows(reader *csv.Reader, opts Options) ([][]string, error) {
	raw := make([][]string, 0, 64)
	limit := 0
	if opts.MaxRows > 0 {
		limit = opts.MaxRows + 2
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return raw, nil
		}
		if err != nil {
			return nil, errors.Wrap(err, "read csv row")
		}
		raw = append(raw, record)
		if limit > 0 && len(raw) > limit {
			return raw, nil
		}
	}
}
