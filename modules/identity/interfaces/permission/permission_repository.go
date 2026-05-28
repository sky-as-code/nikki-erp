package permission

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	c "github.com/sky-as-code/nikki-erp/modules/identity/constants"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain/models"
)

type PermissionRepository interface {
	dyn.DynamicModelRepository

	MatchPermisions(ctx corectx.Context, param RepoMatchUserPermParam) (*dyn.OpResult[[]models.UserPermission], error)
	RebuildUserPermission(ctx corectx.Context, userId model.Id) error
	RebuildAllUserPermissions(ctx corectx.Context) error
	// ListByUser(ctx corectx.Context, param RepoListByUserParam) (*dyn.OpResult[[]models.UserPermission], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.UserPermission]], error)
}

type RepoMatchUserPermParam struct {
	UserId       model.Id
	ActionCode   string
	ResourceCode string
	Scope        c.ResourceScope
	ScopeId      *model.Id
}

type RepoListByUserParam struct {
	Fields    []string
	UserId    *model.Id
	UserEmail *string
}
