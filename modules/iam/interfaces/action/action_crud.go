package action

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_action resource, served at /v1/iam/iam_action. An action is no
// longer a sub-resource nested under its resource: a client filters by resource_id.

type ActionCrudRepository interface {
	dynamicresource.CrudRepository
}

// ActionCrudDomainService is the CRUD of the resource plus its rule: the resource it belongs to
// must exist.
type ActionCrudDomainService interface {
	dynamicresource.CrudDomainService
}

type ActionCrudApplicationService interface {
	dynamicresource.CrudApplicationService
}
