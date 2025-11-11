package inventory

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attribute"
	"github.com/sky-as-code/nikki-erp/modules/inventory/attributegroup"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit"
	"github.com/sky-as-code/nikki-erp/modules/inventory/variant"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
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
	return nil
}

// Version implements NikkiModule.
func (*InventoryModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*InventoryModule) Init() error {
	err := errors.Join(
		deps.Register(newInventoryClient),
		deps.Invoke(unit.InitSubModule),
		deps.Invoke(variant.InitSubModule),
		deps.Invoke(product.InitSubModule),
		deps.Invoke(attribute.InitSubModule),
		deps.Invoke(attributegroup.InitSubModule),
		deps.Invoke(attribute.InitSubModule),
		// deps.Invoke(attributevalue.InitSubModule),
		// deps.Invoke(unit.InitSubModule),
		// deps.Invoke(unitcategory.InitSubModule),
	)

	return err
}

func newInventoryClient(clientOpts *db.EntClientOptions) *ent.Client {
	var client *ent.Client
	if clientOpts.DebugEnabled {
		client = ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	client = ent.NewClient(ent.Driver(clientOpts.Driver))

	err := client.DB().Ping()
	if err != nil {
		panic(err)
	}

	return client
}
