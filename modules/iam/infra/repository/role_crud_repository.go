package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itRole "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/role"
)

// NewRoleCrudRepository is handed the composable default by the iam_role onion.
func NewRoleCrudRepository(base composable.CrudRepository) itRole.RoleCrudRepository {
	return &RoleCrudRepositoryImpl{CrudRepository: base}
}

type RoleCrudRepositoryImpl struct {
	composable.CrudRepository
}
