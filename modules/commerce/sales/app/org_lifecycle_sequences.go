package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	modconstants "github.com/sky-as-code/nikki-erp/modules/commerce/sales/constants"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The vocabulary of Essential's sequence definitions Sales registers with.
const (
	sequenceScopeOrg          = "org"
	sequenceFormatOwnerSystem = "sequence"
	sequenceResetNever        = "never"
	sequenceDefaultPadding    = int32(5)
)

// salesSequence is one document series a new organization gets, and the setting its id goes to.
type salesSequence struct {
	setting   string
	code      string
	prefix    string
	algorithm string
	nameEn    string
	nameVi    string
}

// salesSequences are the four series, one counter per organization each. The public reference is
// pseudorandom so a code cannot be derived from its neighbours; Essential forces its padding to 0
// and grows its length on its own.
var salesSequences = []salesSequence{
	{salessettings.OrgSettingOrderNumberSequenceId, "sales.order", "SO",
		itExt.SequenceAlgorithmIncrement, "Sales order number", "Số đơn bán hàng"},
	{salessettings.OrgSettingOrderPublicReferenceSequenceId, "sales.order.public_reference", "CORE",
		itExt.SequenceAlgorithmPseudorandom, "Sales order public reference", "Mã tham chiếu công khai đơn bán hàng"},
	{salessettings.OrgSettingBillNumberSequenceId, "sales.bill", "BILL",
		itExt.SequenceAlgorithmIncrement, "Sales bill number", "Số hóa đơn bán hàng"},
	{salessettings.OrgSettingReturnNumberSequenceId, "sales.return", "RET",
		itExt.SequenceAlgorithmIncrement, "Sales return number", "Số phiếu trả hàng"},
}

// upsertDefinitionCommand builds the registration of one series for one organization.
func (this salesSequence) upsertDefinitionCommand(orgId model.Id) itExt.UpsertSequenceDefinitionCommand {
	padding := sequenceDefaultPadding
	if this.algorithm == itExt.SequenceAlgorithmPseudorandom {
		padding = 0
	}
	return itExt.UpsertSequenceDefinitionCommand{
		ConfigurationScope: sequenceScopeOrg,
		OrgId:              &orgId,
		OwnerModule:        modconstants.SalesModuleName,
		Code:               this.code,
		Name:               model.LangJson{"en-US": this.nameEn, "vi-VN": this.nameVi},
		FormatOwner:        util.ToPtr(sequenceFormatOwnerSystem),
		Algorithm:          util.ToPtr(this.algorithm),
		PrefixTemplate:     util.ToPtr(this.prefix),
		Padding:            &padding,
		Increment:          util.ToPtr(int64(1)),
		InitialValue:       util.ToPtr(int64(1)),
		ResetPeriod:        util.ToPtr(sequenceResetNever),
	}
}

// registerSequences registers the organization's document series and records their ids in its
// Sales settings. Upsert returns an existing definition unchanged, so a re-run lifecycle points the
// settings at the same series instead of starting a new one.
func (this *OrgLifecycleAppServiceImpl) registerSequences(ctx corectx.Context, orgId model.Id) error {
	registrations := make([]itExt.SequenceOrgRegistration, 0, len(salesSequences))
	for _, sequence := range salesSequences {
		registrations = append(registrations, itExt.SequenceOrgRegistration{
			Setting: sequence.setting, Command: sequence.upsertDefinitionCommand(orgId),
		})
	}
	_, err := itExt.RegisterOrgSequences(ctx, this.param.Sequences, this.param.SettingsSeed,
		modconstants.SettingsModuleKey, orgId, registrations...)
	return err
}
