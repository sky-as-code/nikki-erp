package v1

import (
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

type CreateUserRequest = it.CreateUserCommand

type CreateUserResponse struct {
	Id        string `json:"id,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	Etag      string `json:"etag,omitempty"`
	Status    string `json:"status,omitempty"`
}

func (this *CreateUserResponse) FromUser(user domain.User) {
	this.Id = user.Id.String()
	this.CreatedAt = user.CreatedAt.Unix()
	this.Etag = user.Etag.String()
	this.Status = user.Status.String()
}

type UpdateUserRequest = it.UpdateUserCommand

type UpdateUserResponse struct {
	Id        string `json:"id,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
	Etag      string `json:"etag,omitempty"`
}

func (this *UpdateUserResponse) FromUser(user domain.User) {
	this.Id = user.Id.String()
	this.UpdatedAt = user.UpdatedAt.Unix()
	this.Etag = user.Etag.String()
}

type DeleteUserRequest = it.DeleteUserCommand

type DeleteUserResponse struct {
	DeletedAt int64 `json:"deletedAt,omitempty"`
}

type GetUserByIdRequest = it.GetUserByIdQuery

type GetUserByIdResponse struct {
	Id          string `json:"id,omitempty"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
	Etag        string `json:"etag,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	AvatarUrl   string `json:"avatarUrl,omitempty"`
	Status      string `json:"status,omitempty"`
}

func (this *GetUserByIdResponse) FromUser(user domain.User) {
	this.Id = user.Id.String()
	this.CreatedAt = user.CreatedAt.Unix()
	this.Etag = user.Etag.String()
	// this.DisplayName = *user.DisplayName
	// this.Email = *user.Email
	// this.AvatarUrl = *user.AvatarUrl
	this.Status = user.Status.String()
}
