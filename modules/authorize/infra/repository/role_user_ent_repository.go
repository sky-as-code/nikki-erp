package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"

	entRoleUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/roleuser"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleUserEntRepository(client *ent.Client) itRole.RoleUserRepository {
	return &RoleUserEntRepository{
		client: client,
	}
}

type RoleUserEntRepository struct {
	client *ent.Client
}

func BuildRoleUserDescriptor() *orm.EntityDescriptor {
	entity := ent.RoleUser{}
	builder := orm.DescribeEntity(entRoleUser.Label).
		Aliases("role_users").
		Field(entRoleUser.FieldID, entity.ID).
		Field(entRoleUser.FieldApproverID, entity.ApproverID).
		Field(entRoleUser.FieldReceiverRef, entity.ReceiverRef).
		Field(entRoleUser.FieldReceiverType, entity.ReceiverType).
		Field(entRoleUser.FieldRoleID, entity.RoleID)

	return builder.Descriptor()
}
