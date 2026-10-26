// Package settingsquota declares the layers of the settings_quota resource: how much of a named
// resource one tenant, organization or user may use.
//
// Over REST the resource is read-only. Rows come from seeds; consuming modules read them through
// settings.QuotaAppService.
package settingsquota

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

type SettingsQuotaDomainService interface {
	composable.CrudDomainService

	// FindQuota returns nil when the owner has no quota row of that name.
	FindQuota(ctx corectx.Context, name string, ownerType string, ownerId model.Id) (*models.SettingsQuota, error)
}

type SettingsQuotaApplicationService interface {
	composable.CrudApplicationService
	itSettings.QuotaAppService
}

type (
	GetSettingsQuotaByIdQuery = composable.GetByIdQuery
	SearchSettingsQuotasQuery = composable.SearchQuery
)

type (
	GetSettingsQuotaByIdResult = composable.GetOneResult
	SearchSettingsQuotasResult = composable.SearchResult
)
