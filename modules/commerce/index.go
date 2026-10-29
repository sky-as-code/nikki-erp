// Package commerce is one module made of three parts: inventory, purchase and sales. Each part keeps
// its own schemas and prefix, REST group and translation namespace, and is built by its own
// InitSubModule, the way modules/core is built from its sub-modules. What the platform sees of a
// module - the loader, the dependency graph, usage checks, the org lifecycle, settings - is
// Commerce, once.
package commerce

import (
	stdErr "errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/infra"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales"
)

// ModuleSingleton is the symbol the plugin loader looks up. It is typed DynamicModule, not
// InCodeModule, so that dropping RegisterModels fails the build instead of silently registering
// no schemas.
var ModuleSingleton modules.DynamicModule = &CommerceModule{}

// Both hooks are found by a runtime type assertion, so these assertions turn a rename or a
// signature change into a compile error rather than a silently skipped hook.
var (
	_ modules.InCodeModuleAppStarted = &CommerceModule{}
	_ modules.InCodeModuleDependants = &CommerceModule{}
)

type CommerceModule struct{}

func (*CommerceModule) LabelKey() string {
	return "commerce.moduleLabel"
}

func (*CommerceModule) Name() string {
	return constants.CommerceModuleName
}

// ModelPrefixes implements DynamicModule: one prefix per part.
func (*CommerceModule) ModelPrefixes() []string {
	return constants.ModelPrefixes
}

func (*CommerceModule) HasUi() bool {
	return true
}

func (*CommerceModule) Category() modules.ModuleCategory {
	return modules.ModuleCategoryFoundation
}

// Deps is the union of the modules the parts read through a port. essential and core are injected
// by buildDependencyGraph anyway; essential is named because every part consumes it directly.
func (*CommerceModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		// A warehouse's manager_user edge holds a foreign key to iam_user, so IAM's tables come first.
		"iam",
		// Purchase reads the vendor party, Sales the customer.
		"contacts",
		// Sales registers and reads its settings.
		"settings",
		"paymentinvoice",
		// Sales' reprice action resolves accounting's tax port eagerly at Init, so accounting must
		// have registered its service before Commerce starts.
		"accounting",
	}
}

// Dependants implements InCodeModuleDependants: the modules referencing a product variant, template
// or stock quant, which Inventory asks before deleting one. Purchase and Sales reference them too,
// but they are parts of this module and answer in process (inventory/domain/services
// SetSameModuleParts), not over the bus.
//
// "vendingmachine" ships only in the coremart binary. It is listed anyway, because the list
// belongs to Commerce and states who references its products wherever it runs; the registry skips
// a dependant this binary did not load. See docs/problems/inventory/001.
func (*CommerceModule) Dependants() []string {
	return []string{
		"vendingmachine",
	}
}

func (*CommerceModule) IsInternal() bool {
	return false
}

func (*CommerceModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// RegisterModels registers every part's schemas. Inventory goes first: Purchase and Sales edges
// point at its products, and an edge is resolved against the registry at registration time.
func (*CommerceModule) RegisterModels() error {
	if err := inventory.RegisterModels(); err != nil {
		return err
	}
	return stdErr.Join(
		purchase.RegisterModels(),
		sales.RegisterModels(),
	)
}

// Init builds the parts in dependency order: Purchase and Sales consume Inventory's services, so
// Inventory's engines exist before theirs are built. The module-wide handlers subscribe last,
// because they compose what every part has registered by then.
func (*CommerceModule) Init() error {
	return runSteps(
		infra.InitSubModule,
		inventory.InitSubModule,
		purchase.InitSubModule,
		sales.InitSubModule,
		installSameModuleUsage,
		initCqrsHandlers,
	)
}

// OnAppStarted implements InCodeModuleAppStarted.
func (*CommerceModule) OnAppStarted() error {
	return runSteps(
		assertDependantsSubscribed,
		inventory.OnAppStarted,
		sales.OnAppStarted,
	)
}

// runSteps stops at the first failure: every step builds on the ones before it.
func runSteps(steps ...func() error) error {
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
