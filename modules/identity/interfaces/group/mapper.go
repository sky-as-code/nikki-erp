package group

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (cmd CreateGroupCommand) ToGroup() *domain.Group {
	group := &domain.Group{}
	model.MustCopy(cmd, group)
	return group
}

func (cmd UpdateGroupCommand) ToGroup() *domain.Group {
	group := &domain.Group{}
	model.MustCopy(cmd, group)
	return group
}
