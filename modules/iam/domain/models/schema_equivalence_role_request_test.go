package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The legacy chain below is a verbatim copy of what role_request.go contained before the JSON
// conversion, and exists only as the reference to diff against.

func TestRoleRequestSchemaJsonMatchesLegacy(t *testing.T) {
	requireBaseSchemasRegistered(t)
	expected := roleRequestSchemaBuilderLegacy().Build()
	actual := RoleRequestSchemaBuilder().Build()
	assertSchemasEqual(t, expected, actual)
	assert.Equal(t, expected.DefaultSearchFields(), actual.DefaultSearchFields())
	// The schema exposes no accessor for exclusive required fields, so the rule is compared by
	// what it does: a request naming both receivers is refused the same way by both schemas.
	both := dmodel.DynamicFields{
		RoleReqFieldReceiverGroupId: "01JTESTGROUP00000000000000",
		RoleReqFieldReceiverUserId:  "01JTESTUSER000000000000000",
	}
	_, expectedErrs := expected.Validate(both)
	_, actualErrs := actual.Validate(both)
	assert.Equal(t, expectedErrs.Count(), actualErrs.Count())
}

func roleRequestSchemaBuilderLegacy() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleRequestSchemaName).
		Label(model.NewLangJsonRefSf("%s.label", RoleRequestSchemaName)).
		TableName("iam_grant_requests").
		DefaultSearchFields("status", "type", "request_comment").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RoleReqFieldRoleId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldReceiverGroupId),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldReceiverUserId),
		).
		ExclusiveRequiredFields(RoleReqFieldReceiverGroupId, RoleReqFieldReceiverUserId).
		Field(
			dmodel.DefineField().Name(RoleReqFieldStatus).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RoleReqStatusPending), string(RoleReqStatusApproved), string(RoleReqStatusRejected), string(RoleReqStatusCancelled),
				})).
				RequiredForCreate().
				Default(string(RoleReqStatusPending)),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldType).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(RoleReqTypeGrant), string(RoleReqTypeRevoke),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldAttachmentUrl).
				DataType(dmodel.FieldDataTypeUrl()),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldGrantExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				Description(model.LangJson{"en-US": "The date and time when the grant (if approved) expires."}),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRequestComment).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_COMMENT_LENGTH)),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldRequestorId).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRejectionReason).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_COMMENT_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(RoleReqFieldRespondedAt).
				DataType(dmodel.FieldDataTypeDateTime()),
		).
		Field(
			basemodel.DefineFieldId(RoleReqFieldResponderId),
		).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.TraceableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeRole).
				Label(model.LangJson{"en-US": "Requested Role"}).
				ManyToOne(RoleSchemaName, dmodel.DynamicFields{
					RoleReqFieldRoleId: RoleFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeReceiverGroup).
				Label(model.LangJson{"en-US": "Receiver Group"}).
				ManyToOne(GroupSchemaName, dmodel.DynamicFields{
					RoleReqFieldReceiverGroupId: GroupFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeReceiverUser).
				Label(model.LangJson{"en-US": "Receiver User"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldReceiverUserId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeRequestor).
				Label(model.LangJson{"en-US": "Requestor"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldRequestorId: UserFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleReqEdgeResponder).
				Label(model.LangJson{"en-US": "Responder"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleReqFieldResponderId: UserFieldId,
				}),
		)
}
