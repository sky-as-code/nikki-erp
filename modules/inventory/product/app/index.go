package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewAttributeServiceImpl),
		deps.Register(NewAttributeGroupServiceImpl),
		deps.Register(NewAttributeValueServiceImpl),
		deps.Register(NewProductServiceImpl),
		deps.Register(NewVariantServiceImpl),
		// deps.Register(NewProductCategoryServiceImpl), // TODO: Implement when ready
	)
	return err
}
