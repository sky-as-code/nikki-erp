package computed

import (
	"bytes"
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// JSON decoding for the SQL-compiled kinds. Each parser builds the same node the chained
// constructors produce and runs the node's structural validation immediately, so a malformed
// JSON block fails at parse time with the field's name attached — not later at finalize.

type orderByJsonDto struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func parseAggregateJson(dto *definitionJsonDto, fieldName string) (Expr, error) {
	node := AggregateExpr{
		Source:   dto.Source,
		Function: AggregateFunction(dto.Function),
		Field:    dto.Field,
		Context:  dto.Context,
	}
	if dto.Expression != nil {
		inner, err := parseExprDto(dto.Expression)
		if err != nil {
			return nil, errors.Wrapf(err, "computed field %q", fieldName)
		}
		node.Expr = inner
	}
	if err := parseFilterAndDefault(dto, &node.Filter, &node.Default); err != nil {
		return nil, errors.Wrapf(err, "computed field %q", fieldName)
	}
	return node, errors.Wrapf(node.validate(), "computed field %q", fieldName)
}

func parseExistsJson(dto *definitionJsonDto, fieldName string) (Expr, error) {
	node := ExistsExpr{Source: dto.Source, Context: dto.Context}
	filter, err := parseFilterJson(dto.Filter)
	if err != nil {
		return nil, errors.Wrapf(err, "computed field %q", fieldName)
	}
	node.Filter = filter
	return node, errors.Wrapf(node.validate(), "computed field %q", fieldName)
}

func parseLookupJson(dto *definitionJsonDto, fieldName string) (Expr, error) {
	node := LookupExpr{Source: dto.Source, Field: dto.Field, Context: dto.Context}
	orderBy, err := parseOrderByDtos(dto.OrderBy)
	if err != nil {
		return nil, errors.Wrapf(err, "computed field %q", fieldName)
	}
	node.OrderBy = orderBy
	if err := parseFilterAndDefault(dto, &node.Filter, &node.Default); err != nil {
		return nil, errors.Wrapf(err, "computed field %q", fieldName)
	}
	return node, errors.Wrapf(node.validate(), "computed field %q", fieldName)
}

func parseFilterAndDefault(dto *definitionJsonDto, filter **dmodel.SearchNode, dflt *any) error {
	parsed, err := parseFilterJson(dto.Filter)
	if err != nil {
		return err
	}
	*filter = parsed
	value, err := parseDefaultJson(dto.Default)
	if err != nil {
		return err
	}
	*dflt = value
	return nil
}

// parseFilterJson decodes the filter block through SearchNode's own UnmarshalJSON, so the wire
// shape is byte-for-byte the search shape the rest of the platform uses ({"if": [...]},
// {"and": [...]}, {"or": [...]}) — never a computed-only filter language.
func parseFilterJson(raw json.RawMessage) (*dmodel.SearchNode, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	node := &dmodel.SearchNode{}
	if err := json.Unmarshal(raw, node); err != nil {
		return nil, errors.Wrap(err, "computed filter")
	}
	return node, nil
}

// parseDefaultJson decodes the scalar default the same way expression literals decode: numbers
// via json.Number normalized to int64 or decimal, never float64.
func parseDefaultJson(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, errors.Wrap(err, "default value")
	}
	if number, ok := value.(json.Number); ok {
		return normalizeJsonNumber(number), nil
	}
	switch value.(type) {
	case nil, string, bool:
		return value, nil
	}
	return nil, errors.Errorf("default %s is not a supported scalar (string, number, boolean or null)", raw)
}

func parseOrderByDtos(dtos []orderByJsonDto) ([]OrderBy, error) {
	if len(dtos) == 0 {
		return nil, nil
	}
	out := make([]OrderBy, len(dtos))
	for i, dto := range dtos {
		switch dto.Direction {
		case "", "asc":
			out[i] = OrderBy{Field: dto.Field}
		case "desc":
			out[i] = OrderBy{Field: dto.Field, Desc: true}
		default:
			return nil, errors.Errorf("order_by direction %q must be \"asc\" or \"desc\"", dto.Direction)
		}
	}
	return out, nil
}
