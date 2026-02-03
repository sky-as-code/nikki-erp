package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

type HierarchyLevelDto struct {
	Id        string              `json:"id"`
	CreatedAt time.Time           `json:"createdAt"`
	Children  []HierarchyLevelDto `json:"children,omitempty"`
	Etag      string              `json:"etag"`
	Name      string              `json:"name"`
	Org       *GetGroupRespOrg    `json:"org,omitempty"`
	UpdatedAt *time.Time          `json:"updatedAt,omitempty"`
}

func (this *HierarchyLevelDto) FromHierarchyLevel(hierarchyLevel domain.HierarchyLevel) {
	model.MustCopy(hierarchyLevel.AuditableBase, this)
	model.MustCopy(hierarchyLevel.ModelBase, this)
	model.MustCopy(hierarchyLevel, this)
	if hierarchyLevel.Org != nil {
		this.Org = &GetGroupRespOrg{}
		this.Org.FromOrg(hierarchyLevel.Org)
	}

	if hierarchyLevel.Children != nil {
		this.Children = array.Map(hierarchyLevel.Children, func(child domain.HierarchyLevel) HierarchyLevelDto {
			childDto := HierarchyLevelDto{}
			childDto.FromHierarchyLevel(child)
			return childDto
		})
	}
}

// Request/Response DTOs
type CreateHierarchyLevelRequest = it.CreateHierarchyLevelCommand
type CreateHierarchyLevelResponse = httpserver.RestCreateResponse

type DeleteHierarchyLevelRequest = it.DeleteHierarchyLevelCommand
type DeleteHierarchyLevelResponse = httpserver.RestDeleteResponse

type UpdateHierarchyLevelRequest = it.UpdateHierarchyLevelCommand
type UpdateHierarchyLevelResponse = httpserver.RestUpdateResponse

type GetHierarchyLevelByIdRequest = it.GetHierarchyLevelByIdQuery
type GetHierarchyLevelByIdResponse = HierarchyLevelDto

type SearchHierarchyLevelsRequest = it.SearchHierarchyLevelsQuery

type SearchHierarchyLevelsResponse struct {
	Items []HierarchyLevelDto `json:"items"`
	Total int                 `json:"total"`
	Page  int                 `json:"page"`
	Size  int                 `json:"size"`
}

func (this *SearchHierarchyLevelsResponse) FromResult(result *it.SearchHierarchyLevelsResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(hierarchyLevel domain.HierarchyLevel) HierarchyLevelDto {
		item := HierarchyLevelDto{}
		item.FromHierarchyLevel(hierarchyLevel)
		return item
	})
}

type ManageUsersHierarchyRequest = it.AddRemoveUsersCommand
type ManageUsersHierarchyResponse = httpserver.RestUpdateResponse
