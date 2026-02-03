package middleware

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type PermissionCheckerAdapter struct {
	svc it.AuthorizeService
}

func NewPermissionCheckerAdapter(svc it.AuthorizeService) *PermissionCheckerAdapter {
	return &PermissionCheckerAdapter{svc: svc}
}

func (a *PermissionCheckerAdapter) CheckPermission(
	ctx context.Context,
	subjectRef,
	resourceName,
	actionName,
	scopeRef string,
) (bool, error) {
	crudCtx, ok := ctx.(crud.Context)
	if !ok {
		return false, nil
	}

	result, err := a.svc.IsAuthorized(crudCtx, it.IsAuthorizedQuery{
		SubjectType:  it.SubjectTypeUser,
		SubjectRef:   subjectRef,
		ResourceName: resourceName,
		ActionName:   actionName,
		ScopeRef:     scopeRef,
	})
	fault.PanicOnErr(err)

	if result.ClientError != nil {
		return false, result.ClientError
	}
	return result.Decision != nil && *result.Decision == it.DecisionAllow, nil
}
