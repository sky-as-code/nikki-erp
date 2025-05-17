package user

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateUserCommand)(nil)
	req = (*UpdateUserCommand)(nil)
	req = (*DeleteUserCommand)(nil)
	req = (*GetUserByIdQuery)(nil)
	req = (*GetUserByUsernameQuery)(nil)
	util.Unused(req)
}

var createUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	CreatedBy          string `json:"createdBy"`
	DisplayName        string `json:"displayName"`
	Email              string `json:"email"`
	IsEnabled          bool   `json:"isEnabled"`
	MustChangePassword bool   `json:"mustChangePassword"`
	Password           string `json:"password"`
}

func (CreateUserCommand) Type() cqrs.RequestType {
	return createUserCommandType
}

type CreateUserResult model.OpResult[*domain.User]

var updateUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	Id                 string  `param:"id" query:"id" json:"id"`
	AvatarUrl          *string `json:"avatarUrl,omitempty"`
	DisplayName        *string `json:"displayName,omitempty"`
	Email              *string `json:"email,omitempty"`
	MustChangePassword *bool   `json:"mustChangePassword,omitempty"`
	Password           *string `json:"password,omitempty"`
	IsEnabled          *bool   `json:"isEnabled,omitempty"`
	UpdatedBy          string  `json:"updatedBy,omitempty"`
	Username           *string `json:"username,omitempty"`
}

func (UpdateUserCommand) Type() cqrs.RequestType {
	return updateUserCommandType
}

type UpdateUserResult model.OpResult[*domain.User]

var deleteUserCommandType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "delete",
}

type DeleteUserCommand struct {
	Id        string `json:"id"`
	DeletedBy string `json:"deleted_by"`
}

func (DeleteUserCommand) Type() cqrs.RequestType {
	return deleteUserCommandType
}

var getUserByIdQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserById",
}

type GetUserByIdQuery struct {
	Id string `json:"id"`
}

func (GetUserByIdQuery) Type() cqrs.RequestType {
	return getUserByIdQueryType
}

var getUserByUsernameQueryType = cqrs.RequestType{
	Module:    "identity",
	Submodule: "user",
	Action:    "getUserByUsername",
}

type GetUserByUsernameQuery struct {
	Username string `json:"username"`
}

func (GetUserByUsernameQuery) Type() cqrs.RequestType {
	return getUserByUsernameQueryType
}
