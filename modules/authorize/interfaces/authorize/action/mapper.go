package resource

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateActionCommand) ToAction() *domain.Action {
	return &domain.Action{
		Name:        &this.Name,
		Description: this.Description,
		ResourceId:  &this.ResourceId,
		CreatedBy:   &this.CreatedBy,
	}
}

func (this UpdateActionCommand) ToAction() *domain.Action {
	return &domain.Action{
		ModelBase: model.ModelBase{
			Id:   &this.Id,
			Etag: &this.Etag,
		},
		Description: this.Description,
	}
}
