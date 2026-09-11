package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// codeRepository answers GetOne from a set of taken codes, recording the org it was asked for.
type codeRepository struct {
	composable.CrudRepository
	taken    map[string]bool
	askedOrg string
}

func (this *codeRepository) GetOne(_ corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[dmodel.DynamicFields], error) {
	code, _ := param.Filter[models.ProductCategoryFieldCode].(string)
	this.askedOrg, _ = param.Filter[models.ProductCategoryFieldOrgId].(string)
	return &dyn.OpResult[dmodel.DynamicFields]{HasData: this.taken[code]}, nil
}

type codeBaseService struct {
	composable.CrudDomainService
	repo *codeRepository
}

func (this *codeBaseService) Repository() composable.CrudRepository { return this.repo }

func newCodeService(taken ...string) (*ProductCategoryDomainServiceImpl, *codeRepository) {
	repo := &codeRepository{taken: map[string]bool{}}
	for _, code := range taken {
		repo.taken[code] = true
	}
	return &ProductCategoryDomainServiceImpl{CrudDomainService: &codeBaseService{repo: repo}}, repo
}

func categoryEntity(fields dmodel.DynamicFields) *composable.DynamicEntity {
	return composable.NewDynamicEntityFrom(fields)
}

func TestFillCodeFromNameDerivesACodeWithinTheOrg(t *testing.T) {
	service, repo := newCodeService()
	entity := categoryEntity(dmodel.DynamicFields{
		"name": map[string]any{"vi-VN": "Nước ngọt", "en-US": "Soft Drinks"}, "org_id": "org_1",
	})

	result, err := service.fillCodeFromName(corectx.NewRequestContext(context.Background()), entity, ft.NewClientErrors())

	require.NoError(t, err)
	assert.Equal(t, "soft_drinks", *models.NewProductCategoryFrom(result.GetFieldData()).GetCode(), "en-US wins")
	assert.Equal(t, "org_1", repo.askedOrg)
}

func TestFillCodeFromNameSuffixesATakenCode(t *testing.T) {
	service, _ := newCodeService("nuoc_ngot", "nuoc_ngot_2")
	entity := categoryEntity(dmodel.DynamicFields{"name": map[string]any{"vi-VN": "Nước ngọt"}})

	result, err := service.fillCodeFromName(corectx.NewRequestContext(context.Background()), entity, ft.NewClientErrors())

	require.NoError(t, err)
	assert.Equal(t, "nuoc_ngot_3", *models.NewProductCategoryFrom(result.GetFieldData()).GetCode())
}

func TestFillCodeFromNameLeavesAGivenCodeAndANamelessCreateAlone(t *testing.T) {
	service, _ := newCodeService()
	ctx := corectx.NewRequestContext(context.Background())

	given := categoryEntity(dmodel.DynamicFields{"code": "custom", "name": map[string]any{"en-US": "X"}})
	result, err := service.fillCodeFromName(ctx, given, ft.NewClientErrors())
	require.NoError(t, err)
	assert.Equal(t, "custom", *models.NewProductCategoryFrom(result.GetFieldData()).GetCode())

	nameless := categoryEntity(dmodel.DynamicFields{})
	result, err = service.fillCodeFromName(ctx, nameless, ft.NewClientErrors())
	require.NoError(t, err)
	assert.Nil(t, models.NewProductCategoryFrom(result.GetFieldData()).GetCode())
}
