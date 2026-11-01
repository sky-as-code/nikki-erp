package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

const (
	TenantSettingsSchemaName = "accounting_tenant_settings"

	// TenantSettingSystemCurrency is the system currency of every organization that sets none of
	// its own. A setting name is unique across a module's levels, hence the prefix: the org-level
	// OrgSettingSystemCurrency is the same question asked one level down.
	TenantSettingSystemCurrency = "tenant_system_currency"
)

//go:embed tenant_settings.json
var tenantSettingsSchemaJson string

func TenantSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(tenantSettingsSchemaJson)
}
