package services

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

// Whether a parent row may be hard-deleted: the foreign key already refuses it, so the guards
// exist to say so as a business violation instead of letting the driver error escape as a 500.

// stubChildSearcher answers every search with the same rows, which is all the guards read: they
// ask whether any child exists, not which.
type stubChildSearcher struct {
	rows []dmodel.DynamicFields
}

func (this *stubChildSearcher) Search(
	_ corectx.Context, _ dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.rows},
		HasData: len(this.rows) > 0,
	}, nil
}

func oneChild() *stubChildSearcher {
	return &stubChildSearcher{rows: []dmodel.DynamicFields{{"id": "child-1"}}}
}

func noChild() *stubChildSearcher {
	return &stubChildSearcher{}
}

const testParentId = "parent-1"

func TestBrandWithTemplatesCannotBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertBrandDeletable(callerContext(), oneChild(), testParentId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "a template still carries the brand")
}

func TestBrandWithoutTemplatesCanBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertBrandDeletable(callerContext(), noChild(), testParentId, vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

// An empty id is a malformed request, not a business refusal: schema validation reports it, and
// searching for children of "" would match nothing anyway.
func TestBrandGuardIgnoresEmptyId(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertBrandDeletable(callerContext(), oneChild(), "", vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductTypeWithTemplatesCannotBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductTypeDeletable(callerContext(), oneChild(), testParentId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "a template is still of this product type")
}

func TestProductTypeWithoutTemplatesCanBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductTypeDeletable(callerContext(), noChild(), testParentId, vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductTypeGuardIgnoresEmptyId(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductTypeDeletable(callerContext(), oneChild(), "", vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductAttributeWithValuesCannotBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductAttributeDeletable(callerContext(), oneChild(), testParentId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "the attribute still owns a value")
}

func TestProductAttributeWithoutValuesCanBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductAttributeDeletable(callerContext(), noChild(), testParentId, vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductAttributeGuardIgnoresEmptyId(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductAttributeDeletable(callerContext(), oneChild(), "", vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductCategoryWithTemplatesCannotBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductCategoryDeletable(
		callerContext(), oneChild(), noChild(), testParentId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "a product still sits in this category")
}

func TestProductCategoryWithChildCategoriesCannotBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductCategoryDeletable(
		callerContext(), noChild(), oneChild(), testParentId, vErrs))

	assert.Equal(t, 1, vErrs.Count(), "the self-referencing parent key blocks it too")
}

// Both kinds of child are reported at once, so a user clearing them is not sent back for a second
// round after removing only the products.
func TestProductCategoryReportsEveryBlockingChild(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductCategoryDeletable(
		callerContext(), oneChild(), oneChild(), testParentId, vErrs))

	assert.Equal(t, 2, vErrs.Count())
}

func TestProductCategoryWithoutChildrenCanBeDeleted(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductCategoryDeletable(
		callerContext(), noChild(), noChild(), testParentId, vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

func TestProductCategoryGuardIgnoresEmptyId(t *testing.T) {
	vErrs := ft.NewClientErrors()

	require.NoError(t, AssertProductCategoryDeletable(
		callerContext(), oneChild(), oneChild(), "", vErrs))

	assert.Equal(t, 0, vErrs.Count())
}

var _ models.ProductSearcher = (*stubChildSearcher)(nil)
