package resource

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateResourceCommand) ToResource() *domain.Resource {
	return &domain.Resource{
		Name:         &this.Name,
		Description:  this.Description,
		ResourceType: domain.WrapResourceType(this.ResourceType),
		ResourceRef:  &this.ResourceRef,
		ScopeType:    domain.WrapResourceScopeType(this.ScopeType),
		Actions:      []domain.Action{},
	}
}

func (this UpdateResourceCommand) ToResource() *domain.Resource {
	return &domain.Resource{
		ModelBase: model.ModelBase{
			Id:   &this.Id,
			Etag: &this.Etag,
		},
		Description: this.Description,
	}
}
