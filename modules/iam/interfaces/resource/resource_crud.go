package resource

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_resource resource, served at /v1/iam/iam_resource.

type ResourceCrudRepository interface {
	dynamicresource.CrudRepository
}

// ResourceCrudDomainService is the CRUD of the resource plus its rule: min_scope never exceeds
// max_scope.
type ResourceCrudDomainService interface {
	dynamicresource.CrudDomainService
}

type ResourceCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}
