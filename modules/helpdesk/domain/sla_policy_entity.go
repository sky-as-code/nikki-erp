package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SlaPolicySchemaName = "helpdesk.sla_policy"

	SlaPolicyFieldName                 = "name"
	SlaPolicyFieldFirstResponseMinutes = "first_response_minutes"
	SlaPolicyFieldResolutionMinutes    = "resolution_minutes"
	SlaPolicyFieldBusinessHoursId      = "business_hours_id"
	SlaPolicyFieldEscalationRules      = "escalation_rules"
)

func SlaPolicySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(SlaPolicySchemaName).
		Label(model.LangJson{model.LanguageCodeEnUs: "Sla policy"}).
		TableName("helpdesk_sla_policies").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(dmodel.DefineField().Name(SlaPolicyFieldName).DataType(dmodel.FieldDataTypeString(1, 120)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(SlaPolicyFieldFirstResponseMinutes).DataType(dmodel.FieldDataTypeInt32(0, 1000000)).RequiredForCreate()).
		Field(dmodel.DefineField().Name(SlaPolicyFieldResolutionMinutes).DataType(dmodel.FieldDataTypeInt32(0, 1000000)).RequiredForCreate()).
		Field(basemodel.DefineFieldId(SlaPolicyFieldBusinessHoursId)).
		Field(dmodel.DefineField().Name(SlaPolicyFieldEscalationRules).DataType(dmodel.FieldDataTypeModel())).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder())
}

type SlaPolicy struct{ basemodel.DynamicModelBase }
