package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/infra/sequenceusage"
	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// SequenceUsageChecker answers whether Inventory still numbers documents with a sequence
// definition: the org's transfer or scrap definition, both held in Inventory's org settings.
func SequenceUsageChecker() usagecheck.ResourceUsageChecker {
	return sequenceusage.NewSettingsChecker(resolveSettingsReader, invconstants.InventoryModuleName,
		invsettings.OrgSettingTransferSequenceDefinitionId, invsettings.OrgSettingScrapSequenceDefinitionId)
}

func resolveSettingsReader() (sequenceusage.OrgSettingsReader, error) {
	var reader sequenceusage.OrgSettingsReader
	err := deps.Invoke(func(found itExt.OrgSettingsReaderExtService) {
		reader = found
	})
	return reader, err
}
