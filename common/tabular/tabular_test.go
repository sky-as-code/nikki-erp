package tabular

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	file, err := os.Open(filepath.Join("testdata", name))
	require.NoError(t, err)
	t.Cleanup(func() { file.Close() })
	return file
}

func buildXlsx(t *testing.T, rows [][]any) *bytes.Buffer {
	t.Helper()
	book := excelize.NewFile()
	sheet := book.GetSheetName(0)
	for i, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, i+1)
		require.NoError(t, err)
		require.NoError(t, book.SetSheetRow(sheet, cell, &row))
	}
	buf, err := book.WriteToBuffer()
	require.NoError(t, err)
	return buf
}

func TestReadCsvSniffsCommaAndStripsBom(t *testing.T) {
	table, err := Read(openFixture(t, "comma_bom.csv"), ".CSV", Options{})

	require.NoError(t, err)
	assert.Equal(t, []string{"Name", "Address", "Category"}, table.Headers)
	require.Len(t, table.Rows, 2)
	assert.Equal(t, []string{"Coke", "12 Main St", "Beverage"}, table.Rows[0])
	assert.Equal(t, []string{"Pepsi, Max", "", "Beverage"}, table.Rows[1])
}

func TestReadCsvSniffsSemicolonAndDropsEmptyTrailingRow(t *testing.T) {
	table, err := Read(openFixture(t, "semicolon.csv"), "csv", Options{})

	require.NoError(t, err)
	assert.Equal(t, []string{"Tên", "Địa chỉ"}, table.Headers)
	assert.Equal(t, [][]string{{"Bánh mì", "Sài Gòn"}, {"Phở", "Hà Nội"}}, table.Rows)
}

func TestReadCsvWithoutHeaderFails(t *testing.T) {
	_, err := Read(openFixture(t, "empty.csv"), "csv", Options{})

	assert.ErrorIs(t, err, ErrNoHeader)
}

func TestReadCsvEnforcesRowLimit(t *testing.T) {
	src := strings.NewReader("a,b\n1,2\n3,4\n5,6\n")

	_, err := Read(src, "csv", Options{MaxRows: 2})

	assert.ErrorIs(t, err, ErrTooManyRows)
}

func TestReadCsvAtRowLimitPasses(t *testing.T) {
	src := strings.NewReader("a,b\n1,2\n3,4\n")

	table, err := Read(src, "csv", Options{MaxRows: 2})

	require.NoError(t, err)
	assert.Len(t, table.Rows, 2)
}

func TestReadXlsxFirstSheetNormalisesWidth(t *testing.T) {
	buf := buildXlsx(t, [][]any{
		{"  Tên  sản phẩm ", "Giá", "", ""},
		{"Coke", 12000, "extra cell"},
		{"Pepsi"},
		{},
	})

	table, err := Read(buf, "xlsx", Options{})

	require.NoError(t, err)
	assert.Equal(t, []string{"Tên sản phẩm", "Giá"}, table.Headers)
	assert.Equal(t, [][]string{{"Coke", "12000"}, {"Pepsi", ""}}, table.Rows)
}

func TestReadXlsxSkipsLeadingBlankRows(t *testing.T) {
	buf := buildXlsx(t, [][]any{{}, {"", ""}, {"code", "name"}, {"A1", "Alpha"}})

	table, err := Read(buf, "xlsx", Options{})

	require.NoError(t, err)
	assert.Equal(t, []string{"code", "name"}, table.Headers)
	assert.Equal(t, [][]string{{"A1", "Alpha"}}, table.Rows)
}

func TestReadXlsxEnforcesRowLimit(t *testing.T) {
	buf := buildXlsx(t, [][]any{{"h"}, {"1"}, {"2"}, {"3"}})

	_, err := Read(buf, "xlsx", Options{MaxRows: 2})

	assert.ErrorIs(t, err, ErrTooManyRows)
}

func TestReadRejectsUnknownExtension(t *testing.T) {
	_, err := Read(strings.NewReader("x"), "xls", Options{})

	assert.ErrorIs(t, err, ErrUnsupportedExtension)
}

func TestReadXlsxRejectsGarbage(t *testing.T) {
	_, err := Read(strings.NewReader("not a workbook"), "xlsx", Options{})

	assert.Error(t, err)
}
