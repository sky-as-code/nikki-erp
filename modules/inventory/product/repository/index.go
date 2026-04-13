package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/inventory/infra/ent"
)

func InitRepositories() error {
	err := stdErr.Join(
		deps.Register(newInventoryClient),
		deps.Register(NewProductDynamicRepository),
		deps.Register(NewProductCategoryDynamicRepository),
		deps.Register(NewAttributeDynamicRepository),
		deps.Register(NewAttributeGroupDynamicRepository),
		deps.Register(NewAttributeValueDynamicRepository),
		deps.Register(NewVariantDynamicRepository),
	)

	return err
}

func newInventoryClient(clientOpts *db.EntClientOptions) *ent.Client {
	if clientOpts.DebugEnabled {
		return ent.NewClient(ent.Driver(clientOpts.Driver), ent.Debug())
	}
	return ent.NewClient(ent.Driver(clientOpts.Driver))
}
