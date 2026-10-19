// Package settingsrecord declares the layers of the settings_record resource: one stored value per
// module, setting name and owner.
//
// Over REST the resource is read-only. Rows are written by the per-level write path of the module
// that owns the setting (core/settings port), by owner initialization and by the enforcement
// fan-out, never by a generic CRUD call.
package settingsrecord

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// SettingsRecordRepository adds the set-based statements the settings write path needs on top of
// the row-at-a-time CRUD. Both join the transaction the context carries.
type SettingsRecordRepository interface {
	composable.CrudRepository

	// FanOutValue overwrites one setting's value and override flag on every child record of the
	// tenant, in one UPDATE. It returns the number of rows touched.
	FanOutValue(ctx corectx.Context, param FanOutValueParam) (int, error)

	// CopyTenantRows seeds a new owner with copies of the tenant's own level=tenant rows, in one
	// INSERT ... SELECT that skips rows the owner already holds. It returns the rows inserted.
	CopyTenantRows(ctx corectx.Context, param CopyTenantRowsParam) (int, error)

	// HasTenantColumn tells whether this binary's settings_records carries a tenant column: true
	// in the multi-tenant host, where apptrait injects one, false in the standalone nikkierp.
	HasTenantColumn() bool
}

// FanOutValueParam names the setting to enforce and the owners that receive it.
type FanOutValueParam struct {
	ModuleKey      string
	Name           string
	ChildOwnerType string
	Value          any
	AllowOverride  bool

	// TenantId confines the statement to one tenant. Empty is accepted only when the table has
	// no tenant column; otherwise the statement is refused rather than widened.
	TenantId string
}

// CopyTenantRowsParam names the new owner and the level its rows are stored at.
type CopyTenantRowsParam struct {
	Level     string
	OwnerType string
	OwnerId   string

	// TenantId confines the copy to one tenant, with the same rule as FanOutValueParam.TenantId.
	TenantId string
}

type SettingsRecordDomainService interface {
	composable.CrudDomainService
}

type SettingsRecordApplicationService interface {
	composable.CrudApplicationService
}

type (
	GetSettingsRecordByIdQuery = composable.GetByIdQuery
	SearchSettingsRecordsQuery = composable.SearchQuery
	SettingsRecordExistsQuery  = composable.ExistsQuery
)

type (
	GetSettingsRecordByIdResult = composable.GetOneResult
	SearchSettingsRecordsResult = composable.SearchResult
	SettingsRecordExistsResult  = composable.ExistsResult
)
