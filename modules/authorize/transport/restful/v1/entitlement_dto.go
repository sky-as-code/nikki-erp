package v1

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement"
)

type CreateEntitlementRequest = it.CreateEntitlementCommand
type CreateEntitlementResponse = GetEntitlementByIdResponse

type GetEntitlementByIdResponse struct {
	Id   model.Id   `json:"id"`
	Etag model.Etag `json:"etag"`

	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	ActionId    *model.Id `json:"actionId,omitempty"`
	ActionExpr  *string   `json:"actionExpr,omitempty"`
	SubjectType string    `json:"subjectType"`
	SubjectRef  *model.Id `json:"subjectRef,omitempty"`
	ScopeRef    *model.Id `json:"scopeRef,omitempty"`
	ResourceId  *model.Id `json:"resourceId,omitempty"`
	CreatedBy   model.Id  `json:"createdBy"`
}

func (this *GetEntitlementByIdResponse) FromEntitlement(entitlement domain.Entitlement) {
	this.Id = *entitlement.Id
	this.Name = *entitlement.Name
	this.Description = entitlement.Description
	this.ResourceId = entitlement.ResourceId
	this.Etag = *entitlement.Etag
	this.CreatedBy = *entitlement.CreatedBy
	this.ActionId = entitlement.ActionId
	this.ActionExpr = entitlement.ActionExpr
	this.SubjectType = string(*entitlement.SubjectType)
	this.SubjectRef = entitlement.SubjectRef
	this.ScopeRef = entitlement.ScopeRef
}
