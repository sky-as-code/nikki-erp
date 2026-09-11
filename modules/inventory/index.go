package inventory

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/job"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/inventory/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/inventory/dynamicengines"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
	"github.com/sky-as-code/nikki-erp/modules/inventory/transport/restful"
)

var ModuleSingleton modules.InCodeModule = &InventoryModule{}

// OnAppStarted is found by a runtime type assertion, so this assertion is what turns a rename or a
// signature change into a compile error rather than a silently unregistered sweep.
var _ modules.InCodeModuleAppStarted = &InventoryModule{}

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

// ModelPrefix implements DynamicModule.
func (*InventoryModule) ModelPrefix() string {
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
// The onions are container constructors, so the layers they derive from one another (the location
// service needs the quant's usage port, the warehouse orchestration needs the location service)
// are ordered by the container rather than by this function. REST resolves every onion, which is
// what builds them all before the first request.
//
// The superseded ./legacy implementation is deliberately not initialized; see ./legacy/README.md.
func (*InventoryModule) Init() error {
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// OnAppStarted registers the reservation expiry sweep and checks that every resource the
// services may reach at call time was built. The sweep registers here rather than in Init so it
// never ticks against a half-built container.
func (*InventoryModule) OnAppStarted() error {
	if err := assertEveryResourceInstalled(); err != nil {
		return err
	}
	return deps.Invoke(func(
		transfers itStock.StockTransferMovementService,
		cronjobs job.CronjobRegistry,
		logger logging.LoggerService,
	) error {
		return app.NewReservationExpiryJobs(transfers, logger).RegisterJobs(cronjobs)
	})
}

// assertEveryResourceInstalled fails the boot when an onion was declared but never built: a
// service reaching it through the resource hub would otherwise fail on the first request that
// needs it, long after the boot looked healthy.
func assertEveryResourceInstalled() error {
	installed := map[string]bool{}
	for _, name := range services.InstalledResourceNames() {
		installed[name] = true
	}
	var missing []error
	for _, name := range dynamicengines.SchemaNames() {
		if !installed[name] {
			missing = append(missing, errors.New("the '"+name+"' inventory resource was declared but never built"))
		}
	}
	return errors.Join(missing...)
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
