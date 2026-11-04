package app

import (
	"go.bryk.io/pkg/errors"
	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/commerce/infra/inventory/defaults"
	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/warehouse"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// aliasWarehouseStockLoc names the main warehouse's stock location, where its goods live, for the
// operation types' default source and destination.
const aliasWarehouseStockLoc = "location_main"

// operationTypeSettings maps each Commerce setting that names an operation type to the alias of the
// default type it points at for a new org.
var operationTypeSettings = []struct{ setting, alias string }{
	{models.OrgSettingOutgoingOperationTypeId, "stock_operation_type_main_out"},
	{models.OrgSettingIncomingOperationTypeId, "stock_operation_type_main_in"},
	{models.OrgSettingCorrectionOperationTypeId, "stock_operation_type_correction"},
}

// cleanUpOrder lists every org-scoped resource children first. Attribute links without an
// org_id go with the templates and variants they cascade from. Warehouses go before locations:
// warehouse.root_location_id holds a RESTRICT FK to its root, while location.warehouse_id is SET
// NULL when its warehouse goes.
var cleanUpOrder = []string{
	models.StockMoveDependencySchemaName, models.StockMoveLineSchemaName, models.StockMoveSchemaName,
	models.StockQuantSchemaName, models.StockReservationSchemaName, models.StockScrapSchemaName,
	models.ProductVariantSchemaName, models.StockProductConfigSchemaName, models.ProductTemplateSchemaName,
	models.BrandSchemaName, models.InventoryIntegrationOutboxSchemaName, models.PutawayRuleSchemaName,
	models.StockTransferSchemaName, models.StockOperationTypeSchemaName,
	models.WarehouseProductGuardSchemaName, models.WarehouseSupplyRelationSchemaName, models.WarehouseSchemaName,
	models.InventoryLocationSchemaName,
	models.ProductAttributeValueSchemaName, models.ProductAttributeSchemaName,
	models.ProductCategorySchemaName, models.StorageCategorySchemaName,
}

type OrgLifecycleParam struct {
	dig.In

	StorageCategory itWarehouse.StorageCategoryDomainService
	Category        itProduct.ProductCategoryDomainService
	Warehouse       itWarehouse.WarehouseAppService
	Location        itWarehouse.InventoryLocationDomainService
	OperationType   itStock.StockOperationTypeDomainService
	Settings        itExt.OrgSettingsSeedExtService
	Units           *services.OrgUnitResolver
	Sequences       itExt.SequenceExtService
}

func NewOrgLifecycleAppService(param OrgLifecycleParam) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{param: param, cleanup: dynamicresource.NewOrgDataCleanup(cleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl gives a new org the root category, a main warehouse with the
// locations its flows need, the virtual locations every move starts or ends in, and the
// operation types sales and purchasing use.
type OrgLifecycleAppServiceImpl struct {
	param   OrgLifecycleParam
	cleanup dynamicresource.OrgDataCleanup
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	return dynamicresource.InTransaction(ctx, this.param.Location.Repository(), func(tranxCtx corectx.Context) error {
		aliases := defaults.NewAliases()
		steps := []func(corectx.Context, *defaults.Aliases, model.Id) error{
			this.createStorageCategories, this.createRootCategory, this.createWarehouse,
			this.createVirtualLocations, this.createOperationTypes, this.seedOperationTypeSettings,
			this.seedUnitSettings, this.registerSequences,
		}
		for _, step := range steps {
			if err := step(tranxCtx, aliases, orgId); err != nil {
				return err
			}
		}
		return nil
	})
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	return this.cleanup.CleanUpOrg(ctx, orgId)
}

func (this *OrgLifecycleAppServiceImpl) createStorageCategories(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	return createFromDefaults(ctx, this.param.StorageCategory, aliases, orgId)
}

func (this *OrgLifecycleAppServiceImpl) createRootCategory(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.ProductCategorySchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		result, err := this.param.Category.CreateRoot(ctx, row)
		return createdId(result, err, models.ProductCategorySchemaName)
	})
}

// createWarehouse goes through the warehouse's own orchestration, which provisions the root, Stock
// and flow locations; the Stock's id is remembered for the operation types.
func (this *OrgLifecycleAppServiceImpl) createWarehouse(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.WarehouseSchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		result, err := this.param.Warehouse.CreateWarehouse(ctx, itWarehouse.CreateWarehouseCommand{Fields: row})
		if err != nil {
			return "", err
		}
		if result.ClientErrors.Count() > 0 {
			return "", result.ClientErrors.ToError()
		}
		warehouse := models.NewWarehouseFrom(result.Data.Warehouse)
		if warehouse.GetStockLocationId() == nil {
			return "", errors.Errorf("warehouse %s has no stock location", derefString(warehouse.GetCode()))
		}
		aliases.Remember(aliasWarehouseStockLoc, *warehouse.GetStockLocationId())
		return dynamicresource.ParamId(result.Data.Warehouse), nil
	})
}

