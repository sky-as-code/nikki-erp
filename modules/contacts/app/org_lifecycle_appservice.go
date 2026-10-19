package app

import (
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.CommChannelSchemaName, models.PartySchemaName, models.VendorProfileSchemaName,
}

func NewOrgLifecycleAppService() orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{OrgDataCleanup: composable.NewOrgDataCleanup(orgCleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl writes no defaults: a new organization starts with no parties and their channels.
// Cleanup deletes the organization's.
type OrgLifecycleAppServiceImpl struct {
	composable.OrgDataCleanup
}
