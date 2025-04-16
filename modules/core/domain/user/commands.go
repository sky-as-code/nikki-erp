package user

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/sky-as-code/nikki-erp/common/cqrs"
	utility "github.com/sky-as-code/nikki-erp/common/util"
)

func init() {
	// Assert interface implementation
	var namer cqrs.Namer
	namer = (*CreateUserCommand)(nil)
	namer = (*UpdateUserCommand)(nil)
	namer = (*DeleteUserCommand)(nil)
	utility.Unused(namer)
}

type CreateUserCommand struct {
	CreatedBy          string
	DisplayName        string
	Email              string
	MustChangePassword bool
	Password           string
	Username           string
}

func (CreateUserCommand) Name() string {
	return "core_user.create"
}

type UpdateUserCommand struct {
	Id                 string
	AvatarUrl          string
	DisplayName        string
	MustChangePassword bool
	Status             string
	UpdatedBy          string
}

func (UpdateUserCommand) Name() string {
	return "core_user.update"
}

type DeleteUserCommand struct {
	DeletedBy string
	Id        string
}

func (DeleteUserCommand) Name() string {
	return "core_user.delete"
}

type GetUserByIdQuery struct {
	Id string
}

type GetUserByUsernameQuery struct {
	Username string `watermill:"command_id"`
}

type GetUserByEmailQuery struct {
	Email string `watermill:"command_id"`
}

func NewEventID() string {
	return watermill.NewUUID()
}
