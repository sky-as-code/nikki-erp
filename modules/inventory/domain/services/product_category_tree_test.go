package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func canonicalLangJson(pairs ...string) *model.LangJson {
	value := model.LangJson{}
	for i := 0; i+1 < len(pairs); i += 2 {
		value[model.LanguageCode(pairs[i])] = pairs[i+1]
	}
	return &value
}

func TestBuildCanonicalNameOfTheRootIsItsName(t *testing.T) {
	root := models.ProductCategoryRootName()

	canonical := BuildCanonicalName(nil, &root)

	assert.Equal(t, model.LangJson{"en-US": "All", "vi-VN": "Tất cả"}, canonical)
}

func TestBuildCanonicalNameAppendsPerLanguage(t *testing.T) {
	parent := canonicalLangJson("en-US", "All / Beverages", "vi-VN", "Tất cả / Đồ uống")

	canonical := BuildCanonicalName(parent, canonicalLangJson("en-US", "Soft drinks", "vi-VN", "Nước ngọt"))

	assert.Equal(t, model.LangJson{
		"en-US": "All / Beverages / Soft drinks",
		"vi-VN": "Tất cả / Đồ uống / Nước ngọt",
	}, canonical)
}

func TestBuildCanonicalNameFallsBackToEnUsOnEitherSide(t *testing.T) {
	parent := canonicalLangJson("en-US", "All / Beverages", "vi-VN", "Tất cả / Đồ uống")

	canonical := BuildCanonicalName(parent, canonicalLangJson("en-US", "Cola", "ja-JP", "コーラ"))

	assert.Equal(t, model.LangJson{
		"en-US": "All / Beverages / Cola",
		"ja-JP": "All / Beverages / コーラ",
		"vi-VN": "Tất cả / Đồ uống / Cola",
	}, canonical)
}

func TestBuildCanonicalNameFallsBackToTheFirstLanguageWithoutEnUs(t *testing.T) {
	canonical := BuildCanonicalName(canonicalLangJson("vi-VN", "Tất cả"), canonicalLangJson("ja-JP", "飲料"))

	assert.Equal(t, model.LangJson{"ja-JP": "Tất cả / 飲料", "vi-VN": "Tất cả / 飲料"}, canonical)
}

func TestBuildCanonicalNameUsesARefKeyOnlyAsALastResort(t *testing.T) {
	canonical := BuildCanonicalName(canonicalLangJson("en-US", "All"), canonicalLangJson("$ref", "category.drinks"))

	assert.Equal(t, model.LangJson{"en-US": "All / category.drinks"}, canonical)
}

// rootRepository answers the root search with rootId, or with nothing when rootId is empty.
type rootRepository struct {
	composable.CrudRepository
	rootId string
}

func (this *rootRepository) Search(_ corectx.Context, _ dyn.RepoSearchParam) (*composable.SearchResult, error) {
	if this.rootId == "" {
		return &composable.SearchResult{}, nil
	}
	return &composable.SearchResult{HasData: true, Data: dyn.PagedResultData[dmodel.DynamicFields]{
		Items: []dmodel.DynamicFields{{models.ProductCategoryFieldId: this.rootId}},
	}}, nil
}

type rootBaseService struct {
	composable.CrudDomainService
	repo *rootRepository
}

func (this *rootBaseService) Repository() composable.CrudRepository { return this.repo }

func newRootService(rootId string) *ProductCategoryDomainServiceImpl {
	return &ProductCategoryDomainServiceImpl{
		CrudDomainService: &rootBaseService{repo: &rootRepository{rootId: rootId}},
	}
}

func TestFillParentWithRootDefaultsAMissingParent(t *testing.T) {
	service := newRootService("root_1")
	entity := categoryEntity(dmodel.DynamicFields{"name": map[string]any{"en-US": "X"}, "org_id": "org_1"})
	vErrs := ft.NewClientErrors()

	result, err := service.fillParentWithRoot(corectx.NewRequestContext(context.Background()), entity, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 0, vErrs.Count())
	assert.Equal(t, "root_1", derefString(models.NewProductCategoryFrom(result.GetFieldData()).GetParentCategoryId()))
}

func TestFillParentWithRootKeepsAGivenParent(t *testing.T) {
	service := newRootService("root_1")
	entity := categoryEntity(dmodel.DynamicFields{"parent_category_id": "cat_9", "org_id": "org_1"})

	result, err := service.fillParentWithRoot(corectx.NewRequestContext(context.Background()), entity, ft.NewClientErrors())

	require.NoError(t, err)
	assert.Equal(t, "cat_9", derefString(models.NewProductCategoryFrom(result.GetFieldData()).GetParentCategoryId()))
}

func TestFillParentWithRootReportsAnOrgWithoutRoot(t *testing.T) {
	service := newRootService("")
	entity := categoryEntity(dmodel.DynamicFields{"org_id": "org_2"})
	vErrs := ft.NewClientErrors()

	_, err := service.fillParentWithRoot(corectx.NewRequestContext(context.Background()), entity, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 1, vErrs.Count())
}

func TestRejectServerManagedFields(t *testing.T) {
	vErrs := ft.NewClientErrors()

	rejectServerManagedFields(dmodel.DynamicFields{"canonical_name": "x", "is_root": true, "name": "y"}, vErrs)

	assert.Equal(t, 2, vErrs.Count())
}

func TestAssertRootUnchanged(t *testing.T) {
	root := models.NewProductCategoryFrom(dmodel.DynamicFields{
		"id": "root_1", "parent_category_id": "root_1", "is_root": true,
		"name": map[string]any{"en-US": "All", "vi-VN": "Tất cả"},
	})
	cases := map[string]struct {
		submitted  dmodel.DynamicFields
		violations int
	}{
		"same name and own parent": {dmodel.DynamicFields{
			"parent_category_id": "root_1", "name": map[string]any{"en-US": "All", "vi-VN": "Tất cả"},
		}, 0},
		"renamed": {dmodel.DynamicFields{"name": map[string]any{"en-US": "Everything"}}, 1},
		"moved":   {dmodel.DynamicFields{"parent_category_id": "cat_1"}, 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			vErrs := ft.NewClientErrors()
			assertRootUnchanged(models.NewProductCategoryFrom(tc.submitted), root, vErrs)
			assert.Equal(t, tc.violations, vErrs.Count())
		})
	}
}
