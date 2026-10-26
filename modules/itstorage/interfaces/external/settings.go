// Package external declares the ports IT Storage holds onto other modules. infra/external binds
// them; nothing else in the module imports a peer.
package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// QuotaExtService is IT Storage's port onto the Settings Quota read, narrowed to that one call.
type QuotaExtService interface {
	GetQuota(ctx corectx.Context, query GetQuotaQuery) (*GetQuotaResult, error)
}

type GetQuotaQuery = itSettings.GetQuotaQuery
type GetQuotaResult = itSettings.GetQuotaResult
