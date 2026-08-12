package dynamicengines

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// AC-PROD-019 and AC-PROD-020 hinge on which variants a template archive touches, and on
// whether the last archive leaves the template selectable. Both decisions are pure functions of
// the record's own state, so they are tested directly rather than through a live engine.

// BR-PROD-VAR-002 and AC-PROD-012: a template must not hold two variants with the same
// combination. The record being updated must not be mistaken for a conflict with itself.
func TestAssertUniqueCombination(t *testing.T) {
	const (
		templateId = "01TEMPLATE0000000000000000"
		selfId     = "01VARIANTSELF0000000000000"
		otherId    = "01VARIANTOTHER000000000000"
	)

	testCases := []struct {
		name       string
		existing   []string
		selfId     string
		wantErrKey string
	}{
		{"a free combination is accepted", nil, "", ""},
		{"a combination held by another variant is rejected", []string{otherId}, "",
			"product_variant.duplicate_combination"},
		{"a variant keeping its own combination is accepted", []string{selfId}, selfId, ""},
		{"a variant taking another's combination is rejected", []string{otherId}, selfId,
			"product_variant.duplicate_combination"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &fakeSearcher{}
			for _, id := range testCase.existing {
				repo.items = append(repo.items, dmodel.DynamicFields{
					models.ProductVariantFieldId: id,
				})
			}

			variant := models.NewProductVariant()
			variant.SetProductTemplateId(ptr(templateId))
			variant.SetCombinationKey(ptr("12:101|18:205"))

			vErrs := &ft.ClientErrors{}
			require.NoError(t, checkUniqueCombination(testContext(t), repo, variant, testCase.selfId, vErrs))

			assertViolation(t, vErrs, testCase.wantErrKey)
		})
	}
}

// BR §4.5 and AC-PROD-008: a template with no variant-generating attributes still gets one
// variant, whose combination is the empty string. An empty key is a real key, so it must be
// checked for uniqueness like any other rather than skipped as "missing".
func TestEmptyCombinationIsStillChecked(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{
		{models.ProductVariantFieldId: "01OTHER00000000000000000000"},
	}}

	variant := models.NewProductVariant()
	variant.SetProductTemplateId(ptr("01TEMPLATE0000000000000000"))
	variant.SetCombinationKey(ptr(models.EmptyCombinationKey))

	vErrs := &ft.ClientErrors{}
	require.NoError(t, checkUniqueCombination(testContext(t), repo, variant, "", vErrs))

	assertViolation(t, vErrs, "product_variant.duplicate_combination")
}

// A variant whose template or combination has not been supplied is the schema's problem, not the
// uniqueness rule's: it must not run a search on partial keys.
func TestUniqueCombinationSkipsIncompleteRecord(t *testing.T) {
	repo := &fakeSearcher{items: []dmodel.DynamicFields{
		{models.ProductVariantFieldId: "01OTHER00000000000000000000"},
	}}

	variant := models.NewProductVariant()
	variant.SetProductTemplateId(ptr("01TEMPLATE0000000000000000"))

	vErrs := &ft.ClientErrors{}
	require.NoError(t, checkUniqueCombination(testContext(t), repo, variant, "", vErrs))

	assert.Zero(t, vErrs.Count())
	assert.Zero(t, repo.calls, "must not search without a combination key")
}

// AC-PROD-018: a discontinued variant is not archived, and an archived one is not discontinued.
// Either state alone withdraws it from new business.
func TestVariantSelectability(t *testing.T) {
	testCases := []struct {
		name     string
		archived bool
		status   models.ProductVariantStatus
		want     bool
	}{
		{"active and unarchived is selectable", false, models.ProductVariantStatusActive, true},
		{"discontinued is not selectable", false, models.ProductVariantStatusDiscontinued, false},
		{"archived is not selectable", true, models.ProductVariantStatusActive, false},
		{"both is not selectable", true, models.ProductVariantStatusDiscontinued, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			variant := models.NewProductVariant()
			variant.SetIsArchived(&testCase.archived)
			variant.SetStatus(&testCase.status)

			assert.Equal(t, testCase.want, variant.IsSelectable())
		})
	}
}

func TestMergeVariantForValidation(t *testing.T) {
	stored := models.NewProductVariant()
	stored.SetProductTemplateId(ptr("01TEMPLATE0000000000000000"))
	stored.SetCombinationKey(ptr("old"))
	stored.SetSku(ptr("SKU-1"))

	submitted := models.NewProductVariant()
	submitted.SetCombinationKey(ptr("new"))

	merged := mergeVariantForValidation(submitted, stored)

	// The submitted field wins, and the untouched ones survive: an update is partial, so the
	// rules must see the record the write will produce.
	assert.Equal(t, "new", *merged.GetCombinationKey())
	assert.Equal(t, "SKU-1", *merged.GetSku())
	assert.Equal(t, "01TEMPLATE0000000000000000", *merged.GetProductTemplateId())
}

func ptr[T any](v T) *T {
	return &v
}

func assertViolation(t *testing.T, vErrs *ft.ClientErrors, wantKey string) {
	t.Helper()

	if wantKey == "" {
		assert.Zerof(t, vErrs.Count(), "expected no violation, got %v", vErrs)
		return
	}

	require.NotZero(t, vErrs.Count(), "expected violation %q", wantKey)
	found := false
	for _, item := range *vErrs {
		if item.Key == wantKey {
			found = true
		}
	}
	assert.Truef(t, found, "expected violation %q, got %v", wantKey, *vErrs)
}

// testContext is the request context the repository slice under test never actually reads; the
// fake searcher ignores it entirely.
func testContext(t *testing.T) corectx.Context {
	t.Helper()
	return corectx.NewRequestContext(t.Context())
}

// fakeSearcher returns a fixed row set, so the uniqueness rule can be exercised without a
// database.
type fakeSearcher struct {
	items []dmodel.DynamicFields
	calls int
}

func (this *fakeSearcher) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.calls++
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.items},
		HasData: len(this.items) > 0,
	}, nil
}
