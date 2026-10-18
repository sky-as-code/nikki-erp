package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewUserCrudRepository is handed the composable default by the iam_user onion. The type exists so
// a bespoke query has a home; it adds nothing today.
func NewUserCrudRepository(base composable.CrudRepository) itUser.UserCrudRepository {
	return &UserCrudRepositoryImpl{CrudRepository: base}
}

type UserCrudRepositoryImpl struct {
	composable.CrudRepository
}