// createVirtualLocations writes the org-wide locations outside any warehouse. They are managed by
// the system (their defaults say so) and system-generated, a flag the location service strips
// from input.
func (this *OrgLifecycleAppServiceImpl) createVirtualLocations(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, models.InventoryLocationSchemaName)
	if err != nil {
		return err
	}
	return aliases.CreateAll(rows, orgId, func(row dmodel.DynamicFields) (model.Id, error) {
		result, err := this.param.Location.CreateManaged(ctx, row)
		if err != nil {
			return "", err
		}
		if result.ClientErrors.Count() > 0 {
			return "", result.ClientErrors.ToError()
		}
		id := dynamicresource.ParamId(result.Data)
		return id, services.MarkLocationSystemGenerated(ctx, string(id))
	})
}

// registerSequences registers the org's transfer and scrap numbering and records their ids in the
// Inventory settings. It is the only place an Inventory definition is registered: operation types
// share the transfer definition, each on its own counter, so creating one registers nothing.
func (this *OrgLifecycleAppServiceImpl) registerSequences(
	ctx corectx.Context, _ *defaults.Aliases, orgId model.Id,
) error {
	_, err := itExt.RegisterOrgSequences(ctx, this.param.Sequences, this.param.Settings,
		invconstants.InventoryModuleName, orgId, services.OrgSequenceRegistrations(orgId)...)
	return err
}

func (this *OrgLifecycleAppServiceImpl) createOperationTypes(
	ctx corectx.Context, aliases *defaults.Aliases, orgId model.Id,
) error {
	return createFromDefaults(ctx, this.param.OperationType, aliases, orgId)
}

// seedOperationTypeSettings points the Commerce settings at the operation types just created. A
// module that adds its own types for the same purpose (Vending Machine's kiosk dispense) overwrites
// its setting from its own lifecycle, which runs in a later wave.
func (this *OrgLifecycleAppServiceImpl) seedOperationTypeSettings(
	ctx corectx.Context, aliases *defaults.Aliases, _ model.Id,
) error {
	items := make([]itExt.SetSettingItem, 0, len(operationTypeSettings))
	for _, entry := range operationTypeSettings {
		id, found := aliases.Id(entry.alias)
		if !found {
			return errors.Errorf("default operation type %s was not created", entry.alias)
		}
		items = append(items, itExt.SetSettingItem{Name: entry.setting, Value: string(id)})
	}
	result, err := this.param.Settings.SeedOrgSettings(ctx, itExt.SetSettingsCommand{
		ModuleKey: commconstants.CommerceModuleName,
		Items:     items,
	})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "seed operation type settings")
	}
	return nil
}

// seedUnitSettings points the inventory unit settings at the organization's own kg and cm, which
// Essential's lifecycle created in an earlier wave. A UoM id differs per organization, so the
// schema can carry no default_value and the defaults are resolved by symbol here.
func (this *OrgLifecycleAppServiceImpl) seedUnitSettings(
	ctx corectx.Context, _ *defaults.Aliases, orgId model.Id,
) error {
	ids, err := this.param.Units.DefaultUnitIds(ctx, string(orgId))
	if err != nil || len(ids) == 0 {
		return err
	}
	items := make([]itExt.SetSettingItem, 0, len(ids))
	for name, id := range ids {
		items = append(items, itExt.SetSettingItem{Name: name, Value: id})
	}
	result, err := this.param.Settings.SeedOrgSettings(ctx, itExt.SetSettingsCommand{
		ModuleKey: invconstants.InventoryModuleName,
		Items:     items,
	})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "seed unit settings")
	}
	return nil
}

func createFromDefaults(
	ctx corectx.Context, svc dynamicresource.CrudDomainService, aliases *defaults.Aliases, orgId model.Id,
) error {
	rows, err := defaults.Load(orgDefaults.Files, svc.Schema().Name())
	if err != nil {
		return err
	}
	return dynamicresource.CreateDefaults(ctx, svc, rows, aliases, orgId)
}

func createdId(result *itProduct.CreateProductCategoryResult, err error, schemaName string) (model.Id, error) {
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 {
		return "", result.ClientErrors.ToError()
	}
	if !result.HasData {
		return "", errors.Errorf("create %s returned no record", schemaName)
	}
	return dynamicresource.ParamId(result.Data), nil
}
