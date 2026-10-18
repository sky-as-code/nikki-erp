package user

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable/filefield"
)

// The composable layers of the iam_user resource, served at /v1/iam/iam_user. The "Crud" infix
// keeps them apart from the hand-written UserRepository/UserDomainService/UserAppService, which
// still back login, permission evaluation, /me/context and the command bus.

// UserCrudRepository reads and writes user rows for the composable resource.
type UserCrudRepository interface {
	composable.CrudRepository
}

// UserCrudDomainService is the CRUD of the resource plus its own rules.
type UserCrudDomainService interface {
	composable.CrudDomainService
}

// UserCrudApplicationService is the authorized surface the REST handler serves.
type UserCrudApplicationService interface {
	composable.CrudApplicationService
	filefield.FileFieldsService

	SendInvitation(ctx corectx.Context, cmd SendUserInvitationCommand) (*SendUserInvitationResult, error)
}

type (
	SendUserInvitationCommand = dmodel.DynamicFields
	SendUserInvitationResult  = dyn.OpResult[any]
)
