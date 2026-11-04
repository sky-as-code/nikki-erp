package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	pursettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// CodeSequence is one Purchase code series: the Essential definition it is registered as, the org
// setting its id is kept in, and the field it numbers.
type CodeSequence struct {
	Setting string
	Field   string
	Code    string
	Prefix  string
	NameEn  string
	NameVi  string
}

var (
	SeqOrderCode = CodeSequence{
		Setting: pursettings.OrgSettingOrderCodeSequenceId, Field: models.PurchaseOrderFieldCode,
		Code: "purchase.order", Prefix: "PO", NameEn: "Purchase order code", NameVi: "Mã đơn mua hàng",
	}
	SeqAgreementCode = CodeSequence{
		Setting: pursettings.OrgSettingAgreementCodeSequenceId, Field: models.AgreementFieldCode,
		Code: "purchase.agreement", Prefix: "PA", NameEn: "Purchase agreement code",
		NameVi: "Mã thỏa thuận mua hàng",
	}
)

// CodeSequences are the series every organization gets, one counter per organization each.
func CodeSequences() []CodeSequence {
	return []CodeSequence{SeqOrderCode, SeqAgreementCode}
}

// The vocabulary of Essential's sequence definitions Purchase registers with.
const (
	sequenceScopeOrg          = "org"
	sequenceFormatOwnerSystem = "sequence"
	sequenceResetNever        = "never"
	sequenceDefaultPadding    = int32(5)
)

// UpsertCommand builds the registration of the series for one organization: PO00001, PA00001.
func (this CodeSequence) UpsertCommand(orgId model.Id) itExt.UpsertSequenceDefinitionCommand {
	return itExt.UpsertSequenceDefinitionCommand{
		ConfigurationScope: sequenceScopeOrg,
		OrgId:              &orgId,
		OwnerModule:        purconstants.PurchaseModuleName,
		Code:               this.Code,
		Name:               model.LangJson{"en-US": this.NameEn, "vi-VN": this.NameVi},
		FormatOwner:        util.ToPtr(sequenceFormatOwnerSystem),
		Algorithm:          util.ToPtr(itExt.SequenceAlgorithmIncrement),
		PrefixTemplate:     util.ToPtr(this.Prefix),
		Padding:            util.ToPtr(sequenceDefaultPadding),
		Increment:          util.ToPtr(int64(1)),
		InitialValue:       util.ToPtr(int64(1)),
		ResetPeriod:        util.ToPtr(sequenceResetNever),
	}
}

// OrgSequenceRegistrations are the series the org lifecycle registers for a new organization, with
// the settings their ids are recorded in.
func OrgSequenceRegistrations(orgId model.Id) []itExt.SequenceOrgRegistration {
	registrations := make([]itExt.SequenceOrgRegistration, 0, len(CodeSequences()))
	for _, sequence := range CodeSequences() {
		registrations = append(registrations, itExt.SequenceOrgRegistration{
			Setting: sequence.Setting, Command: sequence.UpsertCommand(orgId),
		})
	}
	return registrations
}

// CodeNumbering issues Purchase's order and agreement codes from the Essential definitions whose
// ids the organization keeps in its Purchase settings. It never registers one: that is the org
// lifecycle's (or the seeds') job alone.
type CodeNumbering struct {
	sequences itExt.SequenceExtService
	reader    itExt.OrgSettingsReaderExtService
}

func NewCodeNumbering(sequences itExt.SequenceExtService, reader itExt.OrgSettingsReaderExtService) *CodeNumbering {
	return &CodeNumbering{sequences: sequences, reader: reader}
}

// Next issues the next code of the series. ctx must carry the write's transaction: NextValue joins
// it, so a refused write gives the code back. Client errors (numbering not set up, the definition
// archived) must make the caller roll back, since the reservation is already written. The operation
// key is the field alone (plus the batch row in a bulk create), so a retry replays its code.
func (this *CodeNumbering) Next(
	ctx corectx.Context, sequence CodeSequence, orgId model.Id,
) (string, *ft.ClientErrors, error) {
	if orgId == "" {
		return "", nil, errors.Errorf("issue %s: the record names no organization", sequence.Field)
	}
	ids, err := SequenceIdsOf(ctx, this.reader, orgId)
	if err != nil {
		return "", nil, err
	}
	definitionId := ids[sequence.Setting]
	if definitionId == "" {
		cErrs := ft.ClientErrors{*itExt.SequenceNotConfiguredError(purconstants.PurchaseModuleName, sequence.Setting)}
		return "", &cErrs, nil
	}
	operationKey := itExt.SequenceOperationKey(ctx, sequence.Field)
	result, err := this.sequences.NextValue(ctx, itExt.NextSequenceValueCommand{
		DefinitionId: definitionId,
		OrgId:        &orgId,
		CallerModule: purconstants.PurchaseModuleName,
		OperationKey: &operationKey,
	})
	if err != nil {
		return "", nil, errors.Wrapf(err, "issue %s", sequence.Field)
	}
	if result.ClientErrors.Count() > 0 {
		return "", &result.ClientErrors, nil
	}
	if !result.HasData || result.Data.FormattedValue == "" {
		return "", nil, errors.Errorf("sequence %s issued no value", definitionId)
	}
	return result.Data.FormattedValue, nil, nil
}

// SequenceIdsOf reads the sequence definition ids one organization's Purchase settings hold, keyed
// by setting name. An unset id is absent. A failed read is an error.
func SequenceIdsOf(
	ctx corectx.Context, reader itExt.OrgSettingsReaderExtService, orgId model.Id,
) (map[string]model.Id, error) {
	if reader == nil {
		return nil, errors.New("SequenceIdsOf: no settings reader")
	}
	values, err := reader.GetOrgSettingValues(ctx, orgId, purconstants.PurchaseModuleName)
	if err != nil {
		return nil, err
	}
	ids := map[string]model.Id{}
	for _, name := range pursettings.SequenceIdSettings() {
		if id, ok := values[name].(string); ok && id != "" {
			ids[name] = model.Id(id)
		}
	}
	return ids, nil
}
