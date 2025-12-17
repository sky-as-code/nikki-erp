package domain

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
)

func UserSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity().
		Field(
			schema.DefineField().
				Name("display_name").
				Label(model.LangJson{"en-US": "Display Name"}).
				DataType(schema.FieldDataTypeString).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("avatar_url").
				Label(model.LangJson{"en-US": "Avatar URL"}).
				DataType(schema.FieldDataTypeUrl).
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_URL_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("status").
				Label(model.LangJson{"en-US": "Status"}).
				DataType(schema.FieldDataTypeEnumString).
				Required().
				Rule(schema.FieldRuleOneOf(UserStatusActive, UserStatusArchived, UserStatusLocked)),
		).
		Field(
			schema.DefineField().
				Name("hierarchy_id").
				DataType(schema.FieldDataTypeUlid),
		)
}
