package rolerequest

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_grant_request resource, served at /v1/iam/iam_grant_request.
// The schema is named "iam_grant_request", not "iam_role_request". It is plain CRUD: the request
// workflow (approve, reject, cancel) is not built yet.

type RoleRequestCrudRepository interface {
	dynamicresource.CrudRepository
}

type RoleRequestCrudDomainService interface {
	dynamicresource.CrudDomainService
}

type RoleRequestCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}
