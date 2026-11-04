package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
	itSettings "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SequenceExtService is Sales' port onto Essential's document numbering. It asserts no permission:
// Sales reaches it only from a write it already authorized, or from the org lifecycle. NextValue
// joins the caller's transaction, so a sale that rolls back releases its number.
type SequenceExtService interface {
	UpsertDefinition(ctx corectx.Context, cmd UpsertSequenceDefinitionCommand) (*UpsertSequenceDefinitionResult, error)
	NextValue(ctx corectx.Context, cmd NextSequenceValueCommand) (*NextSequenceValueResult, error)
}

type UpsertSequenceDefinitionCommand = itSeq.UpsertDefinitionCommand
type UpsertSequenceDefinitionResult = itSeq.UpsertDefinitionResult
type NextSequenceValueCommand = itSeq.NextValueCommand
type NextSequenceValueResult = itSeq.NextValueResult

// The algorithm spellings the port accepts.
const (
	SequenceAlgorithmIncrement    = itSeq.AlgorithmIncrement
	SequenceAlgorithmPseudorandom = itSeq.AlgorithmPseudorandom
)

// OrgSettingsReaderExtService reads the values one organization stores for a module, without
// asserting a permission and without guessing the organization from the request. A missing value is
// absent from the map; a failed read is an error, never an empty map, because the usage check must
// not mistake "could not read" for "nothing stored".
type OrgSettingsReaderExtService = coreSettings.OrgSettingsReader

// SequenceOperationKey names one NextValue call deterministically (see itSeq.OperationKey).
var SequenceOperationKey = itSeq.OperationKey

// RegisterOrgSequences registers definitions for a new organization and records their ids in its
// settings. Only the org lifecycle calls it.
var RegisterOrgSequences = itSeq.RegisterOrgDefinitions

type SequenceOrgRegistration = itSeq.OrgRegistration

// SequenceNotConfiguredError refuses a document whose org holds no definition id for it.
var SequenceNotConfiguredError = itSeq.NotConfiguredError

// OrgSettingsSeedExtService writes, from the org lifecycle, the settings that point at what the
// lifecycle just created for a new organization.
type OrgSettingsSeedExtService interface {
	SeedOrgSettings(ctx corectx.Context, cmd SetSettingsCommand) (*SetSettingsResult, error)
}

type SetSettingsCommand = itSettings.SetSettingsCommand
type SetSettingItem = itSettings.SetSettingItem
type SetSettingsResult = itSettings.SetSettingsResult
