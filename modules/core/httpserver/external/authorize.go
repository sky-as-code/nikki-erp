package external

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

// Copied from identity::permission_service.go
type PermissionExtService interface {
	IsAuthorized(ctx corectx.Context, query IsAuthorizedQuery) (bool, error)
}

func NewPermissionExtServiceImpl(cqrsBus cqrs.CqrsBus) PermissionExtService {
	return &PermissionExtServiceImpl{
		CqrsBus: cqrsBus,
	}
}

type PermissionExtServiceImpl struct {
	CqrsBus cqrs.CqrsBus
}

func (this *PermissionExtServiceImpl) IsAuthorized(ctx corectx.Context, query IsAuthorizedQuery) (bool, error) {
	result := IsAuthorizedResult{}
	err := this.CqrsBus.Request(ctx, &query, &result)
	if err != nil {
		return false, err
	}
	if result.ClientErrors.Count() > 0 {
		return false, errors.Wrap(result.ClientErrors.ToError(), "PermissionExtServiceImpl.IsAuthorized")
	}
	return result.Data, nil
}

var isAuthorizedQueryType = cqrs.RequestType{
	Module:    "authorize",
	Submodule: "permission",
	Action:    "isAuthorized",
}

// Copied from identity::permission::commands.go
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

type IsAuthorizedResult = dyn.OpResult[bool]
