package app

import (
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

func InitServices() error {
	err := errors.Join(
		deps.Register(NewContactServiceImpl),
		deps.Register(NewFieldMetadataServiceImpl),
		deps.Register(NewLanguageServiceImpl),
		deps.Register(NewModelMetadataServiceImpl),
		deps.Register(NewModuleServiceImpl),
		deps.Register(NewUnitServiceImpl),
		deps.Register(NewUnitCategoryServiceImpl),
	)
	return err
}
