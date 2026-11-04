package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/infra/sequenceusage"
	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	pursettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// SequenceUsageChecker answers whether Purchase still numbers documents with a sequence definition:
// it does while the org's Purchase settings hold the definition's id (the order or agreement code
// series).
func SequenceUsageChecker() usagecheck.ResourceUsageChecker {
	return sequenceusage.NewSettingsChecker(resolveSettingsReader, purconstants.PurchaseModuleName,
		pursettings.SequenceIdSettings()...)
}

// resolveSettingsReader resolves at check time: checkers are registered during Init, before the
// container can build the settings services.
func resolveSettingsReader() (sequenceusage.OrgSettingsReader, error) {
	var reader sequenceusage.OrgSettingsReader
	err := deps.Invoke(func(found itExt.OrgSettingsReaderExtService) {
		reader = found
	})
	return reader, err
}
