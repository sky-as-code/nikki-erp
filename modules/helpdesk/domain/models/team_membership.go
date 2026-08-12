package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TeamMembershipSchemaName = "helpdesk_team_membership"

	TeamMembershipFieldTeamId = "team_id"
	TeamMembershipFieldUserId = "user_id"
	TeamMembershipFieldRole   = "role"
)

func TeamMembershipSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TeamMembershipSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", TeamMembershipSchemaName)).
		TableName("helpdesk_team_memberships").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{TeamMembershipFieldTeamId, TeamMembershipFieldUserId}}).
		Field(basemodel.DefineFieldId(TeamMembershipFieldTeamId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TeamMembershipFieldUserId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TeamMembershipFieldRole).DataType(
			dmodel.FieldDataTypeEnumString([]string{"agent", "supervisor"}),
		).RequiredForCreate()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TeamMembership struct{ basemodel.DynamicModelBase }
