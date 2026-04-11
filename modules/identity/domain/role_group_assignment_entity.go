package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	RoleGroupAssignmentSchemaName = "authorize.role_group_assignment"

	RoleGroupAssignFieldId              = basemodel.FieldId
	RoleGroupAssignFieldRoleId          = "role_id"
	RoleGroupAssignFieldReceiverGroupId = "receiver_group_id"
	RoleGroupAssignFieldRoleRequestId   = "role_request_id"
	RoleGroupAssignFieldApproverId      = "approver_id"
	RoleGroupAssignFieldExpiresAt       = "expires_at"

	RoleGroupAssignEdgeRoleRequest = "role_request"
	RoleGroupAssignEdgeApprover    = "approver"
)

func RoleGroupAssignmentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(RoleGroupAssignmentSchemaName).
		Label(model.LangJson{"en-US": "Role-Group Assignment"}).
		TableName("authz_role_group_assignments").
		ShouldBuildDb().
		CompositeUnique(RoleGroupAssignFieldRoleId, RoleGroupAssignFieldReceiverGroupId).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Extend(basemodel.AuditableReadonlyModelSchemaBuilder()).
		Field(
			basemodel.DefineFieldId(RoleGroupAssignFieldRoleId).
				RequiredForCreate(),
		).
		Field(
			basemodel.DefineFieldId(RoleGroupAssignFieldReceiverGroupId).
				Required(),
		).
		Field(
			basemodel.DefineFieldId(RoleGroupAssignFieldApproverId).
				Description(model.LangJson{"en-US": "If request_id is set, the granter is the request approver. " +
					"Otherwise, the granter is the user who directly granted the entitlement.",
				}),
		).
		Field(
			basemodel.DefineFieldId(RoleGroupAssignFieldRoleRequestId),
		).
		Field(
			dmodel.DefineField().Name(RoleGroupAssignFieldExpiresAt).
				DataType(dmodel.FieldDataTypeDateTime()).
				Description(model.LangJson{"en-US": "The date and time when the entitlement grant expires."}),
		).
		EdgeTo(
			dmodel.Edge(RoleGroupAssignEdgeRoleRequest).
				Label(model.LangJson{"en-US": "Role Request"}).
				ManyToOne(RoleRequestSchemaName, dmodel.DynamicFields{
					RoleGroupAssignFieldRoleRequestId: RoleReqFieldId,
				}),
		).
		EdgeTo(
			dmodel.Edge(RoleGroupAssignEdgeApprover).
				Label(model.LangJson{"en-US": "Approver"}).
				ManyToOne(UserSchemaName, dmodel.DynamicFields{
					RoleGroupAssignFieldApproverId: UserFieldId,
				}),
		)
}

type RoleGroupAssignment struct {
	fields dmodel.DynamicFields
}

func NewRoleGroupAssignment() *RoleGroupAssignment {
	return &RoleGroupAssignment{fields: make(dmodel.DynamicFields)}
}

func NewRoleGroupAssignmentFrom(src dmodel.DynamicFields) *RoleGroupAssignment {
	return &RoleGroupAssignment{fields: src}
}

func (this RoleGroupAssignment) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *RoleGroupAssignment) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}
