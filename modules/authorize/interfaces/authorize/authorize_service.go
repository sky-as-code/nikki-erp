package authorize

import (
	"context"

	// "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type AuthorizeService interface {
	IsAuthorized(ctx context.Context, query IsAuthorizedQuery) (*IsAuthorizedResult, error)
}

type IsAuthorizedQuery struct {
	Action      string                         `json:"action"`
	Resource    string                         `json:"resource"`
	SubjectRef  string                         `json:"subjectRef"`
	// SubjectType *domain.EntitlementSubjectType `json:"subjectType"`
	ScopeRef    *string                        `json:"scopeRef"`
}

type IsAuthorizedResult struct {
	IsAuthorized bool `json:"isAuthorized"`
}
