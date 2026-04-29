package models

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RoleUserAssignmentSchemaName = "authz.role_user_assignment"

	RoleUserAssignFieldId             = basemodel.FieldId
	RoleUserAssignFieldRoleId         = "role_id"
	RoleUserAssignFieldReceiverUserId = "receiver_user_id"
	RoleUserAssignFieldRoleRequestId  = "role_request_id"
	RoleUserAssignFieldApproverId     = "approver_id"
	RoleUserAssignFieldExpiresAt      = "expires_at"

	RoleUserAssignEdgeRoleRequest = "role_request"
	RoleUserAssignEdgeApprover    = "approver"
)

func RoleUserAssignmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleUserAssignmentSchemaName).
		Label(model.LangJson{"en-US": "Role-User Assignment"}).
		TableName("authz_role_user_assignments").
		ShouldBuildDb().
		CompositeUnique(RoleUserAssignFieldRoleId, RoleUserAssignFieldReceiverUserId).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.AuditableReadonlyModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RoleUserAssignFieldRoleId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RoleUserAssignFieldReceiverUserId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RoleUserAssignFieldApproverId).
				Description(model.LangJson{"en-US": "If request_id is set, the granter is the request approver. " +
					"Otherwise, the granter is the user who directly granted the entitlement.",
				}),
		).
		Field(
			basemodel.DefineFieldId(RoleUserAssignFieldRoleRequestId),
		).
		Field(
			dmodel.DefineField().Name(RoleUserAssignFieldExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				Description(model.LangJson{"en-US": "The date and time when the entitlement grant expires."}),
		).
		EdgeTo(
			dmodel.Edge(RoleUserAssignEdgeRoleRequest).
				Label(model.LangJson{"en-US": "Role Request"}).
				ManyToOne(RoleRequestSchemaName, dmodel.DynamicFields{
					RoleUserAssignFieldRoleRequestId: RoleReqFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleUserAssignEdgeApprover).
				Label(model.LangJson{"en-US": "Approver"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleUserAssignFieldApproverId: UserFieldId,
				}),
		)
}

type RoleUserAssignment struct {
	fields dmodel.DynamicFields
}

func NewRoleUserAssignment() *RoleUserAssignment {
	return &RoleUserAssignment{fields: make(dmodel.DynamicFields)}
}

func NewRoleUserAssignmentFrom(src dmodel.DynamicFields) *RoleUserAssignment {
	return &RoleUserAssignment{fields: src}
}

func (this RoleUserAssignment) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *RoleUserAssignment) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}
