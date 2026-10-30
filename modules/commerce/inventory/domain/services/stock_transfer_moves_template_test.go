package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func TestLiftMovesOrRefuseAcceptsTheTemplateForm(t *testing.T) {
	delta, cErrs := liftMovesOrRefuse(dmodel.DynamicFields{"moves": map[string]any{
		"add": []any{map[string]any{
			"product_template_id": "T1", "demand_quantity": "1",
			"attributes": []any{map[string]any{"product_attribute_id": "A1", "product_attribute_value_id": "V1"}},
		}},
		"update": []any{map[string]any{"id": "M1", "etag": "e", "product_template_id": "T1", "attributes": []any{}}},
	}})

	require.Empty(t, cErrs)
	require.Len(t, delta.Add, 1)
	assert.Equal(t, "T1", *delta.Add[0].ProductTemplateId)
	assert.Equal(t, "V1", delta.Add[0].Attributes[0].ProductAttributeValueId)
	assert.True(t, itStock.ReadTransferMovesDeltaParts(dmodel.DynamicFields{"moves": map[string]any{
		"update": []any{map[string]any{"id": "M1", "product_template_id": "T1"}},
	}}).Resolve)
}

func TestLiftMovesOrRefuseRefusesAMixedOrIncompleteTemplateForm(t *testing.T) {
	_, cErrs := liftMovesOrRefuse(dmodel.DynamicFields{"moves": map[string]any{
		"add": []any{
			map[string]any{"product_variant_id": "V1", "product_template_id": "T1", "demand_quantity": "1"},
			map[string]any{"product_variant_id": "V1", "attributes": []any{}, "demand_quantity": "1"},
			map[string]any{"product_template_id": "T1", "demand_quantity": "1", "attributes": []any{
				map[string]any{"product_attribute_id": "A1"},
				map[string]any{"product_attribute_id": "A1", "product_attribute_value_id": "V2"},
			}},
		},
		"update": []any{map[string]any{"id": "M1", "product_variant_id": "V1", "product_template_id": "T1"}},
	}})

	assert.ElementsMatch(t, []string{
		"moves.add[0].product_template_id",
		"moves.add[1].attributes",
		"moves.add[2].attributes[0].product_attribute_value_id",
		"moves.add[2].attributes[1].product_attribute_id",
		"moves.update[0].product_template_id",
	}, errorFields(cErrs))
	assert.Equal(t, "inventory:stock_transfer.move_product_ambiguous", cErrs[0].Key)
}

type fakeVariantResolver struct {
	result *itProduct.ResolveCompleteVariantResult
	query  itProduct.ResolveCompleteVariantQuery
}

func (this *fakeVariantResolver) ResolveCompleteVariant(
	_ corectx.Context, query itProduct.ResolveCompleteVariantQuery,
) (*itProduct.ResolveCompleteVariantResult, error) {
	this.query = query
	return this.result, nil
}

type templateRowRepository struct {
	composable.CrudRepository
	found  bool
	filter dmodel.DynamicFields
}

func (this *templateRowRepository) GetOne(
	_ corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.filter = param.Filter
	return &dyn.OpResult[dmodel.DynamicFields]{HasData: this.found, Data: dmodel.DynamicFields{"id": "T1"}}, nil
}

func templateFormWriter(
	t *testing.T, templateFound bool, result *itProduct.ResolveCompleteVariantResult,
) (*transferMovesWriter, *fakeVariantResolver, *templateRowRepository) {
	t.Helper()
	repo := &templateRowRepository{found: templateFound}
	original := repoFor
	t.Cleanup(func() { repoFor = original })
	repoFor = func(string) (composable.CrudRepository, error) { return repo, nil }
	resolver := &fakeVariantResolver{result: result}
	writer := &transferMovesWriter{
		transfer: *models.NewStockTransferFrom(dmodel.DynamicFields{"org_id": "01ORG"}), resolver: resolver,
	}
	return writer, resolver, repo
}

func templateFormDelta() itStock.TransferMovesDelta {
	templateId := "T1"
	return itStock.TransferMovesDelta{
		Add: []itStock.TransferMoveAdd{{ProductVariantId: "V0"}, {ProductTemplateId: &templateId,
			Attributes: []itStock.TransferMoveAttribute{
				{ProductAttributeId: "A1", ProductAttributeValueId: "AV1"},
				{ProductAttributeId: "A2", ProductAttributeValueId: "AV2"},
			}}},
		Update: []itStock.TransferMoveUpdate{{Id: "M1", ProductTemplateId: &templateId}},
	}
}

func TestResolveTemplateFormsMaterializesInTheTransfersOrg(t *testing.T) {
	writer, resolver, repo := templateFormWriter(t, true, &itProduct.ResolveCompleteVariantResult{HasData: true,
		Data: itProduct.ResolveCompleteVariantResultData{VariantId: "V9", Outcome: itProduct.ResolveOutcomeCreated}})
	delta := templateFormDelta()

	cErrs, err := writer.resolveTemplateForms(callerContext(), delta)

	require.NoError(t, err)
	require.Empty(t, cErrs)
	assert.Equal(t, "V0", delta.Add[0].ProductVariantId)
	assert.Equal(t, "V9", delta.Add[1].ProductVariantId)
	assert.Equal(t, "V9", *delta.Update[0].ProductVariantId)
	assert.True(t, resolver.query.MaterializeIfMissing)
	assert.Equal(t, map[string]string{}, resolver.query.Selections)
	assert.Equal(t, "01ORG", repo.filter["org_id"])
}

func TestResolveTemplateFormsRefusesATemplateOfAnotherOrg(t *testing.T) {
	writer, _, _ := templateFormWriter(t, false, nil)

	cErrs, err := writer.resolveTemplateForms(callerContext(), templateFormDelta())

	require.NoError(t, err)
	assert.Equal(t, []string{"moves.add[1].product_template_id"}, errorFields(cErrs))
}

func TestResolveTemplateFormsRefusesNeverAttributes(t *testing.T) {
	writer, resolver, _ := templateFormWriter(t, true, &itProduct.ResolveCompleteVariantResult{HasData: true,
		Data: itProduct.ResolveCompleteVariantResultData{VariantId: "V9", Outcome: itProduct.ResolveOutcomeExisting,
			NeverSelections: []itProduct.AttributeSelection{{AttributeId: "A2", ValueId: "AV2"}}}})

	cErrs, err := writer.resolveTemplateForms(callerContext(), templateFormDelta())

	require.NoError(t, err)
	assert.Equal(t, []string{"moves.add[1].attributes[1]"}, errorFields(cErrs))
	assert.Equal(t, "inventory:stock_transfer.move_attribute_not_variant", cErrs[0].Key)
	assert.Equal(t, map[string]string{"A1": "AV1", "A2": "AV2"}, resolver.query.Selections)
}

func TestResolveTemplateFormsPointsResolverErrorsAtTheRow(t *testing.T) {
	refused := ft.ClientErrors{
		*resolveViolation(ReasonResolveTemplateArchived, "archived", nil),
		*selectionViolation(ReasonSelectionMissingAttr, "missing", map[string]any{"attribute_id": "A3"}),
	}
	writer, _, _ := templateFormWriter(t, true, &itProduct.ResolveCompleteVariantResult{ClientErrors: refused})

	cErrs, err := writer.resolveTemplateForms(callerContext(), templateFormDelta())

	require.NoError(t, err)
	assert.Equal(t, []string{"moves.add[1].product_template_id", "moves.add[1].attributes"}, errorFields(cErrs))
}
