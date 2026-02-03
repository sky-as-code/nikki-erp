package middleware

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/fault"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type CqrsPermissionChecker struct {
	cqrsBus cqrs.CqrsBus
}

func NewCqrsPermissionChecker(cqrsBus cqrs.CqrsBus) *CqrsPermissionChecker {
	return &CqrsPermissionChecker{cqrsBus: cqrsBus}
}

func (c *CqrsPermissionChecker) CheckPermission(
	ctx context.Context,
	subjectRef,
	resourceName,
	actionName,
	scopeRef string,
) (bool, error) {
	query := it.IsAuthorizedQuery{
		SubjectType:  it.SubjectTypeUser,
		SubjectRef:   subjectRef,
		ResourceName: resourceName,
		ActionName:   actionName,
		ScopeRef:     scopeRef,
	}
	result := it.IsAuthorizedResult{}
	err := c.cqrsBus.Request(ctx, &query, &result)
	fault.PanicOnErr(err)

	if result.ClientError != nil {
		return false, result.ClientError
	}
	return result.Decision != nil && *result.Decision == it.DecisionAllow, nil
}
