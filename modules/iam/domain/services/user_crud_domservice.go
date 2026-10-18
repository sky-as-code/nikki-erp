package services

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

// NewUserCrudDomainService is handed the composable default by the iam_user onion. The resource is
// plain CRUD; the type exists so a rule has a home when one is needed.
func NewUserCrudDomainService(base composable.CrudDomainService) itUser.UserCrudDomainService {
	return &UserCrudDomainServiceImpl{CrudDomainService: base}
}

type UserCrudDomainServiceImpl struct {
	composable.CrudDomainService
}
