package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

// InvoiceSequenceCode is the code the invoice series is registered under, once per organization.
const InvoiceSequenceCode = "paymentinvoice.invoice"

// The registration of the invoice series: INV-2026-000001, restarting every year in the
// organization's time zone. The year is the period's own (range_year), so the prefix always
// names the year the counter belongs to. The time zone is left to Essential, which copies the
// acting user's effective system time zone and falls back to UTC.
const (
	invoicePrefixTemplate = "INV-%(range_year)s-"
	invoicePadding        = int32(6)
	sequenceScopeOrg      = "org"
	sequenceFormatOwner   = "sequence"
	sequenceResetYearly   = "yearly"
)

// InvoiceSequenceCommand builds the registration of one organization's invoice series.
func InvoiceSequenceCommand(orgId model.Id) itExt.UpsertSequenceDefinitionCommand {
	return itExt.UpsertSequenceDefinitionCommand{
		ConfigurationScope: sequenceScopeOrg,
		OrgId:              &orgId,
		OwnerModule:        modconstants.PaymentInvoiceModuleName,
		Code:               InvoiceSequenceCode,
		Name:               model.LangJson{"en-US": "Invoice number", "vi-VN": "Số hóa đơn"},
		FormatOwner:        util.ToPtr(sequenceFormatOwner),
		Algorithm:          util.ToPtr(itExt.SequenceAlgorithmIncrement),
		PrefixTemplate:     util.ToPtr(invoicePrefixTemplate),
		Padding:            util.ToPtr(invoicePadding),
		Increment:          util.ToPtr(int64(1)),
		InitialValue:       util.ToPtr(int64(1)),
		ResetPeriod:        util.ToPtr(sequenceResetYearly),
	}
}

// InvoiceNumbering issues invoice numbers from the Essential definition whose id the organization
// keeps in its Payment & Invoice settings. It never registers one: that is the org lifecycle's (or
// the seeds') job alone.
type InvoiceNumbering struct {
	sequences itExt.SequenceExtService
	reader    itExt.OrgSettingsReaderExtService
}

func NewInvoiceNumbering(
	sequences itExt.SequenceExtService, reader itExt.OrgSettingsReaderExtService,
) *InvoiceNumbering {
	return &InvoiceNumbering{sequences: sequences, reader: reader}
}

// IssuedNumber is an invoice's number and the instant it was issued for.
type IssuedNumber struct {
	Number   string
	IssuedAt time.Time
}

// Next issues the number of one invoice. It sends no effective time: Sequence stamps its own
// acceptance time and returns it, and that instant becomes the invoice's issued_at, so the year in
// the number is the year it was issued in and a retry fingerprints the same (a time taken here would
// differ per attempt). The key names the invoice being issued, an existing record. ctx must carry
// the issue transaction: NextValue joins it, so a rolled-back issue gives the number back. Client
// errors (numbering not set up, the definition archived) must make the caller roll back.
func (this *InvoiceNumbering) Next(
	ctx corectx.Context, orgId model.Id, invoiceId string,
) (*IssuedNumber, *ft.ClientErrors, error) {
	if orgId == "" {
		return nil, nil, errors.Errorf("number invoice %s: it names no organization", invoiceId)
	}
	definitionId, err := InvoiceSequenceIdOf(ctx, this.reader, orgId)
	if err != nil {
		return nil, nil, err
	}
	if definitionId == "" {
		cErrs := ft.ClientErrors{*itExt.SequenceNotConfiguredError(
			modconstants.PaymentInvoiceModuleName, pinvsettings.OrgSettingInvoiceNumberSequenceId)}
		return nil, &cErrs, nil
	}
	operationKey := itExt.SequenceOperationKey(ctx, models.InvoiceFieldNumber, invoiceId)
	result, err := this.sequences.NextValue(ctx, itExt.NextSequenceValueCommand{
		DefinitionId: definitionId,
		OrgId:        &orgId,
		CallerModule: modconstants.PaymentInvoiceModuleName,
		OperationKey: &operationKey,
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "issue invoice number")
	}
	if result.ClientErrors.Count() > 0 {
		return nil, &result.ClientErrors, nil
	}
	if !result.HasData {
		return nil, nil, errors.Errorf("sequence %s issued no value", definitionId)
	}
	return issuedNumberOf(result.Data, definitionId)
}

func issuedNumberOf(
	data itExt.NextSequenceValueResultData, definitionId model.Id,
) (*IssuedNumber, *ft.ClientErrors, error) {
	if data.FormattedValue == "" {
		return nil, nil, errors.Errorf("sequence %s issued no value", definitionId)
	}
	issuedAt, err := model.ParseModelDateTime(data.EffectiveAt)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "sequence %s issued an unreadable effective_at", definitionId)
	}
	return &IssuedNumber{Number: data.FormattedValue, IssuedAt: issuedAt.GoTime()}, nil, nil
}

// InvoiceSequenceIdOf reads the invoice sequence id the organization's settings hold; empty when
// unset. A failed read is an error, so the usage check never mistakes it for "not held".
func InvoiceSequenceIdOf(
	ctx corectx.Context, reader itExt.OrgSettingsReaderExtService, orgId model.Id,
) (model.Id, error) {
	if reader == nil {
		return "", errors.New("InvoiceSequenceIdOf: no settings reader")
	}
	values, err := reader.GetOrgSettingValues(ctx, orgId, modconstants.PaymentInvoiceModuleName)
	if err != nil {
		return "", err
	}
	held, _ := values[pinvsettings.OrgSettingInvoiceNumberSequenceId].(string)
	return model.Id(held), nil
}
