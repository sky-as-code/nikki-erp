package service

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The composable layers of the iam_service resource, served at /v1/iam/iam_service. A service is
// a row of iam_users of type service: the repository is the user repository viewed through the
// iam_service schema, and the domain service forces the type.

type ServiceCrudRepository interface {
	dynamicresource.CrudRepository
}

// ServiceCrudDomainService is the CRUD of the resource: every read and write is confined to type
// service, and a role assignment rebuilds the service's permission cache.
type ServiceCrudDomainService interface {
	dynamicresource.CrudDomainService

	ManageRoles(ctx corectx.Context, cmd ManageServiceRolesCommand) (*ManageServiceRolesResult, error)
}

type ServiceCrudApplicationService interface {
	dynamicresource.CrudApplicationService

	ManageRoles(ctx corectx.Context, cmd ManageServiceRolesCommand) (*ManageServiceRolesResult, error)
}

type (
	ManageServiceRolesCommand = dynamicresource.ManageM2mCommand
	ManageServiceRolesResult  = dynamicresource.MutateResult
)
