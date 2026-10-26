package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewAttributeService),
		deps.Register(NewAttributeGroupService),
		deps.Register(NewAttributeValueService),
		deps.Register(NewProductService),
		deps.Register(NewProductCategoryServiceImpl),
		deps.Register(NewVariantService),
	)
	return err
}
