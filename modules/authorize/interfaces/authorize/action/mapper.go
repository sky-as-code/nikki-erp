package action

import (
	"github.com/sky-as-code/nikki-erp/common/model"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

func (this CreateActionCommand) ToAction() *domain.Action {
	action := &domain.Action{}
	model.MustCopy(this, action)

	return action
}

func (this UpdateActionCommand) ToAction() *domain.Action {
	action := &domain.Action{}
	model.MustCopy(this, action)
	return action
}
