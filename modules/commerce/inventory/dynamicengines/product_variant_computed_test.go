package dynamicengines

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// recordingOnion captures the search params of one schema and answers with canned rows.
type recordingOnion struct {
	dynamicresource.DynamicResourceEngineOnion
	repo *recordingRepo
}

func (this *recordingOnion) Repository() dynamicresource.CrudRepository { return this.repo }

type recordingRepo struct {
	dynamicresource.CrudRepository
	params []dyn.RepoSearchParam
	items  []dmodel.DynamicFields
}

func (this *recordingRepo) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dynamicresource.SearchResult, error) {
	this.params = append(this.params, param)
	return &dynamicresource.SearchResult{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.items},
		HasData: len(this.items) > 0,
	}, nil
}

// variantAttributeFixture wires the four schemas the attribute walk reads, each returning the rows
// that link to the next.
func variantAttributeFixture() (
	func(string) (dynamicresource.DynamicResourceEngineOnion, bool), map[string]*recordingRepo,
) {
	repos := map[string]*recordingRepo{
		models.ProductVariantAttributeValueSchemaName: {items: []dmodel.DynamicFields{
			{
				basemodel.FieldId: "j1",
				models.ProductVariantAttributeValueFieldProductVariantId:         "v1",
				models.ProductVariantAttributeValueFieldTemplateAttributeValueId: "tav_big",
			},
			{
				basemodel.FieldId: "j2",
				models.ProductVariantAttributeValueFieldProductVariantId:         "v1",
				models.ProductVariantAttributeValueFieldTemplateAttributeValueId: "tav_nosugar",
			},
		}},
		models.ProductTemplateAttributeValueSchemaName: {items: []dmodel.DynamicFields{
			// Declared out of order, so the sequence sort is what puts them right.
			{
				basemodel.FieldId: "tav_big",
				models.ProductTemplateAttributeValueFieldAttributeValueId: "av_big",
				models.ProductTemplateAttributeValueFieldSequence:         int64(2),
			},
			{
				basemodel.FieldId: "tav_nosugar",
				models.ProductTemplateAttributeValueFieldAttributeValueId: "av_nosugar",
				models.ProductTemplateAttributeValueFieldSequence:         int64(1),
			},
		}},
		models.ProductAttributeValueSchemaName: {items: []dmodel.DynamicFields{
			{
				basemodel.FieldId:                            "av_big",
				models.ProductAttributeValueFieldName:        model.LangJson{"en-US": "Big"},
				models.ProductAttributeValueFieldAttributeId: "attr_size",
			},
			{
				basemodel.FieldId:                            "av_nosugar",
				models.ProductAttributeValueFieldName:        model.LangJson{"en-US": "No sugar"},
				models.ProductAttributeValueFieldAttributeId: "attr_sugar",
			},
		}},
		models.ProductAttributeSchemaName: {items: []dmodel.DynamicFields{
			{basemodel.FieldId: "attr_size", models.ProductAttributeFieldName: model.LangJson{"en-US": "Size"}},
			{basemodel.FieldId: "attr_sugar", models.ProductAttributeFieldName: model.LangJson{"en-US": "Sugar"}},
		}},
	}
	lookup := func(name string) (dynamicresource.DynamicResourceEngineOnion, bool) {
		repo, ok := repos[name]
		if !ok {
			return nil, false
		}
		return &recordingOnion{repo: repo}, true
	}
	return lookup, repos
}

func variantRow() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		basemodel.FieldId:                      "v1",
		models.ProductVariantFieldTemplateName: model.LangJson{"en-US": "Coca Cola"},
		models.ProductVariantFieldSku:          "CC-NS-BIG",
	}
}

