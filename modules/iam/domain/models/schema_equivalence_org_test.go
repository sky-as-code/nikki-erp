package models

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The legacy chains below are verbatim copies of what org.go contained before the JSON
// conversion, and exist only as the reference to diff against.

func TestOrganizationSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, organizationSchemaBuilderLegacy().Build(), OrganizationSchemaBuilder().Build())
}

func TestOrgUserRelSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	assertSchemasEqual(t, orgUserRelSchemaBuilderLegacy().Build(), OrgUserRelSchemaBuilder().Build())
}

func orgUserRelSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(OrgUsrRelSchemaName).
		TableName("iam_org_user_rel").
		ShouldBuildDb().
		// Add `id` column to cascade delete to user_permissions table
		Extend(basemodel.BaseModelSchemaBuilder()).
		CompositeUnique(dmodel.CompositeUniqueParam{Fields: []string{OrgUsrRelFieldOrgId, OrgUsrRelFieldUserId}}).
		Field(
			basemodel.DefineFieldId(OrgUsrRelFieldOrgId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(OrgUsrRelFieldUserId).
				RequiredForCreate(),
		)
}

func organizationSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(OrganizationSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", OrganizationSchemaName)).
		TableName("iam_organizations").
		DefaultSearchFields("display_name", "legal_name", "slug", "phone_number").
		RecordLabelField(OrgFieldDisplayName).
		RecordSubLabelField(OrgFieldSlug).
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().
				Name(OrgFieldAddress).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldAddress)).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldDisplayName).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldDisplayName)).
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldLegalName).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldLegalName)).
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldPhoneNumber).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldPhoneNumber)).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldSlug).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldSlug)).
				DataType(dmodel.FieldDataTypeSlug()).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldStatus).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldStatus)).
				Description(model.LangJson{"en-US": "Lifecycle of the organization: provisioning until every module has written its defaults, ready once they all have, deleted while modules clean up before the row is removed. Written only by the org lifecycle, never by a client."}).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(OrgStatusProvisioning), string(OrgStatusReady), string(OrgStatusDeleted),
				})).
				RequiredForCreate().
				Default(string(OrgStatusProvisioning)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldStatusMessage).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldStatusMessage)).
				Description(model.LangJson{"en-US": "Translation key explaining the status, set when a module failed to set up or clean up the organization."}).
				DataType(dmodel.FieldDataTypeString(0, 200)),
		).
		Field(
			dmodel.DefineField().
				Name(OrgFieldStatusMessageVars).
				Label(model.NewLangJsonRefSf("fields.%s", OrgFieldStatusMessageVars)).
				Description(model.LangJson{"en-US": "Variables interpolated into status_message, such as failed_modules."}).
				DataType(dmodel.FieldDataTypeJsonMap()),
		).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.TraceableModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(OrgEdgeUsers).
				ManyToMany(UserSchemaName, OrgUsrRelSchemaName, "org").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeTo(
			dmodel.Edge(OrgEdgeServices).
				Label(model.LangJson{"en-US": "Services"}).
				ManyToMany(ServiceSchemaName, OrgUsrRelSchemaName, "org").
				OnDelete(dmodel.RelationCascadeCascade),
		).
		EdgeFrom(
			dmodel.Edge(OrgEdgeOrgUnits).
				Label(model.LangJson{"en-US": "Organizational Units"}).
				Existing(OrganizationalUnitSchemaName, OrgUnitEdgeOrg),
		).
		EdgeFrom(
			dmodel.Edge(OrgEdgeEntitlements).
				Label(model.LangJson{"en-US": "Entitlements"}).
				Existing(EntitlementSchemaName, EntitlementEdgeOrg),
		)
}
