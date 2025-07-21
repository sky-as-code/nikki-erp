package v1

import (
	"github.com/thoas/go-funk"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

type RoleSuiteDto struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name                 string   `json:"name"`
	Description          *string  `json:"description,omitempty"`
	OwnerType            string   `json:"ownerType"`
	OwnerRef             string   `json:"ownerRef"`
	IsRequestable        *bool    `json:"isRequestable,omitempty"`
	IsRequiredAttachment *bool    `json:"isRequiredAttachment,omitempty"`
	IsRequiredComment    *bool    `json:"isRequiredComment,omitempty"`
	CreatedBy            model.Id `json:"createdBy"`

	Roles []*RoleDto `json:"roles,omitempty"`
}

type CreateRoleSuiteRequest = it.CreateRoleSuiteCommand
type CreateRoleSuiteResponse = httpserver.RestCreateResponse

type GetRoleSuiteByIdRequest = it.GetRoleSuiteByIdQuery
type GetRoleSuiteByIdResponse = RoleSuiteDto

type SearchRoleSuitesRequest = it.SearchRoleSuitesCommand
type SearchRoleSuitesResponse httpserver.RestSearchResponse[RoleSuiteDto]

func (this *SearchRoleSuitesResponse) FromResult(result *it.SearchRoleSuitesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = funk.Map(result.Items, func(roleSuite domain.RoleSuite) RoleSuiteDto {
		item := RoleSuiteDto{}
		item.FromRoleSuite(roleSuite)
		return item
	}).([]RoleSuiteDto)
}

func (this *GetRoleSuiteByIdResponse) FromRoleSuite(roleSuite domain.RoleSuite) {
	this.Id = *roleSuite.Id
	this.Name = *roleSuite.Name
	this.Description = roleSuite.Description
	this.Etag = *roleSuite.Etag
	this.OwnerType = roleSuite.OwnerType.String()
	this.OwnerRef = *roleSuite.OwnerRef
	this.IsRequestable = roleSuite.IsRequestable
	this.IsRequiredAttachment = roleSuite.IsRequiredAttachment
	this.IsRequiredComment = roleSuite.IsRequiredComment
	this.CreatedBy = *roleSuite.CreatedBy

	this.Roles = funk.Map(roleSuite.Roles, func(role *domain.Role) *RoleDto {
		item := RoleDto{}
		item.FromRole(*role)
		return &item
	}).([]*RoleDto)
}
