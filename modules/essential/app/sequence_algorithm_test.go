package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

func TestApiAlgorithmSchema_ServesApiVocabulary(t *testing.T) {
	enumType := dmodel.FieldDataTypeEnumString([]string{models.SeqAlgorithmIncrement, models.SeqAlgorithmFf1Fpe})
	simplized := map[string]any{"fields": map[string]any{
		models.SeqDefFieldAlgorithm: map[string]any{"data_type": enumType.ToSimplized()},
		"code":                      map[string]any{"name": "code"},
	}}

	schema, err := apiAlgorithmSchema(simplized)
	require.NoError(t, err)

	field := schema["fields"].(map[string]any)[models.SeqDefFieldAlgorithm].(map[string]any)
	options := field["data_type"].(map[string]any)["options"].(map[string]any)
	assert.Equal(t, []any{"increment", "pseudorandom"}, options[string(dmodel.FieldDataTypeOptEnumValues)])
	assert.Contains(t, schema["fields"], "code", "other fields pass through")
}

func TestStoredAlgorithmGraph_MapsNestedConditions(t *testing.T) {
	graph := dmodel.NewSearchGraph()
	node := dmodel.NewSearchNode().NewCondition(models.SeqDefFieldAlgorithm, dmodel.Equals, "pseudorandom")
	other := dmodel.NewSearchNode().NewCondition("code", dmodel.Equals, "pseudorandom")
	graph.And(*node, *other)
	query := dmodel.DynamicFields{basemodel.FieldGraph: graph}

	require.NoError(t, storedAlgorithmGraph(query))

	and := query[basemodel.FieldGraph].(map[string]any)["and"].([]any)
	assert.Equal(t, models.SeqAlgorithmFf1Fpe, and[0].(map[string]any)["if"].([]any)[2])
	assert.Equal(t, "pseudorandom", and[1].(map[string]any)["if"].([]any)[2], "only the algorithm field is mapped")
}

func TestStoredAlgorithmGraph_MapsInListAndText(t *testing.T) {
	query := dmodel.DynamicFields{basemodel.FieldGraph: `{"if":["algorithm","in",["pseudorandom","increment"]]}`}

	require.NoError(t, storedAlgorithmGraph(query))

	condition := query[basemodel.FieldGraph].(map[string]any)["if"].([]any)
	assert.Equal(t, []any{models.SeqAlgorithmFf1Fpe, models.SeqAlgorithmIncrement}, condition[2])
}

func TestStoredAlgorithmGraph_NoGraph(t *testing.T) {
	query := dmodel.DynamicFields{}
	require.NoError(t, storedAlgorithmGraph(query))
	assert.NotContains(t, query, basemodel.FieldGraph)
}
