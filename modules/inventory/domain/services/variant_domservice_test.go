package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// fakeBaseService stands in for the engine's default resource service, recording the params it
// was handed so a test can assert on what the override rewrote.
type fakeBaseService struct {
	drif.DynamicResourceService

	searchCalls  int
	getByIdCalls int
	lastParams   dmodel.DynamicFields

	searchRows []dmodel.DynamicFields
	singleRow  dmodel.DynamicFields
}

func (this *fakeBaseService) Search(
	_ corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.searchCalls++
	this.lastParams = params
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.searchRows, Total: len(this.searchRows)},
		HasData: true,
	}, nil
}

func (this *fakeBaseService) GetById(
	_ corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	this.getByIdCalls++
	this.lastParams = params
	return &dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]]{
		Data:    dyn.SingleResultData[dmodel.DynamicFields]{Item: this.singleRow},
		HasData: true,
	}, nil
}

// countingEngine records how many template reads the fill performed. The whole point of the
// batched design is that this stays at 1 regardless of page size.
type countingEngine struct {
	drif.DynamicResourceEngine

	repo *countingRepo
}

func (this *countingEngine) ResourceRepository() drif.DynamicResourceRepository {
	return this.repo
}

type countingRepo struct {
	drif.DynamicResourceRepository

	searchCalls int
	lastParam   dyn.RepoSearchParam
	rows        []dmodel.DynamicFields
}

func (this *countingRepo) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.searchCalls++
	this.lastParam = param
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.rows, Total: len(this.rows)},
		HasData: true,
	}, nil
}

// withStubbedEngine swaps the package-level engine resolver for the duration of one test. That
// resolver is a var precisely so a unit test can do this: the real registry is populated during
// Init and cannot be built here.
func withStubbedEngine(t *testing.T, repo *countingRepo) {
	t.Helper()
	original := engineFor
	engineFor = func(string) (drif.DynamicResourceEngine, error) {
		return &countingEngine{repo: repo}, nil
	}
	t.Cleanup(func() { engineFor = original })
}

func templateRow(id, name string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		"id":   id,
		"name": model.LangJson{"en-US": name},
	}
}

func variantRow(id, templateId string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		"id":                  id,
		"product_template_id": templateId,
	}
}

// The headline guarantee: a page of 50 variants costs ONE variant query plus ONE template query,
// not one per row. This is the regression test for the whole batched design -- an implementation
// that reverted to per-row edge hydration would show 50 here.
func TestVariantSearch_FillsTemplateFieldsInOneBatchedQuery(t *testing.T) {
	rows := make([]dmodel.DynamicFields, 0, 50)
	for i := 0; i < 50; i++ {
		rows = append(rows, variantRow("v"+string(rune('a'+i%26)), "tpl-1"))
	}
	base := &fakeBaseService{searchRows: rows}
	repo := &countingRepo{rows: []dmodel.DynamicFields{templateRow("tpl-1", "Classic T-Shirt")}}
	withStubbedEngine(t, repo)

	service := NewProductVariantDomainService(base)
	result, err := service.Search(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "template_name"},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, base.searchCalls, "one query for the variants")
	assert.Equal(t, 1, repo.searchCalls, "one batched query for the templates, not one per row")

	for _, row := range result.Data.Items {
		assert.Contains(t, row, "template_name", "every row is filled")
	}
}

// A search that wants no template_* field must cost exactly what it did before this change.
func TestVariantSearch_NoTemplateFieldsRequestedSkipsTemplateRead(t *testing.T) {
	base := &fakeBaseService{searchRows: []dmodel.DynamicFields{variantRow("v1", "tpl-1")}}
	repo := &countingRepo{}
	withStubbedEngine(t, repo)

	service := NewProductVariantDomainService(base)
	_, err := service.Search(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "sku"},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, base.searchCalls)
	assert.Equal(t, 0, repo.searchCalls, "no template_* requested means no template read at all")
}

// The fill keys rows back to templates by product_template_id, so the search must ask for it even
// when the caller did not -- otherwise there is nothing to join on.
func TestVariantSearch_ForcesTemplateIdIntoFieldList(t *testing.T) {
	base := &fakeBaseService{searchRows: []dmodel.DynamicFields{variantRow("v1", "tpl-1")}}
	repo := &countingRepo{rows: []dmodel.DynamicFields{templateRow("tpl-1", "Shirt")}}
	withStubbedEngine(t, repo)

	service := NewProductVariantDomainService(base)
	_, err := service.Search(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "template_name"},
	})

	require.NoError(t, err)
	assert.Contains(t, readRequestedFields(base.lastParams), "product_template_id")
}

