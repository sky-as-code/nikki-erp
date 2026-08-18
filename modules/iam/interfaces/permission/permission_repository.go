package permission

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

type PermissionRepository interface {
	dyn.DynamicModelRepository

	MatchPermisions(ctx corectx.Context, param RepoMatchUserPermParam) (*dyn.OpResult[[]models.UserPermission], error)
	// FindMatchingPermissions returns the provenance rows behind a match, for the
	// permission probe. Same query shape as MatchPermisions, rows kept.
	FindMatchingPermissions(ctx corectx.Context, userId model.Id, candidates []string) ([]models.UserPermission, error)
	// ResolveGrantSourceNames maps assignment ids to the role or group name behind
	// them, so a probe answer can name the grant path in human terms.
	ResolveGrantSourceNames(ctx corectx.Context, directIds []model.Id, groupIds []model.Id) (map[model.Id]string, error)
	RebuildUserPermission(ctx corectx.Context, userId model.Id) error
	RebuildUserPermissionsForGroup(ctx corectx.Context, groupId model.Id) error
	// RebuildUserPermissionsForRole refreshes every holder of the role - directly
	// assigned or through a group - in one round trip. This is what makes a change
	// to a role's entitlements reach the people who already hold that role.
	RebuildUserPermissionsForRole(ctx corectx.Context, roleId model.Id) error
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

	// EvalContext carries the caller's org and unit membership. Without it a bare
	// `org` grant can never answer, which is how org-scoped checks used to
	// degenerate into exact-or-domain ones.
	EvalContext reguard.EvalContext

	// IsRecordOwnedByCaller answers the private scope.
	IsRecordOwnedByCaller bool
}

type RepoListByUserParam struct {
	Fields    []string
	UserId    *model.Id
	UserEmail *string
}
