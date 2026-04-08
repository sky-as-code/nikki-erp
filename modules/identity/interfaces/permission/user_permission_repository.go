package permission

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

type UserPermissionRepository interface {
	RebuildUserPermission(ctx corectx.Context, userId model.Id) error
	RebuildAllUserPermissions(ctx corectx.Context) error
	MatchPermisions(ctx corectx.Context, param RepoMatchUserPermParam) (*dyn.OpResult[[]domain.UserPermission], error)
}

type RepoMatchUserPermParam struct {
	UserId       model.Id
	ActionCode   string
	ResourceCode string
	Scope        domain.ResourceScope
	ScopeId      *model.Id
}
