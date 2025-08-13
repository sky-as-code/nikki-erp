package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
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

	Roles []RoleDto `json:"roles,omitempty"`
}

func (this *RoleSuiteDto) FromRoleSuite(roleSuite domain.RoleSuite) {
	model.MustCopy(roleSuite.AuditableBase, this)
	model.MustCopy(roleSuite.ModelBase, this)
	model.MustCopy(roleSuite, this)

	this.Roles = array.Map(roleSuite.Roles, func(role domain.Role) RoleDto {
		item := RoleDto{}
		item.FromRole(role)
		return item
	})
}

type CreateRoleSuiteRequest = it.CreateRoleSuiteCommand
type CreateRoleSuiteResponse = httpserver.RestCreateResponse

type UpdateRoleSuiteRequest = it.UpdateRoleSuiteCommand
type UpdateRoleSuiteResponse = httpserver.RestUpdateResponse

type DeleteRoleSuiteRequest = it.DeleteRoleSuiteCommand
type DeleteRoleSuiteResponse = httpserver.RestDeleteResponse

type GetRoleSuiteByIdRequest = it.GetRoleSuiteByIdQuery
type GetRoleSuiteByIdResponse = RoleSuiteDto

type SearchRoleSuitesRequest = it.SearchRoleSuitesCommand
type SearchRoleSuitesResponse httpserver.RestSearchResponse[RoleSuiteDto]

func (this *SearchRoleSuitesResponse) FromResult(result *it.SearchRoleSuitesResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(roleSuite domain.RoleSuite) RoleSuiteDto {
		item := RoleSuiteDto{}
		item.FromRoleSuite(roleSuite)
		return item
	})
}
