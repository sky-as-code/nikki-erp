package repository

import (
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"

	entRoleSuiteUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuiteuser"
	itRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteUserEntRepository(client *ent.Client) itRoleSuite.RoleSuiteUserRepository {
	return &RoleSuiteUserEntRepository{
		client: client,
	}
}

type RoleSuiteUserEntRepository struct {
	client *ent.Client
}

func BuildRoleSuiteUserDescriptor() *orm.EntityDescriptor {
	entity := ent.RoleSuiteUser{}
	builder := orm.DescribeEntity(entRoleSuiteUser.Label).
		Aliases("rolesuite_users").
		Field(entRoleSuiteUser.FieldID, entity.ID).
		Field(entRoleSuiteUser.FieldApproverID, entity.ApproverID).
		Field(entRoleSuiteUser.FieldReceiverRef, entity.ReceiverRef).
		Field(entRoleSuiteUser.FieldReceiverType, entity.ReceiverType).
		Field(entRoleSuiteUser.FieldRoleSuiteID, entity.RoleSuiteID)

	return builder.Descriptor()
}
