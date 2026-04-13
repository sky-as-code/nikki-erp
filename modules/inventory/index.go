package inventory

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product"
	productDomain "github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit"
	unitDomain "github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

var ModuleSingleton modules.InCodeModule = &InventoryModule{}

type InventoryModule struct {
}

// LabelKey implements NikkiModule.
func (*InventoryModule) LabelKey() string {
	return "inventory.moduleLabel"
}

// Name implements NikkiModule.
func (*InventoryModule) Name() string {
	return "inventory"
}

// Deps implements NikkiModule.
func (*InventoryModule) Deps() []string {
	return []string{}
}

// Version implements NikkiModule.
func (*InventoryModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*InventoryModule) Init() error {
	err := errors.Join(
		unit.Init(),
		product.Init(),
	)

	return err
}

// RegisterModels implements DynamicModule.
func (*InventoryModule) RegisterModels() error {
	return errors.Join(
		// Product schemas
		dmodel.RegisterSchemaB(productDomain.ProductCategoryRelSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.ProductCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.ProductSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.AttributeGroupSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.AttributeSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.AttributeValueSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.VariantSchemaBuilder()),
		dmodel.RegisterSchemaB(productDomain.VariantAttrValRelSchemaBuilder()),
		// Unit schemas
		dmodel.RegisterSchemaB(unitDomain.UnitCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(unitDomain.UnitSchemaBuilder()),
	)
}
