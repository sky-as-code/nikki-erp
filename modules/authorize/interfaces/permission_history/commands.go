package permission_history

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	util.Unused(req)
}

// START: FindAllByEntitlementIdQuery
var findAllByEntitlementIdQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permissionHistory",
	Action:    "findAllByEntitlementId",
}

type FindAllByEntitlementIdQuery struct {
	EntitlementId model.Id `json:"entitlementId"`
}

func (FindAllByEntitlementIdQuery) CqrsRequestType() cqrs.RequestType {
	return findAllByEntitlementIdQueryType
}

type FindAllByEntitlementIdResult = crud.OpResult[[]domain.PermissionHistory]

// END: FindAllByEntitlementIdQuery

// START: EnableFieldCommand
var enableFieldCommandType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permissionHistory",
	Action:    "enableField",
}

type EnableFieldCommand struct {
	EntitlementId   *model.Id `json:"entitlementId"`
	EntitlementExpr string    `json:"entitlementExpr"`
	AssignmentId    *model.Id `json:"assignmentId"`
	ResolvedExpr    string    `json:"resolvedExpr"`
}

func (EnableFieldCommand) CqrsRequestType() cqrs.RequestType {
	return enableFieldCommandType
}

type EnableFieldResult = crud.OpResult[[]domain.PermissionHistory]

// END: EnableFieldCommand
