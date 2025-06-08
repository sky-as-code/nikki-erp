package group

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (cmd CreateGroupCommand) ToGroup() *domain.Group {
	return &domain.Group{
		Name:        &cmd.Name,
		Description: cmd.Description,
		OrgId:       cmd.OrgId,
	}
}

func (cmd UpdateGroupCommand) ToGroup() *domain.Group {
	return &domain.Group{
		Name:        cmd.Name,
		Description: cmd.Description,
		OrgId:       cmd.OrgId,

		ModelBase: model.ModelBase{
			Id:   &cmd.Id,
			Etag: &cmd.Etag,
		},
	}
}
