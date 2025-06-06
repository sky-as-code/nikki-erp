package group

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func (cmd CreateGroupCommand) ToGroup() *domain.Group {
	return &domain.Group{
		Name:        cmd.Name,
		Description: cmd.Description,
		OrgId:       cmd.OrgId,

		AuditableBase: model.AuditableBase{
			CreatedBy: model.WrapId(cmd.CreatedBy),
		},
	}
}

func (cmd UpdateGroupCommand) ToGroup() *domain.Group {
	return &domain.Group{
		Name:        cmd.Name,
		Description: cmd.Description,
		OrgId:       cmd.OrgId,

		AuditableBase: model.AuditableBase{
			UpdatedBy: model.WrapId(cmd.UpdatedBy),
		},

		ModelBase: model.ModelBase{
			Id:   model.WrapId(cmd.Id),
			Etag: model.WrapEtag(cmd.Etag),
		},
	}
}
