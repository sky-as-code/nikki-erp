package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func TestAlgorithmMapping(t *testing.T) {
	assert.Equal(t, models.SeqAlgorithmFf1Fpe, StoredAlgorithmOf(itSeq.AlgorithmPseudorandom))
	assert.Equal(t, models.SeqAlgorithmIncrement, StoredAlgorithmOf(itSeq.AlgorithmIncrement))
	assert.NotEqual(t, models.SeqAlgorithmFf1Fpe, StoredAlgorithmOf(models.SeqAlgorithmFf1Fpe),
		"the stored spelling is not part of the API")
	assert.Equal(t, itSeq.AlgorithmPseudorandom, ApiAlgorithmOf(models.SeqAlgorithmFf1Fpe))
	assert.Equal(t, itSeq.AlgorithmIncrement, ApiAlgorithmOf(models.SeqAlgorithmIncrement))
}

func TestSequenceOutput_StripsSecretsEverywhere(t *testing.T) {
	record := dmodel.DynamicFields{
		models.SeqDefFieldAlgorithm:       models.SeqAlgorithmFf1Fpe,
		models.SeqDefFieldAlgorithmParams: map[string]any{"key": "secret"},
		"definition":                      map[string]any{"algorithm_params": map[string]any{"key": "secret"}},
		"counters":                        []any{map[string]any{"algorithm_params": map[string]any{"tweak": "t"}}},
	}
	definitionOutput.record(record)
	assert.NotContains(t, record, models.SeqDefFieldAlgorithmParams)
	assert.NotContains(t, record["definition"], "algorithm_params")
	assert.NotContains(t, record["counters"].([]any)[0], "algorithm_params")
	assert.Equal(t, itSeq.AlgorithmPseudorandom, record[models.SeqDefFieldAlgorithm])
}

func TestSequenceOutput_StringifiesLargeIntegers(t *testing.T) {
	record := dmodel.DynamicFields{models.SeqCtrFieldLastNumber: int64(9007199254740993)}
	counterOutput.record(record)
	assert.Equal(t, "9007199254740993", record[models.SeqCtrFieldLastNumber])

	issuance := dmodel.DynamicFields{models.SeqIssFieldAllocatedValue: int64(26)}
	issuanceOutput.record(issuance)
	assert.Equal(t, "26", issuance[models.SeqIssFieldAllocatedValue])
}

func TestGuardQuery(t *testing.T) {
	query := dmodel.DynamicFields{"size": float64(1000)}
	assert.Nil(t, guardQuery(query))
	assert.Equal(t, seqMaxPageSize, query["size"])

	cErrs := guardQuery(dmodel.DynamicFields{"graph": map[string]any{"if": []any{"algorithm_params", "*", "x"}}})
	require.NotNil(t, cErrs)
	assert.Equal(t, "graph", (*cErrs)[0].Field)
}

func TestRefuseSecretWrite(t *testing.T) {
	assert.Nil(t, refuseSecretWrite(dmodel.DynamicFields{"name": "x"}, definitionServerOwned...))
	cErrs := refuseSecretWrite(dmodel.DynamicFields{
		models.SeqDefFieldAlgorithmParams: map[string]any{}, models.SeqDefFieldFirstIssuedAt: "2026-01-01T00:00:00Z",
	}, definitionServerOwned...)
	require.NotNil(t, cErrs)
	assert.Equal(t, 2, cErrs.Count())
}

func TestPreviewDefinition_Defaults(t *testing.T) {
	def, _, cErrs := previewDefinition(dmodel.DynamicFields{
		models.SeqDefFieldAlgorithm: itSeq.AlgorithmPseudorandom, models.SeqDefFieldPrefixTemplate: "CORE",
	})
	require.Nil(t, cErrs)
	assert.Equal(t, models.SeqAlgorithmFf1Fpe, def.GetAlgorithm())
	assert.Equal(t, int64(0), def.GetPadding(), "pseudorandom previews without padding")

	_, _, cErrs = previewDefinition(dmodel.DynamicFields{models.SeqDefFieldPrefixTemplate: "SO/{WH}/"})
	require.NotNil(t, cErrs)
}
