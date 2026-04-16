package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TeamMembershipSchemaName = "helpdesk.team_membership"

	TeamMembershipFieldTeamId = "team_id"
	TeamMembershipFieldUserId = "user_id"
	TeamMembershipFieldRole   = "role"
)

func TeamMembershipSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TeamMembershipSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Team membership"}).
		TableName("helpdesk_team_memberships").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(TeamMembershipFieldTeamId, TeamMembershipFieldUserId).
		Field(basemodel.DefineFieldId(TeamMembershipFieldTeamId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TeamMembershipFieldUserId).RequiredForCreate()).
		Field(dmodel.DefineField().Name(TeamMembershipFieldRole).DataType(
			dmodel.FieldDataTypeEnumString([]string{"agent", "supervisor"}),
		).RequiredForCreate()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type TeamMembership struct{ basemodel.DynamicModelBase }
