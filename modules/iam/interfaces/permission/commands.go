package permission

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*IsAuthorizedQuery)(nil)
	util.Unused(req)
}

var isAuthorizedQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "permission",
	Action:    "isAuthorized",
}

type IsAuthorizedQuery struct {
	UserEmail    *string         `json:"user_email"`
	UserId       *model.Id       `json:"user_id"`
	ActionCode   string          `json:"action_code"`
	ResourceCode string          `json:"resource_code"`
	Scope        c.ResourceScope `json:"scope"`
	ScopeId      *model.Id       `json:"scope_id"`

	// IsRecordOwnedByCaller answers the private scope. The calling module knows
	// who owns the record; this service only knows about grants.
	IsRecordOwnedByCaller bool `json:"is_record_owned_by_caller"`
}

func (IsAuthorizedQuery) CqrsRequestType() cqrs.RequestType {
	return isAuthorizedQueryType
}

func (IsAuthorizedQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.is_authorized_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("user_id")).
				Field(dmodel.DefineField().Name("user_email").
					DataType(dmodel.FieldDataTypeEmail()),
				).
				ExclusiveRequiredFields("user_id", "user_email").
				Field(domain.DefineResourceFieldCode("resource_code").RequiredAlways()).
				Field(domain.DefineActionFieldCode("action_code").RequiredAlways()).
				Field(domain.DefineResourceFieldScope("scope").RequiredAlways()).
				Field(basemodel.DefineFieldId("scope_id"))
		},
	)
}

type IsAuthorizedResult = dyn.OpResult[bool]

var listAllUserPermQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "permission",
	Action:    "listAllUserPermissions",
}

type ListAllUserPermissionsQuery struct {
	UserId    *model.Id `json:"user_id"`
	UserEmail *string   `json:"user_email"`
}

func (ListAllUserPermissionsQuery) CqrsRequestType() cqrs.RequestType {
	return listAllUserPermQueryType
}

func (ListAllUserPermissionsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.get_ent_expressions_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("user_id")).
				Field(dmodel.DefineField().Name("user_email").
					DataType(dmodel.FieldDataTypeEmail()),
				).
				ExclusiveRequiredFields("user_id", "user_email")
		},
	)
}

type ListAllUserPermissionsResult = dyn.OpResult[[]domain.UserPermission]

var getUserEntQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "permission",
	Action:    "getUserEntitlements",
}

type GetUserEntitlementsQuery ListAllUserPermissionsQuery

func (GetUserEntitlementsQuery) CqrsRequestType() cqrs.RequestType {
	return getUserEntQueryType
}

func (GetUserEntitlementsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.get_ent_expressions_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("user_id")).
				Field(dmodel.DefineField().Name("user_email").
					DataType(dmodel.FieldDataTypeEmail()),
				).
				ExclusiveRequiredFields("user_id", "user_email")
		},
	)
}

type GetUserEntitlementsResultData struct {
	IsOwner      bool     `json:"is_owner"`
	Entitlements []string `json:"entitlements"`
	// The org unit that user belongs to (if any)
	OrgUnitId *model.Id `json:"org_unit_id"`
	// The org that the org unit belongs to (if user belongs to an org unit)
	OrgUnitOrgId *model.Id `json:"org_unit_org_id"`
	UserId       model.Id  `json:"user_id"`
	// The orgs that user belongs to (if any)
	UserOrgIds []model.Id           `json:"user_org_ids"`
	User       dmodel.DynamicFields `json:"user"`
}

type GetUserEntitlementsResult = dyn.OpResult[GetUserEntitlementsResultData]

var testMyPermissionsQueryType = cqrs.RequestType{
	Module:    "iam",
	Submodule: "permission",
	Action:    "testMyPermissions",
}

// TestMyPermissionsQuery asks "am I granted this?" about the CALLER.
//
// There is deliberately no user id here. Probing someone else's grants stays an
// internal capability behind IsAuthorizedQuery; making the subject implicit means
// there is no parameter that a future change could forget to authorize.
type TestMyPermissionsQuery struct {
	Expression string `json:"expression"`
}

func (TestMyPermissionsQuery) CqrsRequestType() cqrs.RequestType {
	return testMyPermissionsQueryType
}

func (TestMyPermissionsQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"iam.test_my_permissions_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(domain.DefineEntitlementFieldExpression("expression").RequiredAlways())
		},
	)
}

// PermissionMatch is one grant path that answers the question.
type PermissionMatch struct {
	// SourceKind is direct, group or owner.
	SourceKind string `json:"source_kind"`
	// SourceId is the assignment id for direct/group, the user id for owner.
	SourceId model.Id `json:"source_id"`
	// SourceName is the role name for direct, the group name for group.
	SourceName string `json:"source_name"`
	// EntExpression is the stored grant that matched, which may be wider than the
	// question - that is how the caller learns *why* they are allowed.
	EntExpression string `json:"ent_expression"`
}

type TestMyPermissionsResultData struct {
	IsGranted bool `json:"is_granted"`
	// Every path that answers, not just the first. An empty slice on a denial:
	// the negative answer carries no information beyond "no".
	Matches []PermissionMatch `json:"matches"`
}

type TestMyPermissionsResult = dyn.OpResult[TestMyPermissionsResultData]
