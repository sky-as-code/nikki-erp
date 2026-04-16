package repository

import (
	stdErr "errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitRepositories() error {
	return stdErr.Join(
		deps.Register(NewContactDynamicRepository),
		deps.Register(NewFieldMetadataDynamicRepository),
		deps.Register(NewLanguageDynamicRepository),
		deps.Register(NewModelMetadataDynamicRepository),
		deps.Register(NewModuleDynamicRepository),
		deps.Register(NewUnitDynamicRepository),
		deps.Register(NewUnitCategoryDynamicRepository),
	)
}
