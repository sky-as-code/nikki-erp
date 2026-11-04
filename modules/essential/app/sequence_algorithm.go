package app

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The stored algorithm enum (FF1_FPE) is not the API's (pseudorandom). The helpers below translate
// the two other places the stored spelling would leak: the served meta/schema, and search graphs.

// apiAlgorithmSchema rewrites the algorithm enum of a simplified schema into the API vocabulary, so a
// client building a select from meta/schema offers values the API accepts.
func apiAlgorithmSchema(simplized any) (map[string]any, error) {
	raw, err := json.Marshal(simplized)
	if err != nil {
		return nil, errors.Wrap(err, "apiAlgorithmSchema.Marshal")
	}
	schema := map[string]any{}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, errors.Wrap(err, "apiAlgorithmSchema.Unmarshal")
	}
	fields, _ := schema["fields"].(map[string]any)
	field, _ := fields[models.SeqDefFieldAlgorithm].(map[string]any)
	dataType, _ := field["data_type"].(map[string]any)
	options, _ := dataType["options"].(map[string]any)
	enumKey := string(dmodel.FieldDataTypeOptEnumValues)
	if values, ok := options[enumKey].([]any); ok {
		options[enumKey] = mapAlgorithmValues(values, ApiAlgorithmOf)
	}
	return schema, nil
}

// storedAlgorithmGraph rewrites every condition on the algorithm field of a search graph into the
// stored vocabulary, so `algorithm = pseudorandom` finds the FF1_FPE rows. The graph is normalised to
// its decoded-map form, which every later stage of the search accepts.
func storedAlgorithmGraph(query dmodel.DynamicFields) error {
	graph, ok := query[basemodel.FieldGraph]
	if !ok || graph == nil {
		return nil
	}
	raw, err := graphJson(graph)
	if err != nil {
		return err
	}
	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return errors.Wrap(err, "storedAlgorithmGraph.Unmarshal")
	}
	mapAlgorithmNode(decoded)
	query[basemodel.FieldGraph] = decoded
	return nil
}

func graphJson(graph any) ([]byte, error) {
	if text, ok := graph.(string); ok {
		return []byte(text), nil
	}
	raw, err := json.Marshal(graph)
	return raw, errors.Wrap(err, "storedAlgorithmGraph.Marshal")
}

func mapAlgorithmNode(node map[string]any) {
	if condition, ok := node["if"].([]any); ok && len(condition) > 2 && condition[0] == models.SeqDefFieldAlgorithm {
		node["if"] = append(condition[:2:2], mapAlgorithmValues(condition[2:], StoredAlgorithmOf)...)
	}
	for _, key := range []string{"and", "or"} {
		children, _ := node[key].([]any)
		for _, child := range children {
			if childNode, ok := child.(map[string]any); ok {
				mapAlgorithmNode(childNode)
			}
		}
	}
}

// mapAlgorithmValues maps every string, including those inside a nested list (the `in` operator).
func mapAlgorithmValues(values []any, mapFn func(string) string) []any {
	result := make([]any, len(values))
	for i, item := range values {
		switch typed := item.(type) {
		case string:
			result[i] = mapFn(typed)
		case []any:
			result[i] = mapAlgorithmValues(typed, mapFn)
		default:
			result[i] = item
		}
	}
	return result
}
