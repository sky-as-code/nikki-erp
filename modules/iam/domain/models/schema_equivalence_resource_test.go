package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
)

// The legacy chains below are verbatim copies of what resource.go and action.go contained before
// the JSON conversion, and exist only as the reference to diff against.

func TestResourceSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, resourceSchemaBuilderLegacy().Build(), ResourceSchemaBuilder().Build())
}

func TestActionSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, actionSchemaBuilderLegacy().Build(), ActionSchemaBuilder().Build())
}

func resourceSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ResourceSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", ResourceSchemaName)).
		RecordLabelField(ResourceFieldName).
		TableName("iam_resources").
		DefaultSearchFields("name", "code", "description").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(ResourceFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			DefineResourceFieldCode(ResourceFieldCode).
				RequiredForCreate().
				Unique().
				NoUpdate(),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldOwnerType).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ResourceOwnerTypeNikki), string(ResourceOwnerTypeCustom),
				})).
				Default(string(ResourceOwnerTypeNikki)).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "A resource can be owned by one of NikkiERP modules, or by a 3rd party system. " +
					"If owner_type=nikkierp, this resource is used for NikkiERP module authorization. " +
					"If owner_type=custom, this resource is used for 3rd party system authorization.",
				}),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldMaxScope).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(c.ResourceScopeTenant), string(c.ResourceScopeOrg),
					string(c.ResourceScopeOrgUnit), string(c.ResourceScopePrivate),
				})).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "The largest scope of the resource. " +
					"No entitlement can be granted with a scope larger than this. ",
				}),
		).
		Field(
			dmodel.DefineField().Name(ResourceFieldMinScope).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(c.ResourceScopeTenant), string(c.ResourceScopeOrg),
					string(c.ResourceScopeOrgUnit), string(c.ResourceScopePrivate),
				})).
				RequiredForCreate().
				Description(model.LangJson{"en-US": "The smallest scope of the resource. " +
					"No entitlement can be granted with a scope less than this. ",
				}),
		).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeFrom(
			dmodel.Edge(ResourceEdgeActions).
				Label(model.LangJson{"en-US": "Actions"}).
				Existing(ActionSchemaName, ActionEdgeResource),
		)
}

func actionSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(ActionSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", ActionSchemaName)).
		RecordLabelField(ActionFieldName).
		TableName("iam_actions").
		DefaultSearchFields("name", "code", "description").
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{ActionFieldResourceId, ActionFieldCode}}).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(ActionFieldName).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_TINY_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			DefineActionFieldCode(ActionFieldCode).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(ActionFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(ActionFieldResourceId).
				RequiredAlways(),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(ActionEdgeResource).
				Label(model.LangJson{"en-US": "Resource"}).
				ManyToOne(ResourceSchemaName, dmodel.DynamicFields{
					ActionFieldResourceId: ResourceFieldId,
				}).
				OnDelete(dmodel.RelationCascadeNoAction),
		).
		EdgeFrom(
			dmodel.Edge(ActionEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeAction),
		)
}
