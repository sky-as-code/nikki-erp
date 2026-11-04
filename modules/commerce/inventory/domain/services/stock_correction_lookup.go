package services

import (
	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
)

// The lookups a correction resolves before it writes: the virtual counterparty location and the
// operation type the generated transfer is recorded under.

// FindLocationByType resolves an org's virtual counterparty location of a given type. Inventory
// adjustments balance against `inventory_loss` and scraps move to `scrap`, both seeded per org.
// The first match wins; two matching locations is a configuration problem, not a reason to refuse.
func FindLocationByType(
	ctx corectx.Context, orgId string, locationType string,
) (*models.InventoryLocation, error) {
	engine, err := repoFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	// The org's virtual locations: platform-managed, outside any warehouse, one per usage.
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldLocationType, dmodel.Equals, locationType),
		*dmodel.NewSearchNode().NewCondition(
			basemodel.FieldManagedByType, dmodel.Equals, basemodel.ManagedByTypeSystem),
		*dmodel.NewSearchNode().NewCondition(models.InventoryLocationFieldWarehouseId, dmodel.IsNotSet),
	)

	found, err := engine.Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  1,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindLocationByType")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return models.NewInventoryLocationFrom(found.Data.Items[0]), nil
}

// resolveMovementOperationType picks the operation type a generated movement is recorded under:
// the one the caller named, else the seeded correction type.
func resolveMovementOperationType(
	ctx corectx.Context, request CorrectionRequest,
) (*models.StockOperationType, error) {
	if request.OperationTypeId == "" {
		return findCorrectionOperationType(ctx, request.OrgId)
	}
	engine, err := repoFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{
		models.StockOperationTypeFieldId: request.OperationTypeId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "resolveMovementOperationType")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return models.NewStockOperationTypeFrom(found.Data), nil
}

// correctionSettingsFor resolves the settings port the correction type is read from. It is a
// variable, like repoFor, so a test can substitute it: the container is filled during Init.
var correctionSettingsFor = func() (itExt.EffectiveSettingsExtService, error) {
	var settings itExt.EffectiveSettingsExtService
	err := deps.Invoke(func(svc itExt.EffectiveSettingsExtService) { settings = svc })
	return settings, err
}

// findCorrectionOperationType resolves the operation type the org's correction_operation_type_id
// setting names. An unset setting, or one naming another org's type, finds nothing.
func findCorrectionOperationType(
	ctx corectx.Context, orgId string,
) (*models.StockOperationType, error) {
	typeId, err := correctionOperationTypeId(ctx)
	if err != nil || typeId == "" {
		return nil, err
	}
	engine, err := repoFor(models.StockOperationTypeSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.FindByKeys(ctx, dmodel.DynamicFields{models.StockOperationTypeFieldId: typeId})
	if err != nil {
		return nil, errors.Wrap(err, "findCorrectionOperationType")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	operationType := models.NewStockOperationTypeFrom(found.Data)
	if orgOf := operationType.GetOrgId(); orgOf == nil || string(*orgOf) != orgId {
		return nil, nil
	}
	return operationType, nil
}

func correctionOperationTypeId(ctx corectx.Context) (string, error) {
	settings, err := correctionSettingsFor()
	if err != nil {
		return "", errors.Wrap(err, "correctionOperationTypeId")
	}
	result, err := settings.GetEffectiveSettings(ctx, itExt.GetEffectiveSettingsQuery{
		ModuleKeys: []string{commconstants.CommerceModuleName},
	})
	if err != nil {
		return "", errors.Wrap(err, "correctionOperationTypeId")
	}
	if result == nil || !result.HasData || result.ClientErrors.Count() > 0 {
		return "", nil
	}
	key := commconstants.CommerceModuleName + "." + models.OrgSettingCorrectionOperationTypeId
	typeId, _ := result.Data.Values[key].(string)
	return typeId, nil
}
