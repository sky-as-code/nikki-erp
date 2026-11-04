// Package settings holds Payment & Invoice's settings schemas. Unlike a model schema, a settings
// schema owns no table: its values are stored as settings_records rows by the settings module,
// under the "paymentinvoice" module key.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// OrgSettingInvoiceNumberSequenceId holds the id of the Essential sequence definition that numbers
// the organization's invoices (paymentinvoice.invoice). The org lifecycle writes it; the module's
// settings writer refuses it from a client, because re-pointing it would restart or duplicate the
// invoice series and hide the real definition from the usage check.
const OrgSettingInvoiceNumberSequenceId = "invoice_number_sequence_id"

// SequenceIdSettings lists the settings holding a sequence definition id. Every one is server-owned.
func SequenceIdSettings() []string {
	return []string{OrgSettingInvoiceNumberSequenceId}
}

// OrgSettingsSchemaName is the name Payment & Invoice registers its org-level settings under.
const OrgSettingsSchemaName = "paymentinvoice_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the settings an organization holds for its invoicing. The
// document declares no table_name or should_build_db and extends no base model, because it owns no
// table.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
