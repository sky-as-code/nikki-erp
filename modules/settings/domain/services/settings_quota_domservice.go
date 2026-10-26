package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	itQuota "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsquota"
)

// NewSettingsQuotaDomainService is handed the composable default by the onion. The resource is
// read-only over REST, so it carries no write rule.
func NewSettingsQuotaDomainService(base composable.CrudDomainService) itQuota.SettingsQuotaDomainService {
	return &SettingsQuotaDomainServiceImpl{CrudDomainService: base}
}

type SettingsQuotaDomainServiceImpl struct {
	composable.CrudDomainService
}

// FindQuota looks the row up by its unique key. The multi-tenant repository confines the read to
// the acting tenant, so a tenant can never read another tenant's quota by naming its id.
func (this *SettingsQuotaDomainServiceImpl) FindQuota(
	ctx corectx.Context, name string, ownerType string, ownerId model.Id,
) (*models.SettingsQuota, error) {
	found, err := this.Repository().FindByKeys(ctx, dmodel.DynamicFields{
		models.SettingsQuotaFieldName:      name,
		models.SettingsQuotaFieldOwnerType: ownerType,
		models.SettingsQuotaFieldOwnerId:   string(ownerId),
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindQuota")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "FindQuota")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewSettingsQuotaFrom(found.Data), nil
}
