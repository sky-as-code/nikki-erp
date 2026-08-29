package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// NewStorageCategoryDomainService derives the storage category service from the engine's default.
func NewStorageCategoryDomainService(base drif.DynamicResourceService) *StorageCategoryDomainServiceImpl {
	return &StorageCategoryDomainServiceImpl{DynamicResourceService: base}
}

// StorageCategoryDomainServiceImpl guards archiving a category something still uses: withdrawing
// one while a live location points at it leaves that location citing a policy nobody can look up.
type StorageCategoryDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*StorageCategoryDomainServiceImpl)(nil)

// SetArchived refuses to archive a category an unarchived location still uses. Historical use does
// not count: an archived location that once carried it is no reason to keep it forever.
func (this *StorageCategoryDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if !readBoolParam(params, paramIsArchived) {
		return this.DynamicResourceService.SetArchived(ctx, params)
	}

	categoryId := readStringParam(params, models.StorageCategoryFieldId)
	inUse, err := countLocationsUsingCategory(ctx, categoryId)
	if err != nil {
		return nil, err
	}
	if inUse > 0 {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation(models.StorageCategorySchemaName,
			"storage_category.in_use",
			"locations still use this storage category; reassign them first"))
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.DynamicResourceService.SetArchived(ctx, params)
}

// countLocationsUsingCategory reports how many live locations carry a storage category.
func countLocationsUsingCategory(ctx corectx.Context, categoryId string) (int, error) {
	if categoryId == "" {
		return 0, nil
	}

	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return 0, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldStorageCategoryId, dmodel.Equals, categoryId),
	)

	total, err := countMatching(ctx, engine, graph)
	return total, errors.Wrap(err, "countLocationsUsingCategory")
}
