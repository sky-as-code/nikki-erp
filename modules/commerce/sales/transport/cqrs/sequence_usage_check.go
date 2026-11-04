package cqrs

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/model"
	salesconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// SequenceUsageChecker answers whether Sales still numbers documents with a sequence definition:
// it does while any of its org settings holds the definition's id. Sales' definitions are org
// scoped, so a definition named without an org is none of Sales' and is not in use here.
//
// The settings reader is resolved at check time, like the repositories of the table checkers:
// checkers are registered during Init, before every port is guaranteed to be built.
func SequenceUsageChecker() usagecheck.ResourceUsageChecker {
	return usagecheck.CheckerFunc(func(ctx corectx.Context, ref usagecheck.ResourceRef) (bool, string, error) {
		var reader itExt.OrgSettingsReaderExtService
		if err := deps.Invoke(func(found itExt.OrgSettingsReaderExtService) { reader = found }); err != nil {
			return false, "", err
		}
		return sequenceUsedBySettings(ctx, reader, ref)
	})
}

func sequenceUsedBySettings(
	ctx corectx.Context, reader itExt.OrgSettingsReaderExtService, ref usagecheck.ResourceRef,
) (bool, string, error) {
	id := ref.Identifier[usagecheck.IdentifierKeyId]
	orgId := ref.Identifier[usagecheck.IdentifierKeyOrgId]
	if id == "" || orgId == "" {
		return false, "", nil
	}
	ids, err := services.SequenceIdsOf(ctx, reader, model.Id(orgId))
	if err != nil {
		return false, "", err
	}
	for setting, held := range ids {
		if string(held) == id {
			return true, salesconstants.SalesModuleName + "." + setting, nil
		}
	}
	return false, "", nil
}
