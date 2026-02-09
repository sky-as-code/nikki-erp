package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/orm"
)

func InitRepositories() error {
	err := stdErr.Join(
		orm.RegisterEntity(BuildAttributeDescriptor()),
		orm.RegisterEntity(BuildAttributeGroupDescriptor()),
		orm.RegisterEntity(BuildAttributeValueDescriptor()),
		orm.RegisterEntity(BuildProductDescriptor()),
		orm.RegisterEntity(BuildVariantDescriptor()),
		// orm.RegisterEntity(BuildProductCategoryDescriptor()), // TODO: Implement when ready
	)
	if err != nil {
		return err
	}

	err = stdErr.Join(
		deps.Register(NewAttributeEntRepository),
		deps.Register(NewAttributeGroupEntRepository),
		deps.Register(NewAttributeValueEntRepository),
		deps.Register(NewProductEntRepository),
		deps.Register(NewVariantEntRepository),
		// deps.Register(NewProductCategoryEntRepository), // TODO: Implement when ready
	)

	return err
}
