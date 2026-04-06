package domain

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	EntitlementSchemaName = "authorize.entitlement"

	EntitlementFieldId          = "id"
	EntitlementFieldDescription = "description"
	EntitlementFieldName        = "name"
	EntitlementFieldActionId    = "action_id"
	EntitlementFieldScope       = "scope"
	EntitlementFieldOrgId       = "org_id"
	EntitlementFieldOrgUnitId   = "org_unit_id"
	EntitlementFieldRoleId      = "role_id"

	// SailPoint-like fields
	//EntitlementFieldApplicationId = "application_id" // 'nikkierp', 'active_directory', 'sap', 'client_xyz'
	//EntitlementFieldAttributeId   = "attribute_id"   // 'memberOf' (AD), 'role' (SAP), 'group' (Nikki ERP)
	//EntitlementFieldAttributeValue   = "attribute_value"   // '{AD Group Name}' (AD)
	// For example: Map between Active Directory Group and Nikki ERP Entitlement

	EntitlementEdgeAction  = "action"
	EntitlementEdgeOrg     = "org"
	EntitlementEdgeOrgUnit = "org_unit"
	EntitlementEdgeRole    = "role"
)

func EntitlementSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel(EntitlementSchemaName).
		Label(model.LangJson{"en-US": "Entitlement"}).
		TableName("authz_entitlements").
		ShouldBuildDb().
		CompositeUnique(EntitlementFieldActionId, EntitlementFieldRoleId).
		Extend(basemodel.BaseModelSchemaBuilder()).
		Field(
			dmodel.DefineField().Name(EntitlementFieldName).
				DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_SHORT_NAME_LENGTH)).
				RequiredForCreate().
				Unique(),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldDescription).
				DataType(dmodel.FieldDataTypeString(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldActionId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldRoleId).
				DataType(dmodel.FieldDataTypeUlid()).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldScope).
				DataType(dmodel.FieldDataTypeEnumString([]string{
					string(ResourceScopeDomain), string(ResourceScopeOrg),
					string(ResourceScopeOrgUnit), string(ResourceScopePrivate),
				})).
				RequiredForCreate(),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldOrgId).
				DataType(dmodel.FieldDataTypeUlid()).
				Description(model.LangJson{"en-US": "If scope=org, this field is required. "}),
		).
		Field(
			dmodel.DefineField().Name(EntitlementFieldOrgUnitId).
				DataType(dmodel.FieldDataTypeUlid()).
				Description(model.LangJson{"en-US": "If scope=org_unit, this field is required. "}),
		).
		ExclusiveFields(EntitlementFieldOrgId, EntitlementFieldOrgUnitId).
		// Field(
		// 	dmodel.DefineField().Name(EntitlementFieldAttributeValue).
		// 		DataType(dmodel.FieldDataTypeUlid()).
		// 		Description(model.LangJson{"en-US": "After discovery process, users with matching attribute value will be granted this entitlement. " +
		// 			"If the attribute allow multiple values, the UI will allow creating an entitlement for each value.",
		// 		}),
		// ).
		Extend(basemodel.ArchivableModelSchemaBuilder()).
		Extend(basemodel.AuditableModelSchemaBuilder()).
		Extend(basemodel.VersionedModelSchemaBuilder()).
		EdgeTo(
			dmodel.Edge(EntitlementEdgeAction).
				Label(model.LangJson{"en-US": "Action"}).
				ManyToOne(ActionSchemaName, dmodel.DynamicFields{
					EntitlementFieldActionId: ActionFieldId,
				}).
				OnDelete(dmodel.RelationCascadeNoAction),
		).
		EdgeTo(
			dmodel.Edge(EntitlementEdgeOrgUnit).
				Label(model.LangJson{"en-US": "Organizational Unit"}).
				ManyToOne(OrganizationalUnitSchemaName, dmodel.DynamicFields{
					EntitlementFieldOrgUnitId: OrgUnitFieldId,
				}).
				OnDelete(dmodel.RelationCascadeNoAction),
		).
		EdgeTo(
			dmodel.Edge(EntitlementEdgeOrg).
				Label(model.LangJson{"en-US": "Organization"}).
				ManyToOne(OrganizationSchemaName, dmodel.DynamicFields{
					EntitlementFieldOrgId: OrgFieldId,
				}).
				OnDelete(dmodel.RelationCascadeNoAction),
		).
		EdgeTo(
			dmodel.Edge(EntitlementEdgeRole).
				Label(model.LangJson{"en-US": "Owned by role"}).
				ManyToOne(RoleSchemaName, dmodel.DynamicFields{
					EntitlementFieldRoleId: RoleFieldId,
				}).
				OnDelete(dmodel.RelationCascadeCascade),
		)
}

// Entitlement is a permission to perform an action on a resource within a scope.
// The scope is typically a orgunit level (organizational unit - OU), entitlement against an OU only has effect on the OU itself,
// which means:
// - User from parent OUs will not implicitly have the entitlement.
// - When this OU is moved to under a different parent OU, no user's access rights will be affected.
// - All user's access rights are explicitly managed by the entitlement grant process, and can be audited.
type Entitlement struct {
	fields dmodel.DynamicFields
}

func NewEntitlement() *Entitlement {
	return &Entitlement{fields: make(dmodel.DynamicFields)}
}

func NewEntitlementFrom(src dmodel.DynamicFields) *Entitlement {
	return &Entitlement{fields: src}
}

func (this Entitlement) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *Entitlement) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this Entitlement) GetActionId() *model.Id {
	return this.fields.GetModelId(EntitlementFieldActionId)
}

func (this *Entitlement) SetActionId(v *model.Id) {
	this.fields.SetModelId(EntitlementFieldActionId, v)
}

func (this Entitlement) GetScope() *ResourceScope {
	s := this.fields.GetString(EntitlementFieldScope)
	if s == nil {
		return nil
	}
	v := ResourceScope(*s)
	return &v
}

func (this *Entitlement) SetScope(v *ResourceScope) {
	if v == nil {
		this.fields.SetString(EntitlementFieldScope, nil)
		return
	}
	s := string(*v)
	this.fields.SetString(EntitlementFieldScope, &s)
}

func (this Entitlement) GetOrgUnitId() *model.Id {
	return this.fields.GetModelId(EntitlementFieldOrgUnitId)
}

func (this *Entitlement) SetOrgUnitId(v *model.Id) {
	this.fields.SetModelId(EntitlementFieldOrgUnitId, v)
}
