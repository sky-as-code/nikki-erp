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
	return restful.InitRestfulHandlers()
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
	)
}
