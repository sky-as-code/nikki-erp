package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

type fakeSequences struct {
	upserts []itExt.UpsertSequenceDefinitionCommand
	nexts   []itExt.NextSequenceValueCommand
	refuse  bool
}

func (this *fakeSequences) UpsertDefinition(
	_ corectx.Context, cmd itExt.UpsertSequenceDefinitionCommand,
) (*itExt.UpsertSequenceDefinitionResult, error) {
	this.upserts = append(this.upserts, cmd)
	result := &itExt.UpsertSequenceDefinitionResult{HasData: true}
	result.Data.DefinitionId = "01DEFREGISTERED"
	return result, nil
}

func (this *fakeSequences) NextValue(
	_ corectx.Context, cmd itExt.NextSequenceValueCommand,
) (*itExt.NextSequenceValueResult, error) {
	this.nexts = append(this.nexts, cmd)
	if this.refuse {
		refusal := ft.NewClientErrors()
		refusal.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey("sequence.archived"), "refused"))
		return &itExt.NextSequenceValueResult{ClientErrors: *refusal}, nil
	}
	result := &itExt.NextSequenceValueResult{HasData: true}
	result.Data.FormattedValue = "INV-2026-000043"
	result.Data.EffectiveAt = "2026-11-04T09:30:00Z"
	return result, nil
}

type fakeSettings struct {
	values map[string]any
}

func (this *fakeSettings) GetOrgSettingValues(_ corectx.Context, _ model.Id, moduleKey string) (map[string]any, error) {
	if moduleKey != modconstants.PaymentInvoiceModuleName {
		return map[string]any{}, nil
	}
	return this.values, nil
}

func newTestNumbering(values map[string]any) (*InvoiceNumbering, *fakeSequences) {
	sequences := &fakeSequences{}
	return NewInvoiceNumbering(sequences, &fakeSettings{values: values}), sequences
}

// The invoice is issued at the instant Sequence accepted the number, so the year in the number is
// the year it was issued in, and the request sends no time of its own: a retry fingerprints the same.
func TestNextIssuesAtTheSequencesInstantWithADeterministicKey(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{
		pinvsettings.OrgSettingInvoiceNumberSequenceId: "01DEFINVOICE",
	})
	ctx := corectx.NewRequestContext(t.Context())

	issued, cErrs, err := numbering.Next(ctx, "01ORG", "01INVOICE")
	require.NoError(t, err)
	require.Nil(t, cErrs)
	_, _, err = numbering.Next(ctx, "01ORG", "01INVOICE")
	require.NoError(t, err)

	assert.Equal(t, "INV-2026-000043", issued.Number)
	assert.Equal(t, time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC), issued.IssuedAt)
	require.Len(t, sequences.nexts, 2)
	sent := sequences.nexts[0]
	assert.Equal(t, model.Id("01DEFINVOICE"), sent.DefinitionId)
	assert.Equal(t, "number:01INVOICE", *sent.OperationKey)
	assert.Equal(t, *sent.OperationKey, *sequences.nexts[1].OperationKey)
	assert.Nil(t, sent.EffectiveAt)
	assert.Nil(t, sent.IdempotencyId, "the port takes the request's idempotency id")
	assert.Equal(t, modconstants.PaymentInvoiceModuleName, sent.CallerModule)
	assert.Empty(t, sequences.upserts)
}

// An org without its series is refused; issuing never registers a definition.
func TestNextRefusesAnOrgWithoutTheSeries(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{})

	_, cErrs, err := numbering.Next(corectx.NewRequestContext(t.Context()), "01ORG", "01INVOICE")

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Equal(t, "paymentinvoice:sequence_not_configured", (*cErrs)[0].Key)
	assert.Empty(t, sequences.upserts)
	assert.Empty(t, sequences.nexts)
}

func TestNextReturnsARefusalAsClientErrors(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{
		pinvsettings.OrgSettingInvoiceNumberSequenceId: "01DEFINVOICE",
	})
	sequences.refuse = true

	issued, cErrs, err := numbering.Next(corectx.NewRequestContext(t.Context()), "01ORG", "01INVOICE")

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Nil(t, issued)
}

func TestNextRefusesAnInvoiceWithoutOrganization(t *testing.T) {
	numbering, _ := newTestNumbering(map[string]any{})

	_, _, err := numbering.Next(corectx.NewRequestContext(t.Context()), "", "01INVOICE")

	assert.Error(t, err)
}

// The registration pins the format quoted on invoices today: INV-{year}-{6 digits}, yearly.
func TestTheInvoiceSeriesKeepsTodaysFormat(t *testing.T) {
	cmd := InvoiceSequenceCommand("01ORG")

	assert.Equal(t, "org", cmd.ConfigurationScope)
	assert.Equal(t, InvoiceSequenceCode, cmd.Code)
	assert.Equal(t, "INV-%(range_year)s-", *cmd.PrefixTemplate)
	assert.Equal(t, int32(6), *cmd.Padding)
	assert.Equal(t, "yearly", *cmd.ResetPeriod)
	assert.Equal(t, "sequence", *cmd.FormatOwner)
	assert.Nil(t, cmd.Timezone, "Essential copies the org's time zone")
}
