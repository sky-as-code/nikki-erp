package permission

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

// PermissionHistoryRepository is the append-only audit trail of grants and
// revocations.
//
// It deliberately exposes no update or delete: an audit trail that can be edited
// answers no question worth asking. Reading is a search like any other; writing
// happens inside the same transaction as the mutation being recorded, so a
// permission change and its audit row commit together or not at all.
type PermissionHistoryRepository interface {
	dyn.DynamicModelRepository

	Insert(ctx corectx.Context, entry models.PermissionHistory) (*dyn.OpResult[int], error)
	Search(ctx corectx.Context, param dyn.RepoSearchParam) (*dyn.OpResult[dyn.PagedResultData[models.PermissionHistory]], error)
}
