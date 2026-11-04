package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type fakeSequences struct {
	calls   []itExt.NextSequenceValueCommand
	refuse  bool
	counter int
}

func (this *fakeSequences) UpsertDefinition(
	corectx.Context, itExt.UpsertSequenceDefinitionCommand,
) (*itExt.UpsertSequenceDefinitionResult, error) {
	return nil, nil
}

func (this *fakeSequences) NextValue(
	_ corectx.Context, cmd itExt.NextSequenceValueCommand,
) (*itExt.NextSequenceValueResult, error) {
	this.calls = append(this.calls, cmd)
	if this.refuse {
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewAnonymousBusinessViolation("sequence.archived", "archived"))
		return &itExt.NextSequenceValueResult{ClientErrors: *vErrs}, nil
	}
	this.counter++
	result := &itExt.NextSequenceValueResult{HasData: true}
	result.Data.FormattedValue = fmt.Sprintf("%s-%05d", cmd.DefinitionId, this.counter)
	return result, nil
}

type fakeOrgSettings struct {
	values map[string]any
	orgs   []model.Id
}

func (this *fakeOrgSettings) GetOrgSettingValues(
	_ corectx.Context, orgId model.Id, _ string,
) (map[string]any, error) {
	this.orgs = append(this.orgs, orgId)
	return this.values, nil
}

// installFakeNumbering gives every Sales sequence a definition named after its setting, so an
// issued number reads "<setting>-00001".
func installFakeNumbering(t *testing.T) (*fakeSequences, *fakeOrgSettings) {
	t.Helper()
	sequences := &fakeSequences{}
	values := map[string]any{}
	for _, name := range salessettings.SequenceIdSettings() {
		values[name] = name
	}
	settings := &fakeOrgSettings{values: values}
	previous := documentNumbering
	t.Cleanup(func() { documentNumbering = previous })
	SetDocumentNumbering(DocumentNumbering{Sequences: sequences, Settings: settings})
	return sequences, settings
}

func TestIssueDocumentNumbersKeysEachIssuanceByFieldAndSource(t *testing.T) {
	sequences, settings := installFakeNumbering(t)
	ctx := corectx.NewRequestContext(context.Background())

	numbers, err := issueDocumentNumbers(ctx, "org1", []string{"order", "src1"}, SeqOrderNumber, SeqOrderPublicReference)

	require.NoError(t, err)
	require.Equal(t, []string{"order_number_sequence_id-00001", "order_public_reference_sequence_id-00002"}, numbers)
	require.Equal(t, []model.Id{"org1"}, settings.orgs, "the settings are read once, for the record's org")
	require.Len(t, sequences.calls, 2)
	require.Equal(t, "order_number:order:src1", *sequences.calls[0].OperationKey)
	require.Equal(t, "public_reference:order:src1", *sequences.calls[1].OperationKey)
	require.Nil(t, sequences.calls[0].IdempotencyId, "the port takes the request's idempotency id")
	require.Equal(t, model.Id("org1"), *sequences.calls[0].OrgId)
	require.Equal(t, "sales", sequences.calls[0].CallerModule)
}

func TestIssueDocumentNumbersRefusesAnOrgWithoutTheSequence(t *testing.T) {
	_, settings := installFakeNumbering(t)
	delete(settings.values, salessettings.OrgSettingReturnNumberSequenceId)

	_, err := issueDocumentNumbers(corectx.NewRequestContext(context.Background()), "org1", nil, SeqReturnNumber)

	cErrs := numberingRefusalOf(err)
	require.NotNil(t, cErrs)
	require.Len(t, *cErrs, 1)
	require.Equal(t, "sales:sequence_not_configured", (*cErrs)[0].Key)
}

// A refusal must surface as an error: NextValue may already have reserved inside the caller's
// transaction, and only an error rolls that back. The entry points still answer it as client errors.
func TestIssueDocumentNumbersTurnsARefusalIntoAnError(t *testing.T) {
	sequences, _ := installFakeNumbering(t)
	sequences.refuse = true

	_, err := issueDocumentNumbers(corectx.NewRequestContext(context.Background()), "org1", nil, SeqBillNumber)

	require.Error(t, err)
	require.NotNil(t, numberingRefusalOf(errors.Wrap(err, "withTransaction")))
}

func TestIssueDocumentNumbersNeedsTheNumberingInstalled(t *testing.T) {
	previous := documentNumbering
	t.Cleanup(func() { documentNumbering = previous })
	documentNumbering = nil

	_, err := issueDocumentNumbers(corectx.NewRequestContext(context.Background()), "org1", nil, SeqBillNumber)

	require.ErrorContains(t, err, "not installed")
}

// A retry of the same request (same ctx idempotency id, same input) keys the same issuance, so the
// port replays its number instead of drawing a second one.
func TestIssueDocumentNumbersKeysARetryTheSame(t *testing.T) {
	sequences, _ := installFakeNumbering(t)
	ctx := corectx.NewRequestContext(context.Background())

	_, err := issueDocumentNumbers(ctx, "org1", nil, SeqOrderNumber)
	require.NoError(t, err)
	_, err = issueDocumentNumbers(ctx, "org1", nil, SeqOrderNumber)
	require.NoError(t, err)

	require.Equal(t, "order_number", *sequences.calls[0].OperationKey)
	require.Equal(t, *sequences.calls[0].OperationKey, *sequences.calls[1].OperationKey)
}
