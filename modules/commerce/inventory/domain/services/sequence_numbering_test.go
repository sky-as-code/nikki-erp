package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	invsettings "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/external"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

const (
	testOrgId       = "01JWNY20G23KD4RV5VWYABQYHD"
	testOptypeId    = "01K5INV00000000CENRECEIPTS"
	testScrapDefId  = "01KSEQ00000000000000000002"
	testPaddedValue = "00007"
)

// fakeSequences records what Inventory asks of Sequence.
type fakeSequences struct {
	upserts []itExt.UpsertSequenceDefinitionCommand
	nexts   []itExt.NextSequenceValueCommand
}

func (this *fakeSequences) UpsertDefinition(
	_ corectx.Context, cmd itExt.UpsertSequenceDefinitionCommand,
) (*itExt.UpsertSequenceDefinitionResult, error) {
	this.upserts = append(this.upserts, cmd)
	result := &itExt.UpsertSequenceDefinitionResult{HasData: true}
	result.Data.DefinitionId = testScrapDefId
	return result, nil
}

func (this *fakeSequences) NextValue(
	_ corectx.Context, cmd itExt.NextSequenceValueCommand,
) (*itExt.NextSequenceValueResult, error) {
	this.nexts = append(this.nexts, cmd)
	result := &itExt.NextSequenceValueResult{HasData: true}
	result.Data.FormattedValue = testPaddedValue
	return result, nil
}

type fakeSettings struct {
	values map[string]any
}

func (this *fakeSettings) GetOrgSettingValues(corectx.Context, model.Id, string) (map[string]any, error) {
	return this.values, nil
}

func newTestContext(t *testing.T) corectx.Context {
	return corectx.NewRequestContext(t.Context())
}

func TestFormatTransferNumber(t *testing.T) {
	assert.Equal(t, "MAIN/IN/00007", FormatTransferNumber("MAIN", "IN", testPaddedValue))
	// A type predating the prefix or the warehouse leaves the segment out rather than doubling "/".
	assert.Equal(t, "IN/00007", FormatTransferNumber("", "IN", testPaddedValue))
	assert.Equal(t, "MAIN/00007", FormatTransferNumber("MAIN", "", testPaddedValue))
}

func TestNumberingSegmentsFallBackToTheTypeIdWithoutAPrefix(t *testing.T) {
	warehouse, prefix := numberingSegments("MAIN", "IN", testOptypeId)
	assert.Equal(t, "MAIN", warehouse)
	assert.Equal(t, "IN", prefix)

	warehouse, prefix = numberingSegments("MAIN", "", testOptypeId)
	assert.Equal(t, testOptypeId+"/00007", FormatTransferNumber(warehouse, prefix, testPaddedValue))
}

func TestTransferSequenceCommandIsOneOrgDefinitionCountedPerType(t *testing.T) {
	cmd := TransferSequenceCommand(testOrgId)

	assert.Equal(t, "inventory.transfer", cmd.Code)
	assert.Equal(t, "org", cmd.ConfigurationScope)
	assert.Equal(t, testOrgId, string(*cmd.OrgId))
	assert.Equal(t, "inventory", cmd.OwnerModule)
	assert.Equal(t, "inventory_optype", *cmd.SubScope)
	assert.Equal(t, "consumer", *cmd.FormatOwner)
	assert.Nil(t, cmd.PrefixTemplate)
	assert.Nil(t, cmd.SuffixTemplate)
	assert.Equal(t, int32(5), *cmd.Padding)
	assert.Equal(t, int64(1), *cmd.Increment)
	assert.Equal(t, int64(1), *cmd.InitialValue)
	assert.Equal(t, "never", *cmd.ResetPeriod)
}

func TestScrapSequenceCommandIsSequenceFormattedWithPrefix(t *testing.T) {
	cmd := ScrapSequenceCommand(testOrgId)

	assert.Equal(t, ScrapSequenceCode, cmd.Code)
	assert.Nil(t, cmd.SubScope)
	assert.Equal(t, "sequence", *cmd.FormatOwner)
	assert.Equal(t, "SCR", *cmd.PrefixTemplate)
	assert.Equal(t, int32(5), *cmd.Padding)
}

func TestOrgSequenceRegistrationsRecordBothIds(t *testing.T) {
	registrations := OrgSequenceRegistrations(testOrgId)

	require.Len(t, registrations, 2)
	assert.Equal(t, invsettings.OrgSettingTransferSequenceDefinitionId, registrations[0].Setting)
	assert.Equal(t, TransferSequenceCode, registrations[0].Command.Code)
	assert.Equal(t, invsettings.OrgSettingScrapSequenceDefinitionId, registrations[1].Setting)
	assert.Equal(t, ScrapSequenceCode, registrations[1].Command.Code)
}

func TestScrapNextUsesTheStoredDefinitionAndADeterministicKey(t *testing.T) {
	sequences := &fakeSequences{}
	settings := &fakeSettings{values: map[string]any{
		invsettings.OrgSettingScrapSequenceDefinitionId: testScrapDefId,
	}}
	ctx := newTestContext(t)
	numbering := NewScrapSequence(sequences, settings)

	number, cErrs, err := numbering.Next(ctx, testOrgId)
	require.NoError(t, err)
	assert.Nil(t, cErrs)
	assert.Equal(t, testPaddedValue, number)
	_, _, err = numbering.Next(ctx, testOrgId)
	require.NoError(t, err)

	assert.Empty(t, sequences.upserts, "numbering never registers a definition")
	require.Len(t, sequences.nexts, 2)
	assert.Equal(t, testScrapDefId, string(sequences.nexts[0].DefinitionId))
	assert.Nil(t, sequences.nexts[0].ScopeObjectId)
	assert.Nil(t, sequences.nexts[0].IdempotencyId, "the port takes the request's idempotency id")
	assert.Equal(t, "scrap_number", *sequences.nexts[0].OperationKey)
	assert.Equal(t, *sequences.nexts[0].OperationKey, *sequences.nexts[1].OperationKey,
		"a retry keys the same issuance")
}

func TestScrapNextRefusesAnOrgWithoutNumbering(t *testing.T) {
	sequences := &fakeSequences{}
	settings := &fakeSettings{values: map[string]any{}}
	_, cErrs, err := NewScrapSequence(sequences, settings).Next(newTestContext(t), testOrgId)

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Equal(t, "inventory:sequence_not_configured", (*cErrs)[0].Key)
	assert.Empty(t, sequences.upserts, "a missing definition is not registered on use")
	assert.Empty(t, sequences.nexts)
}

func TestCorrectionDiscriminatorIsTheCallersKey(t *testing.T) {
	assert.Equal(t, "reservation:1", correctionDiscriminator(CorrectionRequest{
		IdempotencyKey: "reservation:1", OriginReference: "count:Q1",
	}))
	assert.Equal(t, "count:Q1", correctionDiscriminator(CorrectionRequest{OriginReference: "count:Q1"}))
	movement := CorrectionRequest{ProductVariantId: "V1", SourceLocationId: "L1", DestinationLocationId: "L2"}
	assert.Equal(t, correctionDiscriminator(movement), correctionDiscriminator(movement))
	assert.NotEqual(t, correctionDiscriminator(movement),
		correctionDiscriminator(CorrectionRequest{ProductVariantId: "V2", SourceLocationId: "L1"}))
}
