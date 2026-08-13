package inventory

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/inventory/dynamicengines"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
	"github.com/sky-as-code/nikki-erp/modules/inventory/transport/restful"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
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
// The steps are ordered, and each depends on the one before it, so they run in sequence rather
// than being joined: the engines must exist before a service can be derived from one, the
// derived service must be installed before the actions that type-assert it are reachable, and
// the REST layer registers the engines' routes last.
//
// The superseded Products implementation under ./product is deliberately not initialized: it is
// kept as a historical folder pending manual deletion. See ./product/README.md.
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
	return restful.InitRestfulHandlers()
}

// initStockTransferService installs the derived transfer service on the Stock Transfer engine.
//
// The six movement actions type-assert the engine's service to the derived type, so without this
// every one of them fails at the assertion rather than at a request: confirm, reserve, validate
// and the rest live on that type and nowhere else.
func initStockTransferService() error {
	transferEngine, ok := dynamicresource.Registry().GetEngine(models.StockTransferSchemaName)
	if !ok {
		return errors.New("the '" + models.StockTransferSchemaName + "' engine is not registered")
	}

	transferEngine.SetResourceService(services.NewStockTransferDomainService(transferEngine.ResourceService()))
	return nil
}

// initStockQuantService installs the derived quant service on the Stock Quant engine.
//
// It is what fills available_quantity on a read: the field has no database column, so without
// this the engine would advertise it in meta/schema and serve it as permanently null.
func initStockQuantService() error {
	quantEngine, ok := dynamicresource.Registry().GetEngine(models.StockQuantSchemaName)
	if !ok {
		return errors.New("the '" + models.StockQuantSchemaName + "' engine is not registered")
	}

	quantEngine.SetResourceService(services.NewStockQuantDomainService(quantEngine.ResourceService()))
	return nil
}

// initStockScrapService installs the derived scrap service on the Stock Scrap engine.
//
// Two things depend on it: Do Scrap type-asserts to the derived type, and the create/update/delete
// overrides are what stop a done scrap being edited or deleted. Without this the document rules
// would silently not apply — the CRUD would still work, which is what makes the omission easy to
// miss.
func initStockScrapService() error {
	scrapEngine, ok := dynamicresource.Registry().GetEngine(models.StockScrapSchemaName)
	if !ok {
		return errors.New("the '" + models.StockScrapSchemaName + "' engine is not registered")
	}

	scrapEngine.SetResourceService(services.NewStockScrapDomainService(scrapEngine.ResourceService()))
	return nil
}

// initProductService installs the derived Products service on the Product Template engine.
//
// The engine is created with the default resource service; this replaces it with one that embeds
// the default and adds the Products capabilities, so that built-in CRUD is untouched while a
// custom action can reach the extra methods through ProcessInput.ResourceService.
//
// SetResourceService is a plain field assignment with no locking, so it is safe here — during
// Init, before any request is served — and nowhere else.
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

// initProductVariantService installs the derived variant service on the Product Variant engine.
//
// It is what fills the template_* virtual fields on a read: they have no database column, so
// without this the engine would serve them as permanently absent.
func initProductVariantService() error {
	variantEngine, ok := dynamicresource.Registry().GetEngine(models.ProductVariantSchemaName)
	if !ok {
		return errors.New("the '" + models.ProductVariantSchemaName + "' engine is not registered")
	}

	derived := services.NewProductVariantDomainService(variantEngine.ResourceService())
	variantEngine.SetResourceService(derived)

	// Published for consumers that reach these reads outside an engine action. The same instance
	// serves all three ports, so a consumer gets the batched template_* fill whichever it injects.
	return errors.Join(
		deps.Register(func() itProduct.ProductVariantDomainService { return derived }),
		deps.Register(func() itProduct.ProductTemplateReadService { return derived }),
		deps.Register(func() itProduct.ProductCategoryReadService { return derived }),
	)
}

// RegisterModels implements DynamicModule.
//
// Schemas must be registered referenced-before-referencing, because an edge is resolved against
// the schema registry at registration time.
func (*InventoryModule) RegisterModels() error {
	return errors.Join(
		// Master data: referenced by the template, so registered first.
		dmodel.RegisterSchemaB(models.ProductTypeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(models.BrandSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ProductPriceSchemaBuilder()),

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

		// Stock. The location is referenced by both of the others, and the quant references the
		// variant registered above, so this order is the only one that resolves.
		dmodel.RegisterSchemaB(models.StockLocationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockOperationTypeSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockQuantSchemaBuilder()),

		// Movement. The transfer references the operation type and locations above; the move
		// references the transfer, and the line and the dependency both reference the move, so
		// this order is the only one that resolves.
		dmodel.RegisterSchemaB(models.StockTransferSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveLineSchemaBuilder()),
		dmodel.RegisterSchemaB(models.StockMoveDependencySchemaBuilder()),

		// Corrections. The scrap references the transfer, the variant and two locations, all
		// registered above, so it comes last.
		dmodel.RegisterSchemaB(models.StockScrapSchemaBuilder()),
	)
}
