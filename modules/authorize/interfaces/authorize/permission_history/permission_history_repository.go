package permission_history

import (
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
)

type PermissionHistoryRepository interface {
	Create(ctx crud.Context, permissionHistory domain.PermissionHistory) (*domain.PermissionHistory, error)
}
