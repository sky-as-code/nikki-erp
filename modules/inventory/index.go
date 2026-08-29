package inventory

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/inventory/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/inventory/dynamicengines"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/inventory/transport/restful"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
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
	return modconstants.InventoryModuleName
}

// Deps implements NikkiModule.
func (*InventoryModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
	}
}

// IsInternal implements InCodeModule.
func (*InventoryModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*InventoryModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
//
// The steps must run in this order: engines exist before a service is derived from one, a derived
// service is installed before the actions that type-assert it, and REST registers routes last.
//
// The superseded ./product implementation is deliberately not initialized; see ./product/README.md.
func (*InventoryModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := initProductService(); err != nil {
		return err
	}
	if err := initStockQuantService(); err != nil {
		return err
	}
	if err := initStockTransferService(); err != nil {
		return err
	}
	if err := initStockScrapService(); err != nil {
		return err
	}
	// Must follow the quant service, which answers the location lifecycle guards.
	if err := initWarehouseServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// initWarehouseServices installs the warehouse and location services and the layer above them.
//
// Order is load-bearing: the location service needs the quant service's usage port, and the
// application service composes both services created here.
func initWarehouseServices() error {
	warehouseEngine, ok := dynamicresource.Registry().GetEngine(models.WarehouseSchemaName)
	if !ok {
		return errors.New("the '" + models.WarehouseSchemaName + "' engine is not registered")
	}
	locationEngine, ok := dynamicresource.Registry().GetEngine(models.InventoryLocationSchemaName)
	if !ok {
		return errors.New("the '" + models.InventoryLocationSchemaName + "' engine is not registered")
	}

	var usageReader itStock.LocationUsageReadService
	if err := deps.Invoke(func(reader itStock.LocationUsageReadService) { usageReader = reader }); err != nil {
		return errors.Join(errors.New("the stock usage reader is not registered"), err)
	}

	warehouseSvc := services.NewWarehouseDomainService(warehouseEngine.ResourceService())
	warehouseEngine.SetResourceService(warehouseSvc)

	locationSvc := services.NewInventoryLocationDomainService(locationEngine.ResourceService(), usageReader)
	locationEngine.SetResourceService(locationSvc)

	if err := installDerivedService(models.StorageCategorySchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewStorageCategoryDomainService(base)
		}); err != nil {
		return err
	}
	if err := installDerivedService(models.WarehouseSupplyRelationSchemaName,
		func(base drif.DynamicResourceService) drif.DynamicResourceService {
			return services.NewSupplyRelationDomainService(base)
		}); err != nil {
		return err
	}

	return app.InitApplicationServices(warehouseSvc, locationSvc)
}

// installDerivedService replaces one engine's resource service with a derived one. The warehouse
// and location services are built by hand above instead, because the layer over them needs the
// concrete types.
func installDerivedService(
	schemaName string, derive func(drif.DynamicResourceService) drif.DynamicResourceService,
) error {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return errors.New("the '" + schemaName + "' engine is not registered")
	}
	engine.SetResourceService(derive(engine.ResourceService()))
	return nil
}

// initStockTransferService installs the derived transfer service on the Stock Transfer engine.
// The movement actions (confirm, reserve, validate, ...) type-assert the engine's service to the
// derived type, so without this each one fails at the assertion.
func initStockTransferService() error {
	transferEngine, ok := dynamicresource.Registry().GetEngine(models.StockTransferSchemaName)
	if !ok {
		return errors.New("the '" + models.StockTransferSchemaName + "' engine is not registered")
	}

	derived := services.NewStockTransferDomainService(transferEngine.ResourceService())
	transferEngine.SetResourceService(derived)

	// Published for consumers that sequence a movement outside an engine action, from their own
	// transaction boundaries. The narrowed interface is published, never the struct: handing over
	// the embedded CRUD would make the lifecycle rules optional.
	return deps.Register(func() itStock.StockTransferMovementService { return derived })
}

// initStockQuantService installs the derived quant service on the Stock Quant engine. It fills
// available_quantity on a read: the field has no database column, so without this the engine
// advertises it in meta/schema and serves it as permanently null.
func initStockQuantService() error {
	quantEngine, ok := dynamicresource.Registry().GetEngine(models.StockQuantSchemaName)
	if !ok {
		return errors.New("the '" + models.StockQuantSchemaName + "' engine is not registered")
	}

	derived := services.NewStockQuantDomainService(quantEngine.ResourceService())
	quantEngine.SetResourceService(derived)

	// The same instance answers what Stock holds at a location, consulted before a location is
	// suspended or archived. Publishing it as a port keeps the dependency one-way: the warehouse
	// services read this contract and never a stock table.
	return deps.Register(func() itStock.LocationUsageReadService { return derived })
}