// Several variants of one template must not each trigger a read.
func TestVariantSearch_DuplicateTemplateIdsFetchedOnce(t *testing.T) {
	base := &fakeBaseService{searchRows: []dmodel.DynamicFields{
		variantRow("v1", "tpl-1"),
		variantRow("v2", "tpl-1"),
		variantRow("v3", "tpl-2"),
	}}
	repo := &countingRepo{rows: []dmodel.DynamicFields{
		templateRow("tpl-1", "Shirt"),
		templateRow("tpl-2", "Hat"),
	}}
	withStubbedEngine(t, repo)

	service := NewProductVariantDomainService(base)
	_, err := service.Search(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "template_name"},
	})

	require.NoError(t, err)
	assert.Equal(t, 1, repo.searchCalls)
	assert.Len(t, distinctTemplateIds(base.searchRows), 2, "two distinct templates, three rows")
}

// A variant whose template is gone must read as "unknown", not as a product with an empty name.
func TestVariantSearch_MissingTemplateLeavesFieldsAbsent(t *testing.T) {
	base := &fakeBaseService{searchRows: []dmodel.DynamicFields{variantRow("v1", "tpl-missing")}}
	repo := &countingRepo{rows: []dmodel.DynamicFields{}}
	withStubbedEngine(t, repo)

	service := NewProductVariantDomainService(base)
	result, err := service.Search(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "template_name"},
	})

	require.NoError(t, err)
	assert.NotContains(t, result.Data.Items[0], "template_name")
}

// GetById asks for the template edge so the repository hydrates it, then flattens it onto the
// row's template_* fields.
func TestVariantGetById_AddsTemplateEdgeAndFills(t *testing.T) {
	base := &fakeBaseService{singleRow: dmodel.DynamicFields{
		"id":                  "v1",
		"product_template_id": "tpl-1",
		"template":            templateRow("tpl-1", "Classic T-Shirt"),
	}}
	withStubbedEngine(t, &countingRepo{})

	service := NewProductVariantDomainService(base)
	result, err := service.GetById(nil, dmodel.DynamicFields{
		paramFieldNames: []string{"id", "template_name"},
	})

	require.NoError(t, err)
	assert.Contains(t, readRequestedFields(base.lastParams), "template",
		"the edge must be requested or there is nothing to flatten")
	assert.Contains(t, result.Data.Item, "template_name")
	assert.Contains(t, result.Data.Item, "template", "the edge itself stays in the response")
}

// An empty field list already means "everything", so it must be left alone rather than narrowed.
func TestVariantGetById_EmptyFieldListLeftUntouched(t *testing.T) {
	base := &fakeBaseService{singleRow: dmodel.DynamicFields{"id": "v1"}}
	withStubbedEngine(t, &countingRepo{})

	service := NewProductVariantDomainService(base)
	_, err := service.GetById(nil, dmodel.DynamicFields{})

	require.NoError(t, err)
	assert.Empty(t, readRequestedFields(base.lastParams))
}

// A virtual field has no column, so filtering or sorting on it only works by rewriting it to the
// edge path the SQL layer can resolve.
func TestRewriteTemplateFieldPath(t *testing.T) {
	testCases := []struct {
		field   string
		want    string
		rewrote bool
	}{
		{"template_name", "template.name", true},
		{"template_category_id", "template.category_id", true},
		{"template_sale_ok", "template.sale_ok", true},
		{"sku", "sku", false},
		{"id", "id", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.field, func(t *testing.T) {
			got, rewrote := RewriteTemplateFieldPath(testCase.field)
			assert.Equal(t, testCase.want, got)
			assert.Equal(t, testCase.rewrote, rewrote)
		})
	}
}

// template.name is exactly one dot, which is the ORM's limit for both select and order-by. A
// second hop would be rejected, so the rewrite must never produce one.
func TestRewriteTemplateFieldPath_StaysWithinOneDot(t *testing.T) {
	for field := range map[string]string{"template_name": "", "template_status": ""} {
		rewritten, ok := RewriteTemplateFieldPath(field)
		require.True(t, ok)
		assert.Equal(t, 1, countDots(rewritten), "%q must stay within the 1-dot limit", rewritten)
	}
}

func countDots(value string) int {
	count := 0
	for _, char := range value {
		if char == '.' {
			count++
		}
	}
	return count
}

// The source map drives both the batched fill and the rewrite, so a missing entry would silently
// make a field unfillable.
func TestTemplateSourceColumns_AlwaysIncludesId(t *testing.T) {
	columns := templateSourceColumns([]string{"template_name", "template_status"})

	assert.Contains(t, columns, "id", "rows are keyed by template id")
	assert.Contains(t, columns, "name")
	assert.Contains(t, columns, "status")
}
