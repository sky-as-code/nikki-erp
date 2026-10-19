package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

func NewOrgLifecycleAppService(settingsSvc it.SettingsDomainService) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{settingsSvc: settingsSvc}
}

// OrgLifecycleAppServiceImpl removes an org's settings. Setting them up is not here: iam seeds an
// org's settings inside the transaction that creates it, so the org never exists without them.
type OrgLifecycleAppServiceImpl struct {
	settingsSvc it.SettingsDomainService
}

func (this *OrgLifecycleAppServiceImpl) InitOrg(corectx.Context, model.Id) error {
	return nil
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	_, err := this.settingsSvc.DeleteOwnerSettings(ctx, c.OwnerTypeOrg, orgId)
	return err
}