// initStockScrapService installs the derived scrap service on the Stock Scrap engine. Do Scrap
// type-asserts to the derived type, and the create/update/delete overrides stop a done scrap being
// edited or deleted. Without this the CRUD still works but the document rules silently do not apply.
func initStockScrapService() error {
	scrapEngine, ok := dynamicresource.Registry().GetEngine(models.StockScrapSchemaName)
	if !ok {
		return errors.New("the '" + models.StockScrapSchemaName + "' engine is not registered")
	}

	scrapEngine.SetResourceService(services.NewStockScrapDomainService(scrapEngine.ResourceService()))
	return nil
}

// initProductService installs the derived Products service on the Product Template engine. The
// replacement embeds the default service, so built-in CRUD is untouched while custom actions reach
// the extra methods through ProcessInput.ResourceService.
//
// SetResourceService is an unlocked field assignment, so it is safe only during Init, before any
// request is served.
func initProductService() error {
	templateEngine, ok := dynamicresource.Registry().GetEngine(models.ProductTemplateSchemaName)
	if !ok {
		return errors.New("the '" + models.ProductTemplateSchemaName + "' engine is not registered")
	}

	derived := services.NewProductTemplateDomainService(templateEngine.ResourceService())
	templateEngine.SetResourceService(derived)

	if err := initProductVariantService(); err != nil {
		return err
	}

	// Published for consumers that reach the capability outside an engine action.
	return deps.Register(func() itProduct.ProductService { return derived })
}

// initProductVariantService installs the derived variant service on the Product Variant engine. It
// fills the template_* virtual fields on a read: they have no database column, so without this the
// engine serves them as permanently absent.
func initProductVariantService() error {
	variantEngine, ok := dynamicresource.Registry().GetEngine(models.ProductVariantSchemaName)
	if !ok {
		return errors.New("the '" + models.ProductVariantSchemaName + "' engine is not registered")
	}

	derived := services.NewProductVariantDomainService(variantEngine.ResourceService())
	variantEngine.SetResourceService(derived)

	// One instance serves all four ports, so a consumer gets the batched template_* fill whichever
	// it injects. The pricing-basis port stays separate because it grants strictly less: a price
	// calculator gets the pricing inputs without a general product reader.
	return errors.Join(
		deps.Register(func() itProduct.ProductVariantDomainService { return derived }),
		deps.Register(func() itProduct.ProductTemplateReadService { return derived }),
		deps.Register(func() itProduct.ProductCategoryReadService { return derived }),
		deps.Register(func() itProduct.ProductPricingBasisService { return derived }),
	)
}

// RegisterModels implements DynamicModule.
//
// Schemas must be registered referenced-before-referencing: an edge is resolved against the schema
// registry at registration time.
func (*InventoryModule) RegisterModels() error {
	return errors.Join(
		// Master data: referenced by the template, so registered first.
		dmodel.RegisterSchemaB(models.ProductTypeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(models.BrandSchemaBuilder()),

		// Attributes: the value points at the attribute, so the attribute comes first.
		dmodel.RegisterSchemaB(models.ProductAttributeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductAttributeValueSchemaBuilder()),

		// The template references type, category and brand, all registered above.
		dmodel.RegisterSchemaB(models.ProductTemplateSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductTemplateAttributeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductTemplateAttributeValueSchemaBuilder()),

		// The variant references the template; its value junction references both the variant
		// and the template-scoped allowed value.
		dmodel.RegisterSchemaB(models.ProductVariantSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductVariantAttributeValueSchemaBuilder()),

		// Warehouse topology. Both are referenced by a location, so they precede it.
		dmodel.RegisterSchemaB(models.WarehouseSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StorageCategorySchemaBuilder()),

		// The shared location master, owned by neither Warehouse nor Stock. Both stock schemas below
		// reference it, and the quant also references the variant registered above.
		dmodel.RegisterSchemaB(models.InventoryLocationSchemaBuilder()),

		// Warehouse configuration that points at both a warehouse and a location, so it comes
		// after each of them.
		dmodel.RegisterSchemaB(models.WarehouseSupplyRelationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PutawayRuleSchemaBuilder()),

		// Stock.
		dmodel.RegisterSchemaB(models.StockOperationTypeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockQuantSchemaBuilder()),

		// Movement. The transfer references the operation type and locations above; the move
		// references the transfer; the line and the dependency reference the move.
		dmodel.RegisterSchemaB(models.StockTransferSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveDependencySchemaBuilder()),

		// Corrections. The scrap references the transfer, the variant and two locations above.
		dmodel.RegisterSchemaB(models.StockScrapSchemaBuilder()),

		// Stock's per-product settings; references the product template. The UoM it names lives in
		// Essential and is held as a plain id, so nothing from that module must be registered first.
		dmodel.RegisterSchemaB(models.StockProductConfigSchemaBuilder()),
	)
}
