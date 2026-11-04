package settings

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgSettingsReader reads the values one organization stores for a module, without asserting a
// permission and without guessing the organization from the request: a module reads its own
// settings (a sequence definition id, say) on behalf of a caller who need not hold settings access.
// A missing value is absent from the map; a failed read is an error, never an empty map, so "could
// not read" is never taken for "not set". The Settings module provides the one implementation.
type OrgSettingsReader interface {
	GetOrgSettingValues(ctx corectx.Context, orgId model.Id, moduleKey string) (map[string]any, error)
}
