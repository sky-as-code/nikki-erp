package user

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
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
	Module:    "core",
	Submodule: "user",
	Action:    "create",
}

type CreateUserCommand struct {
	CreatedBy          string `json:"created_by"`
	DisplayName        string `json:"display_name"`
	Email              string `json:"email"`
	IsEnabled          bool   `json:"is_enabled"`
	MustChangePassword bool   `json:"must_change_password"`
	Password           string `json:"password"`
	Username           string `json:"username"`
}

func (CreateUserCommand) Type() cqrs.RequestType {
	return createUserCommandType
}

type CreateUserResult struct {
	Errors []string `json:"errors"`

	Id        *string    `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Etag      *string    `json:"etag,omitempty"`
	Status    *string    `json:"status,omitempty"`
}

var updateUserCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "user",
	Action:    "update",
}

type UpdateUserCommand struct {
	Id                  string  `json:"id"`
	AvatarUrl           *string `json:"avatar_url,omitempty"`
	CreatedAt           string  `json:"created_at"`
	CreatedBy           string  `json:"created_by"`
	DeletedAt           *string `json:"deleted_at,omitempty"`
	DeletedBy           *string `json:"deleted_by,omitempty"`
	DisplayName         *string `json:"display_name,omitempty"`
	Email               *string `json:"email,omitempty"`
	FailedLoginAttempts *int    `json:"failed_login_attempts,omitempty"`
	LastLoginAt         *string `json:"last_login_at,omitempty"`
	LockedUntil         *string `json:"locked_until,omitempty"`
	MustChangePassword  *bool   `json:"must_change_password,omitempty"`
	PasswordChangedAt   *string `json:"password_changed_at,omitempty"`
	PasswordHash        *string `json:"password_hash,omitempty"`
	Status              *string `json:"status,omitempty"`
	UpdatedAt           *string `json:"updated_at,omitempty"`
	UpdatedBy           *string `json:"updated_by,omitempty"`
	Username            *string `json:"username,omitempty"`
}

func (UpdateUserCommand) Type() cqrs.RequestType {
	return updateUserCommandType
}

var deleteUserCommandType = cqrs.RequestType{
	Module:    "core",
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
	Module:    "core",
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
	Module:    "core",
	Submodule: "user",
	Action:    "getUserByUsername",
}

type GetUserByUsernameQuery struct {
	Username string `json:"username"`
}

func (GetUserByUsernameQuery) Type() cqrs.RequestType {
	return getUserByUsernameQueryType
}
