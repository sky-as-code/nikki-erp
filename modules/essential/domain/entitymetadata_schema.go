package domain

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/model"
)

func EntitySchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity().
		TableName("essential_entities").
		Field(
			schema.DefineField().
				Name("id").
				DataType(schema.FieldDataTypeUlid).
				Required().
				Rule(schema.FieldRulePrimary()),
		).
		Field(
			schema.DefineField().
				Name("domain_id").
				DataType(schema.FieldDataTypeUlid).
				Required().
				Rule(schema.FieldRulePrimary()).
				Rule(schema.FieldRuleTenant()),
		).
		Field(
			schema.DefineField().
				Name("name").
				DataType(schema.FieldDataTypeString).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("label").
				DataType(schema.FieldDataTypeLangJson).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("description").
				DataType(schema.FieldDataTypeLangJson).
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("table_name").
				DataType(schema.FieldDataTypeString).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		)
}

func EntityRelationSchemaBuilder() *schema.EntitySchemaBuilder {
	return schema.DefineEntity().
		TableName("essential_entity_relations").
		Field(
			schema.DefineField().
				Name("src_entity_id").
				DataType(schema.FieldDataTypeUlid).
				Required().
				Foreign(schema.Edge("src_entity").OneToOne("essential_entities", "id")),
		).
		Field(
			schema.DefineField().
				Name("src_field").
				DataType(schema.FieldDataTypeString).
				Required(),
		).
		Field(
			schema.DefineField().
				Name("dest_entity_id").
				DataType(schema.FieldDataTypeUlid).
				Required().
				Rule(schema.FieldRuleLength(1, model.MODEL_RULE_TINY_NAME_LENGTH)),
		).
		Field(
			schema.DefineField().
				Name("dest_field").
				DataType(schema.FieldDataTypeString).
				Required(),
		).
		Field(
			schema.DefineField().
				Name("relation_type").
				DataType(
					schema.FieldDataTypeEnumString,
					schema.FieldDataTypeOptions{
						schema.FieldDataTypeOptEnumValues: []schema.RelationType{
							schema.RelationTypeOneToOne,
							schema.RelationTypeManyToOne,
							schema.RelationTypeManyToMany,
						},
					}).
				Required(),
		)
}
