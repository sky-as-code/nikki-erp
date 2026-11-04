// Package settings holds Inventory's settings schemas. Unlike a model schema, a settings schema owns
// no table: its values are stored as settings_records rows by the settings module, under the
// "inventory" module key.
package settings

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The setting names Inventory declares: the unit the template's weight is expressed in, and the
// unit its length, width and height are expressed in. Each holds a UoM id of the organization.
const (
	OrgSettingDefaultWeightUomId = "default_weight_uom_id"
	OrgSettingDefaultLengthUomId = "default_length_uom_id"
)

// OrgSettingScrapSequenceDefinitionId holds the id of the Sequence definition that numbers the
// organization's scrap documents (code ScrapSequenceCode). The org lifecycle writes it; the module's
// settings writer refuses it from a client, because re-pointing it would renumber scraps from
// another definition and hide the real one from the usage check.
const OrgSettingScrapSequenceDefinitionId = "scrap_sequence_definition_id"

// OrgSettingTransferSequenceDefinitionId holds the id of the one Sequence definition that numbers
// the organization's transfers (code TransferSequenceCode), one counter per operation type. Server
// owned for the same reason as the scrap definition.
const OrgSettingTransferSequenceDefinitionId = "transfer_sequence_definition_id"

// ServerOwnedOrgSettings are the settings only Inventory itself writes.
var ServerOwnedOrgSettings = []string{OrgSettingScrapSequenceDefinitionId, OrgSettingTransferSequenceDefinitionId}

// The symbols of the units a new organization starts with, and the units an organization with no
// stored value reads as. UoM ids are per organization, so a default can only name a symbol; the org
// lifecycle resolves it to the organization's own unit.
const (
	DefaultWeightUomSymbol = "kg"
	DefaultLengthUomSymbol = "cm"
)

// DefaultUomSymbolOf maps each unit setting to the symbol of its default unit.
var DefaultUomSymbolOf = map[string]string{
	OrgSettingDefaultWeightUomId: DefaultWeightUomSymbol,
	OrgSettingDefaultLengthUomId: DefaultLengthUomSymbol,
}

// OrgSettingsSchemaName is the name Inventory registers its org-level settings under.
const OrgSettingsSchemaName = "inventory_org_settings"

//go:embed org_settings.json
var orgSettingsSchemaJson string

// OrgSettingsSchemaBuilder declares the units an organization measures its products in. The
// document declares no table_name or should_build_db and extends no base model, because it owns no
// table. There is no default_value: a UoM id is per organization, so the default is written by the
// org lifecycle instead.
func OrgSettingsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgSettingsSchemaJson)
}
