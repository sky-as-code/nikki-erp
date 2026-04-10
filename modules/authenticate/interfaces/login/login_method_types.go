package login

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type FindByIdParam = GetAttemptQuery

type LoginParam struct {
	PrincipalType domain.PrincipalType `json:"principal_type"`
	Username      string               `json:"username"`
	Password      string               `json:"password"`
}

type LoginMethod interface {
	Execute(ctx corectx.Context, param LoginParam) (*ExecuteResult, error)
	Name() string
	SkipMethod() *SkippedMethod
}

type ExecuteResult struct {
	IsVerified   bool                `json:"is_verified"`
	FailedReason *ft.ClientErrorItem `json:"failed_reason,omitempty"`
	ClientErrors ft.ClientErrors
}

type SkippedMethod string

const (
	SkippedMethodAll      SkippedMethod = "*"
	SkippedMethodPassword SkippedMethod = "password"
)
