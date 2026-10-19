package rolerequest

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The composable layers of the iam_grant_request resource, served at /v1/iam/iam_grant_request.
// The schema is named "iam_grant_request", not "iam_role_request". It is plain CRUD: the request
// workflow (approve, reject, cancel) is not built yet.

type RoleRequestCrudRepository interface {
	composable.CrudRepository
}

type RoleRequestCrudDomainService interface {
	composable.CrudDomainService
}

type RoleRequestCrudApplicationService interface {
	composable.CrudApplicationService
}
