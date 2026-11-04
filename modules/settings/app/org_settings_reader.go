package app

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// NewOrgSettingsReader reads through the Settings DOMAIN service: the application services either
// assert a settings permission the caller (a warehouse operator scrapping goods, a kiosk device, an
// admin archiving a sequence) need not hold, or swallow a failed level read, and neither suits a
// reader that must answer "which definition numbers this org's documents" or fail loudly.
func NewOrgSettingsReader(settings itSettings.SettingsDomainService) coreSettings.OrgSettingsReader {
	return &orgSettingsReader{settings: settings}
}

type orgSettingsReader struct {
	settings itSettings.SettingsDomainService
}

func (this *orgSettingsReader) GetOrgSettingValues(
	ctx corectx.Context, orgId model.Id, moduleKey string,
) (map[string]any, error) {
	if orgId == "" {
		return nil, errors.New("GetOrgSettingValues: an organization id is required")
	}
	// A clone, so the acting org recorded for this read never leaks into the caller's request.
	scoped := corectx.CloneRequestContext(ctx)
	dyn.WithActingOrg(scoped, orgId)
	result, err := this.settings.GetSettings(scoped, c.LevelOrg, c.OwnerTypeOrg, itSettings.GetSettingsQuery{
		ModuleKey: moduleKey, OrgId: &orgId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetOrgSettingValues")
	}
	if result == nil || result.ClientErrors.Count() > 0 {
		return nil, errors.Errorf("GetOrgSettingValues: settings of %s in org %s are unreadable", moduleKey, orgId)
	}
	values := map[string]any{}
	for _, item := range result.Data.Items {
		if item.Value != nil {
			values[item.Name] = item.Value
		}
	}
	return values, nil
}
