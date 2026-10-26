package app

import (
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
	itQuota "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsquota"
)

// NewSettingsQuotaApplicationService is handed the composable default by the onion: the
// authorized reads, plus the unauthorized GetQuota port.
func NewSettingsQuotaApplicationService(
	base composable.CrudApplicationService,
) itQuota.SettingsQuotaApplicationService {
	return &SettingsQuotaApplicationServiceImpl{CrudApplicationService: base}
}

type SettingsQuotaApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// GetQuota is the cross-module read, so it asserts nothing; see itSettings.QuotaAppService.
func (this *SettingsQuotaApplicationServiceImpl) GetQuota(
	ctx corectx.Context, query itSettings.GetQuotaQuery,
) (*itSettings.GetQuotaResult, error) {
	domSvc := this.DomainService().(itQuota.SettingsQuotaDomainService)
	quota, err := domSvc.FindQuota(ctx, query.Name, query.OwnerType, query.OwnerId)
	if err != nil || quota == nil {
		return &itSettings.GetQuotaResult{}, err
	}
	return &itSettings.GetQuotaResult{
		HasData: true,
		Data: itSettings.GetQuotaResultData{
			Id:                       util.ValueOrZeroOf(quota.GetId()),
			Name:                     util.ValueOrZeroOf(quota.GetName()),
			OwnerType:                util.ValueOrZeroOf(quota.GetOwnerType()),
			OwnerId:                  util.ValueOrZeroOf(quota.GetOwnerId()),
			ItStorageCapacityMibytes: quota.GetItStorageCapacityMibytes(),
		},
	}, nil
}
