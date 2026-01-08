package v1

import "github.com/sky-as-code/nikki-erp/common/model"

type UserSummaryDto struct {
	Id          *model.Id `json:"id"`
	DisplayName *string   `json:"name"`
}

func (this *UserSummaryDto) FromUserSummary(id model.Id, displayName *string) {
	this.Id = &id
	this.DisplayName = displayName
}

// Target (Role/Suite) on Grant_Request
type TargetSummaryDto struct {
	Id   *model.Id `json:"id"`
	Name *string   `json:"name"`
}

func (this *TargetSummaryDto) FromTargetSummary(id model.Id, name *string) {
	this.Id = &id
	this.Name = name
}

// Organization summary
type OrganizationSummaryDto struct {
	Id          *model.Id `json:"id"`
	DisplayName *string   `json:"name"`
}

func (this *OrganizationSummaryDto) FromOrganization(orgId *model.Id, displayName *string) {
	this.Id = orgId
	this.DisplayName = displayName
}
