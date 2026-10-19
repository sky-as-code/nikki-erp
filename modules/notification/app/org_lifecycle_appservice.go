package app

import (
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
)

// orgCleanUpOrder lists every org-scoped resource children first, as their foreign keys require.
var orgCleanUpOrder = []string{
	models.DeliverySchemaName, models.NotificationSchemaName, models.RecipientSchemaName,
}

func NewOrgLifecycleAppService() orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{OrgDataCleanup: composable.NewOrgDataCleanup(orgCleanUpOrder...)}
}

// OrgLifecycleAppServiceImpl writes no defaults: a new organization starts with no notifications and their deliveries.
// Cleanup deletes the organization's.
type OrgLifecycleAppServiceImpl struct {
	composable.OrgDataCleanup
}
