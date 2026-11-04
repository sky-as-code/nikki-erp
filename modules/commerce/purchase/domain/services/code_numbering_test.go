package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	purconstants "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	pursettings "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type fakeSequences struct {
	upserts   []itExt.UpsertSequenceDefinitionCommand
	nexts     []itExt.NextSequenceValueCommand
	refuseKey string
}

func (this *fakeSequences) UpsertDefinition(
	_ corectx.Context, cmd itExt.UpsertSequenceDefinitionCommand,
) (*itExt.UpsertSequenceDefinitionResult, error) {
	this.upserts = append(this.upserts, cmd)
	result := &itExt.UpsertSequenceDefinitionResult{HasData: true}
	result.Data.DefinitionId = model.Id("DEF-" + cmd.Code)
	return result, nil
}

func (this *fakeSequences) NextValue(
	_ corectx.Context, cmd itExt.NextSequenceValueCommand,
) (*itExt.NextSequenceValueResult, error) {
	this.nexts = append(this.nexts, cmd)
	if this.refuseKey != "" {
		refusal := ft.NewClientErrors()
		refusal.Append(*ft.NewAnonymousBusinessViolation(ft.ErrorKey(this.refuseKey), "refused"))
		return &itExt.NextSequenceValueResult{ClientErrors: *refusal}, nil
	}
	result := &itExt.NextSequenceValueResult{HasData: true}
	result.Data.FormattedValue = "PO00007"
	return result, nil
}

type fakeSettings struct {
	values map[string]any
}

func (this *fakeSettings) GetOrgSettingValues(_ corectx.Context, _ model.Id, moduleKey string) (map[string]any, error) {
	if moduleKey != purconstants.PurchaseModuleName {
		return map[string]any{}, nil
	}
	return this.values, nil
}

func newTestNumbering(values map[string]any) (*CodeNumbering, *fakeSequences) {
	sequences := &fakeSequences{}
	return NewCodeNumbering(sequences, &fakeSettings{values: values}), sequences
}

func TestNextIssuesFromTheSettingDefinition(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{
		pursettings.OrgSettingOrderCodeSequenceId: "01DEFORDER",
	})
	ctx := corectx.NewRequestContext(t.Context())

	code, cErrs, err := numbering.Next(ctx, SeqOrderCode, "01ORG")
	require.NoError(t, err)
	require.Nil(t, cErrs)
	assert.Equal(t, "PO00007", code)
	_, _, err = numbering.Next(ctx, SeqOrderCode, "01ORG")
	require.NoError(t, err)

	require.Len(t, sequences.nexts, 2)
	assert.Equal(t, model.Id("01DEFORDER"), sequences.nexts[0].DefinitionId)
	assert.Equal(t, "code", *sequences.nexts[0].OperationKey)
	assert.Equal(t, *sequences.nexts[0].OperationKey, *sequences.nexts[1].OperationKey, "a retry keys the same issuance")
	assert.Nil(t, sequences.nexts[0].IdempotencyId, "the port takes the request's idempotency id")
	assert.Equal(t, purconstants.PurchaseModuleName, sequences.nexts[0].CallerModule)
	assert.Empty(t, sequences.upserts)
}

// An org without its series is refused; numbering never registers a definition.
func TestNextRefusesAnOrgWithoutTheSeries(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{})

	_, cErrs, err := numbering.Next(corectx.NewRequestContext(t.Context()), SeqAgreementCode, "01ORG")

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Equal(t, "purchase:sequence_not_configured", (*cErrs)[0].Key)
	assert.Empty(t, sequences.upserts)
	assert.Empty(t, sequences.nexts)
}

func TestOrgSequenceRegistrationsCoverBothSeries(t *testing.T) {
	registrations := OrgSequenceRegistrations("01ORG")

	require.Len(t, registrations, 2)
	assert.Equal(t, pursettings.OrgSettingOrderCodeSequenceId, registrations[0].Setting)
	assert.Equal(t, "purchase.order", registrations[0].Command.Code)
	assert.Equal(t, pursettings.OrgSettingAgreementCodeSequenceId, registrations[1].Setting)
}

func TestNextReturnsRefusalAsClientErrors(t *testing.T) {
	numbering, sequences := newTestNumbering(map[string]any{
		pursettings.OrgSettingOrderCodeSequenceId: "01DEFORDER",
	})
	sequences.refuseKey = "sequence.archived"

	code, cErrs, err := numbering.Next(corectx.NewRequestContext(t.Context()), SeqOrderCode, "01ORG")

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Equal(t, 1, cErrs.Count())
	assert.Empty(t, code)
}

func TestUpsertCommandMatchesThePlan(t *testing.T) {
	cmd := SeqOrderCode.UpsertCommand("01ORG")

	assert.Equal(t, "org", cmd.ConfigurationScope)
	assert.Equal(t, "purchase.order", cmd.Code)
	assert.Equal(t, "PO", *cmd.PrefixTemplate)
	assert.Equal(t, int32(5), *cmd.Padding)
	assert.Equal(t, "never", *cmd.ResetPeriod)
	assert.Equal(t, "sequence", *cmd.FormatOwner)
	assert.Equal(t, itExt.SequenceAlgorithmIncrement, *cmd.Algorithm)
	assert.Equal(t, "PA", *SeqAgreementCode.UpsertCommand("01ORG").PrefixTemplate)
}
