package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RoleAssignmentSchemaName = "authorize.role_assignment"

	RoleAssignFieldRoleId          = "role_id"
	RoleAssignFieldReceiverGroupId = "receiver_group_id"
	RoleAssignFieldReceiverUserId  = "receiver_user_id"
	RoleAssignFieldRoleRequestId   = "role_request_id"
	RoleAssignFieldApproverId      = "approver_id"
	RoleAssignFieldExpiresAt       = "expires_at"

	RoleAssignEdgeRoleRequest = "role_request"
	RoleAssignEdgeApprover    = "approver"
)

func RoleAssignmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleAssignmentSchemaName).
		Label(model.LangJson{"en-US": "Role Assignment"}).
		TableName("authz_role_assignments").
		ShouldBuildDb().
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.AuditableReadonlyModelSchemaBuilder()).
		CompositeUnique(RoleAssignFieldRoleId, RoleAssignFieldReceiverGroupId).
		CompositeUnique(RoleAssignFieldRoleId, RoleAssignFieldReceiverUserId).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldRoleId).
				DataType(dmodel.FieldDataTypeUlid()).
				PrimaryKey(),
		).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldReceiverGroupId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldReceiverUserId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldApproverId).
				DataType(dmodel.FieldDataTypeUlid()).
				Description(model.LangJson{"en-US": "If request_id is set, the granter is the request approver. " +
					"Otherwise, the granter is the user who directly granted the entitlement.",
				}),
		).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldRoleRequestId).
				DataType(dmodel.FieldDataTypeUlid()),
		).
		Field(
			dmodel.DefineField().Name(RoleAssignFieldExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				Description(model.LangJson{"en-US": "The date and time when the entitlement grant expires."}),
		).
		EdgeTo(
			dmodel.Edge(RoleAssignEdgeRoleRequest).
				Label(model.LangJson{"en-US": "Role Request"}).
				ManyToOne(RoleRequestSchemaName, dmodel.DynamicFields{
					RoleAssignFieldRoleRequestId: RoleReqFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleAssignEdgeApprover).
				Label(model.LangJson{"en-US": "Approver"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleAssignFieldApproverId: UserFieldId,
				}),
		)
}

type RoleAssignment struct {
	fields dmodel.DynamicFields
}

func NewRoleAssignment() *RoleAssignment {
	return &RoleAssignment{fields: make(dmodel.DynamicFields)}
}

func NewRoleAssignmentFrom(src dmodel.DynamicFields) *RoleAssignment {
	return &RoleAssignment{fields: src}
}

func (this RoleAssignment) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *RoleAssignment) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}
