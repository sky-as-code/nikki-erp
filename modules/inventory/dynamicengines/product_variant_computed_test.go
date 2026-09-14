package dynamicengines

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// recordingOnion captures the search params of one schema and answers with canned rows.
type recordingOnion struct {
	composable.DynamicResourceEngineOnion
	repo *recordingRepo
}

func (this *recordingOnion) Repository() composable.CrudRepository { return this.repo }

type recordingRepo struct {
	composable.CrudRepository
	params []dyn.RepoSearchParam
	items  []dmodel.DynamicFields
}

func (this *recordingRepo) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*composable.SearchResult, error) {
	this.params = append(this.params, param)
	return &composable.SearchResult{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.items},
		HasData: len(this.items) > 0,
	}, nil
}

// variantAttributeFixture wires the four schemas the attribute walk reads, each returning the rows
// that link to the next.
func variantAttributeFixture() (
	func(string) (composable.DynamicResourceEngineOnion, bool), map[string]*recordingRepo,
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
				basemodel.FieldId:                             "av_big",
				models.ProductAttributeValueFieldName:         model.LangJson{"en-US": "Big"},
				models.ProductAttributeValueFieldAttributeId:  "attr_size",
			},
			{
				basemodel.FieldId:                             "av_nosugar",
				models.ProductAttributeValueFieldName:         model.LangJson{"en-US": "No sugar"},
				models.ProductAttributeValueFieldAttributeId:  "attr_sugar",
			},
		}},
		models.ProductAttributeSchemaName: {items: []dmodel.DynamicFields{
			{basemodel.FieldId: "attr_size", models.ProductAttributeFieldName: model.LangJson{"en-US": "Size"}},
			{basemodel.FieldId: "attr_sugar", models.ProductAttributeFieldName: model.LangJson{"en-US": "Sugar"}},
		}},
	}
	lookup := func(name string) (composable.DynamicResourceEngineOnion, bool) {
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
		basemodel.FieldId:                          "v1",
		models.ProductVariantFieldTemplateName:     model.LangJson{"en-US": "Coca Cola"},
		models.ProductVariantFieldSku:              "CC-NS-BIG",
	}
}

// The bug this pins: the attribute name sits three edges from the variant, and a first cut asked
// for it as one `template_attribute_value.attribute_value.attribute.name` projection. A `fields=`
// selection resolves exactly one dot, so every read failed with "field path exceeds maximum of 1
// dot separators" and no variant page would load.
func TestVariantAttributeReadsStayWithinTheSelectionDotCap(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	_, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{
		SchemaName: models.ProductVariantSchemaName,
		FieldName:  "display_name",
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

func TestVariantDisplayNameNamesTheTemplateAndItsValuesInSequence(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	out, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, "Coca Cola (No sugar, Big)", out[0])
}

func TestVariantAttributeSummaryPairsEachAttributeWithItsValue(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)

	out, err := fns[VariantAttributeSummaryFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})

	require.NoError(t, err)
	require.Len(t, out, 1)
	assert.Equal(t, []string{"Sugar: No sugar", "Size: Big"}, out[0])
}

// A template with no variant-generating attributes is the ordinary single-variant case, so its
// variant is named by the template alone rather than gaining an empty pair of brackets.
func TestVariantWithNoAttributesIsNamedByItsTemplateAlone(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	repos[models.ProductVariantAttributeValueSchemaName].items = nil
	fns := newVariantComputedFunctions(lookup)

	names, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})
	require.NoError(t, err)
	summaries, err := fns[VariantAttributeSummaryFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow()},
	})
	require.NoError(t, err)

	assert.Equal(t, "Coca Cola", names[0])
	assert.Empty(t, summaries[0])
}

// The engine pairs values with rows by position, so a short or reordered answer misnames records
// rather than failing. Two variants sharing one attribute row exercise the grouping.
func TestVariantComputedReturnsOneValuePerRowInOrder(t *testing.T) {
	lookup, _ := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)
	other := dmodel.DynamicFields{
		basemodel.FieldId:                      "v2",
		models.ProductVariantFieldTemplateName: model.LangJson{"en-US": "Pepsi"},
	}

	out, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{variantRow(), other},
	})

	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, "Coca Cola (No sugar, Big)", out[0])
	assert.Equal(t, "Pepsi", out[1], "a variant with no attribute rows keeps its own name")
}

// The whole page is read in a fixed number of queries: one per hop, never one per row.
func TestVariantAttributeReadIsBatchedPerHop(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	fns := newVariantComputedFunctions(lookup)
	rows := []dmodel.DynamicFields{variantRow(), {
		basemodel.FieldId:                      "v2",
		models.ProductVariantFieldTemplateName: model.LangJson{"en-US": "Pepsi"},
	}}

	_, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{Models: rows})

	require.NoError(t, err)
	for schemaName, repo := range repos {
		assert.LessOrEqualf(t, len(repo.params), 1, "%s was queried more than once", schemaName)
	}
}

// A missing template name would otherwise render the Go zero value of a LangJson into the label.
func TestVariantDisplayNameFallsBackToSkuWithoutATemplateName(t *testing.T) {
	lookup, repos := variantAttributeFixture()
	repos[models.ProductVariantAttributeValueSchemaName].items = nil
	fns := newVariantComputedFunctions(lookup)
	row := dmodel.DynamicFields{basemodel.FieldId: "v1", models.ProductVariantFieldSku: "CC-NS-BIG"}

	out, err := fns[VariantDisplayNameFn](nil, composable.ComputeFnRequest{
		Models: []dmodel.DynamicFields{row},
	})

	require.NoError(t, err)
	assert.Equal(t, "CC-NS-BIG", out[0])
}
