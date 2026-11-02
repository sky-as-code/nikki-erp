package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itEnum "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

// NewEnumRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewEnumRepository(base dynamicresource.CrudRepository) itEnum.EnumRepository {
	return &EnumRepositoryImpl{CrudRepository: base}
}

type EnumRepositoryImpl struct {
	dynamicresource.CrudRepository
}
