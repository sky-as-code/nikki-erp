package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	return stdErr.Join(
		deps.Register(NewProductDynamicRepository),
		deps.Register(NewProductCategoryDynamicRepository),
		deps.Register(NewAttributeDynamicRepository),
		deps.Register(NewAttributeGroupDynamicRepository),
		deps.Register(NewAttributeValueDynamicRepository),
		deps.Register(NewVariantDynamicRepository),
	)
}