// The bug this pins: the attribute name sits three edges from the variant, and a first cut asked
// for it as one `template_attribute_value.attribute_value.attribute.name` projection. A `fields=`
// selection resolves exactly one dot, so every read failed with "field path exceeds maximum of 1
// dot separators" and no variant page would load.
func TestVariantAttributeReadsStayWithinTheSelectionDotCap(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	_, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{
		SchemaName: models.ProductVariantSchemaName,
		FieldName:  "attribute_summary",
		Models:     []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	for schemaName, repo := range repos {
		for _, param := range repo.params {
			for _, field := range param.Fields {
				assert.LessOrEqualf(t, strings.Count(field, "."), orm.MaxSelectGraphColumnDots,
					"%s selects %q, deeper than a projection resolves", schemaName, field)
			}
		}
	}
}

// display_name is stored now; only the summary stays a function.
func TestVariantDisplayNameIsNoLongerAFunction(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	assert.Len(t, fns, 1)
	assert.Contains(t, fns, VariantAttributeSummaryFn)
}

func TestVariantAttributeSummaryPairsEachAttributeWithItsValue(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	out, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, []string{"Sugar: No sugar", "Size: Big"}, out[0])
}

// A template with no variant-generating attributes is the ordinary single-variant case.
func TestVariantWithNoAttributesHasAnEmptySummary(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	repos[models.ProductVariantAttributeValueSchemaName].items = nil
	fns := newVariantComputedFunctions(lookup)

	summaries, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	assert.Empty(t, summaries[0])
}

// The engine pairs values with rows by position, so a short or reordered answer mislabels records
// rather than failing. Two variants sharing one attribute row exercise the grouping.
func TestVariantComputedReturnsOneValuePerRowInOrder(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)
	other := dmodel.DynamicFields{
		basemodel.FieldId:                      "v2",
		models.ProductVariantFieldTemplateName: model.LangJson{"en-US": "Pepsi"},
	}

	out, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow(), other},
	})

	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, []string{"Sugar: No sugar", "Size: Big"}, out[0])
	assert.Empty(t, out[1], "a variant with no attribute rows has no summary")
}

// The whole page is read in a fixed number of queries: one per hop, never one per row.
func TestVariantAttributeReadIsBatchedPerHop(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)
	rows := []dmodel.DynamicFields{variantRow(), {
		basemodel.FieldId:                      "v2",
		models.ProductVariantFieldTemplateName: model.LangJson{"en-US": "Pepsi"},
	}}

	_, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{Models: rows})

	require.NoError(t, err)
	for schemaName, repo := range repos {
		assert.LessOrEqualf(t, len(repo.params), 1, "%s was queried more than once", schemaName)
	}
}

// The template's line order wins over the value order within a line: the values' own sequences
// are positions in two different lines, and comparing them across lines ordered the same product
// "Big, No sugar" or "No sugar, Big" depending on which value happened to sit first in its line.
func TestVariantAttributeSummaryFollowsTheTemplateLineOrder(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	templateValues := repos[models.ProductTemplateAttributeValueSchemaName].items
	templateValues[0][models.ProductTemplateAttributeValueFieldTemplateAttributeId] = "line_size"
	templateValues[0][models.ProductTemplateAttributeValueFieldSequence] = int64(1)
	templateValues[1][models.ProductTemplateAttributeValueFieldTemplateAttributeId] = "line_sugar"
	templateValues[1][models.ProductTemplateAttributeValueFieldSequence] = int64(5)
	lineRepo := &recordingRepo{items: []dmodel.DynamicFields{
		{basemodel.FieldId: "line_size", models.ProductTemplateAttributeFieldSequence: int64(2)},
		{basemodel.FieldId: "line_sugar", models.ProductTemplateAttributeFieldSequence: int64(1)},
	}}
	withLines := func(name string) (dynamicresource.DynamicResourceEngineOnion, bool) {
		if name == models.ProductTemplateAttributeSchemaName {
			return &recordingOnion{repo: lineRepo}, true
		}
		return lookup(name)
	}
	fns := newVariantComputedFunctions(withLines)

	out, err := fns[VariantAttributeSummaryFn](nil, dynamicresource.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"Sugar: No sugar", "Size: Big"}, out[0])
	assert.Len(t, lineRepo.params, 1, "the line sequences are read once for the page")
}
