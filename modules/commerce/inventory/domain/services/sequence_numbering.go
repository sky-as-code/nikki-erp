package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	invconstants "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/constants"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// Inventory's document numbering runs on Essential's Sequence capability. Each organization gets two
// definitions from its org lifecycle (or the seeds), never while a document is processed, both at
// org scope, owned by "inventory", never resetting, starting at 1, stepping by 1, padded to 5:
//
//   - inventory.transfer, sub-scope inventory_optype, consumer-formatted: every operation type is a
//     scope object with its own counter, the port returns the padded number and Inventory writes
//     {warehouse code}/{sequence prefix}/{number};
//   - inventory.scrap, Sequence-formatted with prefix SCR.
//
// Their ids are kept in Inventory's org settings.
const (
	ScrapSequenceCode       = "inventory.scrap"
	TransferSequenceCode    = "inventory.transfer"
	scrapSequencePrefix     = "SCR"
	optypeSequenceSubScope  = "inventory_optype"
	sequenceConfigScopeOrg  = "org"
	sequenceFormatConsumer  = "consumer"
	sequenceFormatSequence  = "sequence"
	sequenceResetNever      = "never"
	sequencePadding         = int32(5)
	sequenceStep            = int64(1)
	sequenceFirstNumber     = int64(1)
	transferNumberSeparator = "/"
)

func baseSequenceCommand(orgId model.Id, code string, name model.LangJson) itExt.UpsertSequenceDefinitionCommand {
	return itExt.UpsertSequenceDefinitionCommand{
		ConfigurationScope: sequenceConfigScopeOrg,
		OrgId:              &orgId,
		OwnerModule:        invconstants.InventoryModuleName,
		Code:               code,
		Names:              name,
		Padding:            util.ToPtr(sequencePadding),
		Increment:          util.ToPtr(sequenceStep),
		InitialValue:       util.ToPtr(sequenceFirstNumber),
		ResetPeriod:        util.ToPtr(sequenceResetNever),
	}
}

// TransferSequenceCommand registers the definition numbering an organization's transfers, one
// counter per operation type.
func TransferSequenceCommand(orgId model.Id) itExt.UpsertSequenceDefinitionCommand {
	cmd := baseSequenceCommand(orgId, TransferSequenceCode, model.LangJson{
		model.LanguageCode("en-US"): "Transfer documents",
		model.LanguageCode("vi-VN"): "Phiếu chuyển kho",
	})
	cmd.SubScope = util.ToPtr(optypeSequenceSubScope)
	cmd.FormatOwner = util.ToPtr(sequenceFormatConsumer)
	return cmd
}

// ScrapSequenceCommand registers the definition numbering an organization's scraps.
func ScrapSequenceCommand(orgId model.Id) itExt.UpsertSequenceDefinitionCommand {
	cmd := baseSequenceCommand(orgId, ScrapSequenceCode, model.LangJson{
		model.LanguageCode("en-US"): "Scrap documents",
		model.LanguageCode("vi-VN"): "Phiếu hủy hàng",
	})
	cmd.FormatOwner = util.ToPtr(sequenceFormatSequence)
	cmd.PrefixTemplate = util.ToPtr(scrapSequencePrefix)
	return cmd
}

// OrgSequenceRegistrations are the definitions the org lifecycle registers for a new organization,
// with the settings their ids are recorded in.
func OrgSequenceRegistrations(orgId model.Id) []itExt.SequenceOrgRegistration {
	return []itExt.SequenceOrgRegistration{
		{Setting: invsettings.OrgSettingTransferSequenceDefinitionId, Command: TransferSequenceCommand(orgId)},
		{Setting: invsettings.OrgSettingScrapSequenceDefinitionId, Command: ScrapSequenceCommand(orgId)},
	}
}

// FormatTransferNumber writes {warehouse code}/{sequence prefix}/{padded number}. A segment that is
// empty (an operation type predating the prefix or the warehouse) is left out rather than doubled.
func FormatTransferNumber(warehouseCode string, sequencePrefix string, paddedNumber string) string {
	parts := make([]string, 0, 3)
	for _, part := range []string{warehouseCode, sequencePrefix, paddedNumber} {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, transferNumberSeparator)
}

// sequenceDefinitionOf reads the definition id an org setting holds. Unset is a client error: the
// org's lifecycle never ran or the seeds missed it, and a definition is never registered here.
func sequenceDefinitionOf(
	ctx corectx.Context, reader itExt.OrgSettingsReaderExtService, orgId model.Id, setting string,
) (model.Id, *ft.ClientErrors, error) {
	if reader == nil {
		return "", nil, errors.New("sequenceDefinitionOf: no settings reader")
	}
	values, err := reader.GetOrgSettingValues(ctx, orgId, invconstants.InventoryModuleName)
	if err != nil {
		return "", nil, err
	}
	if stored, ok := values[setting].(string); ok && stored != "" {
		return model.Id(stored), nil, nil
	}
	cErrs := ft.ClientErrors{*itExt.SequenceNotConfiguredError(invconstants.InventoryModuleName, setting)}
	return "", &cErrs, nil
}

// sequenceRequest names one number to issue: the definition, the org, the custom-scope object if
// the definition has one, and the operation key that tells this request's numbers apart.
type sequenceRequest struct {
	definitionId  model.Id
	orgId         model.Id
	scopeObjectId *model.Id
	operationKey  string
}

// nextSequenceValue issues one number and returns the formatted value: the whole number for a
// Sequence-formatted definition, the padded number alone for a consumer-formatted one. Client
// errors (an archived definition, an exhausted counter) are returned for the caller to roll back on.
// IdempotencyId is left nil so the port takes the request's own from ctx.
func nextSequenceValue(
	ctx corectx.Context, sequences itExt.SequenceExtService, req sequenceRequest,
) (string, *ft.ClientErrors, error) {
	result, err := sequences.NextValue(ctx, itExt.NextSequenceValueCommand{
		DefinitionId:  req.definitionId,
		OrgId:         &req.orgId,
		CallerModule:  invconstants.InventoryModuleName,
		ScopeObjectId: req.scopeObjectId,
		OperationKey:  &req.operationKey,
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "issue sequence number")
	}
	if result.ClientErrors.Count() > 0 {
		return "", &result.ClientErrors, nil
	}
	if !result.HasData || result.Data.FormattedValue == "" {
		return "", nil, errors.Errorf("sequence %s issued no value", req.definitionId)
	}
	return result.Data.FormattedValue, nil, nil
}
