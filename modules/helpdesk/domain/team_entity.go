package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	TeamSchemaName = "helpdesk.team"

	TeamFieldName      = "name"
	TeamFieldOrgId     = "org_id"
	TeamFieldManagerId = "manager_id"
	TeamFieldGroupId   = "group_id"
)

func TeamSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(TeamSchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Team"}).
		TableName("helpdesk_teams").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(TeamFieldName).DataType(dmodel.FieldDataTypeString(1, 120)).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TeamFieldOrgId).RequiredForCreate()).
		Field(basemodel.DefineFieldId(TeamFieldManagerId)).
		Field(basemodel.DefineFieldId(TeamFieldGroupId)).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type Team struct{ basemodel.DynamicModelBase }
