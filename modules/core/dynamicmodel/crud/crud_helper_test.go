package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// testRow is a minimal DynamicModel: the constraint only needs the field-data accessors.
type testRow struct {
	fields dmodel.DynamicFields
}

func (this *testRow) GetFieldData() dmodel.DynamicFields     { return this.fields }
func (this *testRow) SetFieldData(data dmodel.DynamicFields) { this.fields = data }

func newSearchResult(rows ...dmodel.DynamicFields) *dyn.OpResult[dyn.PagedResultData[testRow]] {
	items := make([]testRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, testRow{fields: row})
	}
	result := &dyn.OpResult[dyn.PagedResultData[testRow]]{}
	result.Data.Items = items
	return result
}

func setFields(result *dyn.OpResult[dyn.PagedResultData[testRow]], query dyn.SearchQuery) {
	setDefaultDesiredFields[testRow, *testRow](result, query, nil)
}

// A search result reaches the client as JSON, where a nil DesiredFields marshals to `null`.
// Clients read that as "no columns" and render a populated result as an empty table, so every
// path out of Search must name the columns it returned.
func TestSetDefaultDesiredFields(t *testing.T) {
	t.Run("uses the requested fields when the caller asked for some", func(t *testing.T) {
		result := newSearchResult(dmodel.DynamicFields{"code": "A", "name": "N"})

		setFields(result, dyn.SearchQuery{Fields: []string{"code", "name"}})

		assert.Equal(t, []string{"code", "name"}, result.Data.DesiredFields)
	})

	t.Run("falls back to the fields the rows actually carry", func(t *testing.T) {
		result := newSearchResult(
			dmodel.DynamicFields{"code": "A", "name": "N"},
			dmodel.DynamicFields{"code": "B", "colour": "red"},
		)

		setFields(result, dyn.SearchQuery{})

		// Union across rows, so a field only some rows carry still gets a column.
		assert.Equal(t, []string{"code", "colour", "name"}, result.Data.DesiredFields)
	})

	t.Run("is an empty slice, never nil, when there are no rows", func(t *testing.T) {
		result := newSearchResult()

		setFields(result, dyn.SearchQuery{})

		// Nil would marshal to a JSON `null` and crash clients that index into it.
		assert.NotNil(t, result.Data.DesiredFields)
		assert.Empty(t, result.Data.DesiredFields)
	})

	t.Run("leaves an already-assigned value alone", func(t *testing.T) {
		// UiSearch assigns DesiredFields from the user's saved column preferences after its
		// SearchFn returns. This default must never win over that.
		result := newSearchResult(dmodel.DynamicFields{"code": "A"})
		result.Data.DesiredFields = []string{"chosen_by_ui_search"}

		setFields(result, dyn.SearchQuery{Fields: []string{"code"}})

		assert.Equal(t, []string{"chosen_by_ui_search"}, result.Data.DesiredFields)
	})

	t.Run("stays nil on a failed search", func(t *testing.T) {
		result := newSearchResult(dmodel.DynamicFields{"code": "A"})
		result.ClientErrors.Append(*ft.NewAnonymousBusinessViolation("test.err", "search failed"))

		setFields(result, dyn.SearchQuery{Fields: []string{"code"}})

		assert.Nil(t, result.Data.DesiredFields)
	})
}
