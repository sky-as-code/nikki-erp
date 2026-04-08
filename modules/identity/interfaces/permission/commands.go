package permission

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*IsAuthorizedQuery)(nil)
	util.Unused(req)
}

var isAuthorizedQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permission",
	Action:    "isAuthorized",
}

type IsAuthorizedQuery struct {
	UserId       model.Id             `json:"user_id"`
	ActionCode   string               `json:"action_code"`
	ResourceCode string               `json:"resource_code"`
	Scope        domain.ResourceScope `json:"scope"`
	ScopeId      *model.Id            `json:"scope_id"`
}

func (IsAuthorizedQuery) CqrsRequestType() cqrs.RequestType {
	return isAuthorizedQueryType
}

func (IsAuthorizedQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"authorize.is_authorized_query",
		func() *dmodel.ModelSchemaBuilder {
			return dmodel.DefineModel("_").
				Field(basemodel.DefineFieldId("user_id").Required()).
				Field(domain.DefineResourceFieldCode("resource_code").Required()).
				Field(domain.DefineActionFieldCode("action_code").Required()).
				Field(domain.DefineResourceFieldScope("scope").Required()).
				Field(basemodel.DefineFieldId("scope_id").Required())
		},
	)
}

type IsAuthorizedResult = dyn.OpResult[bool]

var checkPermissionsQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permission",
	Action:    "checkPermissions",
}

type CheckPermissionsQuery IsAuthorizedQuery

func (CheckPermissionsQuery) CqrsRequestType() cqrs.RequestType {
	return checkPermissionsQueryType
}

func (CheckPermissionsQuery) GetSchema() *dmodel.ModelSchema {
	return IsAuthorizedQuery{}.GetSchema()
}

type CheckPermissionsResultData struct {
	IsAuthorized bool                    `json:"is_authorized"`
	RejectReason string                  `json:"reject_reason"`
	Permissions  []domain.UserPermission `json:"permissions"`
}

type CheckPermissionsResult = dyn.OpResult[CheckPermissionsResultData]
