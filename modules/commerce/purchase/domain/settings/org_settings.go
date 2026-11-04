// Package settings holds Purchase's settings schemas. Unlike a model schema, a settings schema owns
// no table: its values are stored as settings_records rows by the settings module, under the
// "purchase" module key.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The Essential sequence definitions Purchase numbers its orders and agreements from, registered by
// the org lifecycle. Only Purchase writes them: re-pointing one would restart or duplicate a code
// series and hide the real definition from the usage check that keeps it from being archived.
const (
	OrgSettingOrderCodeSequenceId     = "order_code_sequence_id"
	OrgSettingAgreementCodeSequenceId = "agreement_code_sequence_id"
)

// SequenceIdSettings lists the settings holding a sequence definition id. Every one is server-owned.
func SequenceIdSettings() []string {
	return []string{OrgSettingOrderCodeSequenceId, OrgSettingAgreementCodeSequenceId}
}

// OrgSettingsSchemaName is the name Purchase registers its org-level settings under.
const OrgSettingsSchemaName = "purchase_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the settings an organization holds for its purchasing. The
// document declares no table_name or should_build_db and extends no base model, because it owns no
// table.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
