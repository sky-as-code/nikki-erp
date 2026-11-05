package external

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// SequenceExtService is Inventory's port onto Essential's document numbering. It asserts no
// permission: Inventory reaches it only from a write it already authorized, or from the org
// lifecycle. NextValue joins the caller's transaction, so a document that rolls back releases its
// number; a caller that receives client errors must roll back, since the reservation is written.
type SequenceExtService interface {
	UpsertDefinition(ctx corectx.Context, cmd UpsertSequenceDefinitionCommand) (*UpsertSequenceDefinitionResult, error)
	NextValue(ctx corectx.Context, cmd NextSequenceValueCommand) (*NextSequenceValueResult, error)
}

type (
	UpsertSequenceDefinitionCommand = itSeq.UpsertDefinitionCommand
	UpsertSequenceDefinitionResult  = itSeq.UpsertDefinitionResult
	NextSequenceValueCommand        = itSeq.NextValueCommand
	NextSequenceValueResult         = itSeq.NextValueResult
)

// OrgSettingsReaderExtService reads the values one organization stores for a module, without
// asserting a permission: the scrap numbering and the usage check read Inventory's own settings on
// behalf of a caller who need not hold settings access. A missing value is absent from the map; a
// failed read is an error, never an empty map, so "could not read" is never taken for "not set".
type OrgSettingsReaderExtService = coreSettings.OrgSettingsReader

// OrgLanguage is the language an organization's defaults and generated names are written in: its
// essential.system_language, else model.DefaultLanguageCode.
var OrgLanguage = itLang.OrgLanguage

// SequenceOperationKey names one NextValue call deterministically (see itSeq.OperationKey).
var SequenceOperationKey = itSeq.OperationKey

// RegisterOrgSequences registers definitions for a new organization and records their ids in its
// settings. Only the org lifecycle calls it.
var RegisterOrgSequences = itSeq.RegisterOrgDefinitions

type SequenceOrgRegistration = itSeq.OrgRegistration

// SequenceNotConfiguredError refuses a document whose org holds no definition id for it.
var SequenceNotConfiguredError = itSeq.NotConfiguredError
