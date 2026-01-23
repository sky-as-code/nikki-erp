package action

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateActionCommand) ToDomainModel() *domain.Action {
	action := &domain.Action{}
	model.MustCopy(this, action)

	return action
}

func (this UpdateActionCommand) ToDomainModel() *domain.Action {
	action := &domain.Action{}
	model.MustCopy(this, action)
	return action
}

func (this DeleteActionHardByIdCommand) ToDomainModel() *domain.Action {
	action := &domain.Action{}
	model.MustCopy(this, action)
	return action
}
