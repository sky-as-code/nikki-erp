package computed

import (
	"bytes"
	"encoding/json"
	"strconv"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The JSON side of the DSL. ParseDefinitionJson produces exactly the same Expr tree the chained
// Go constructors build, so the two authoring forms are interchangeable. Structural validity is
// largely guaranteed by the model JSON Schema ($defs/computed) before this code runs; the checks
// here defend the direct-call path and decode semantics the schema cannot express.

func init() {
	// model_builder_json.go decodes a field's "computed" block through this seam; the import
	// direction (computed -> model) makes a direct call impossible. See model/computed_hooks.go.
	dmodel.RegisterComputedJsonParser(func(raw []byte, fieldName string) (any, bool, error) {
		expr, isStored, err := ParseDefinitionJson(raw, fieldName)
		return expr, isStored, err
	})
}

type definitionJsonDto struct {
	Kind       string       `json:"kind"`
	IsStored   *bool        `json:"is_stored"`
	Field      string       `json:"field"`
	Expression *exprJsonDto `json:"expression"`
}

type exprJsonDto struct {
	Field    string          `json:"field"`
	Value    json.RawMessage `json:"value"`
	Op       string          `json:"op"`
	Function string          `json:"function"`
	Args     []exprJsonDto   `json:"args"`
	Case     *caseJsonDto    `json:"case"`
}

type caseJsonDto struct {
	When []whenJsonDto `json:"when"`
	Else *exprJsonDto  `json:"else"`
}

type whenJsonDto struct {
	If   *exprJsonDto `json:"if"`
	Then *exprJsonDto `json:"then"`
}

// ParseDefinitionJson decodes a field's "computed" JSON block into the root expression and the
// declared is_stored flag — the exact pair FieldBuilder.Computed accepts.
func ParseDefinitionJson(raw []byte, fieldName string) (Expr, bool, error) {
	var dto definitionJsonDto
	if err := json.Unmarshal(raw, &dto); err != nil {
		return nil, false, errors.Wrapf(err, "computed definition of field %q", fieldName)
	}
	if dto.IsStored == nil {
		return nil, false, errors.Errorf("computed field %q must declare is_stored", fieldName)
	}
	root, err := parseDefinitionRoot(&dto, fieldName)
	if err != nil {
		return nil, false, err
	}
	return root, *dto.IsStored, nil
}

func parseDefinitionRoot(dto *definitionJsonDto, fieldName string) (Expr, error) {
	switch ComputeKind(dto.Kind) {
	case ComputeRelated:
		if dto.Field == "" {
			return nil, errors.Errorf("computed field %q: kind \"related\" requires a source field path", fieldName)
		}
		return Related(dto.Field), nil
	case ComputeExpression:
		if dto.Expression == nil {
			return nil, errors.Errorf("computed field %q: kind \"expression\" requires an expression", fieldName)
		}
		expr, err := parseExprDto(dto.Expression)
		return expr, errors.Wrapf(err, "computed field %q", fieldName)
	}
	return nil, errors.Errorf("computed field %q has unsupported kind %q", fieldName, dto.Kind)
}

func parseExprDto(dto *exprJsonDto) (Expr, error) {
	switch {
	case dto.Field != "":
		return F(dto.Field), nil
	case len(dto.Value) > 0:
		return parseLiteral(dto.Value)
	case dto.Op != "":
		return parseOperator(dto)
	case dto.Function != "":
		return parseFunctionCall(dto)
	case dto.Case != nil:
		return parseCase(dto.Case)
	}
	return nil, errors.New("expression node must set one of: field, value, op, function, case")
}

// parseLiteral decodes a literal into the native Go type the evaluator works with. Numbers are
// decoded via json.Number and normalized to int64 or decimal so money literals never pass
// through float64.
func parseLiteral(raw json.RawMessage) (Expr, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, errors.Wrap(err, "literal value")
	}
	if number, ok := value.(json.Number); ok {
		return Lit(normalizeJsonNumber(number)), nil
	}
	switch value.(type) {
	case nil, string, bool:
		return Lit(value), nil
	}
	return nil, errors.Errorf("literal %s is not a supported scalar (string, number, boolean or null)", raw)
}

func normalizeJsonNumber(number json.Number) any {
	if whole, err := strconv.ParseInt(number.String(), 10, 64); err == nil {
		return whole
	}
	parsed, err := coerceDecimal(number)
	if err != nil {
		// Unreachable for JSON-syntax numbers; keep the raw form rather than panic.
		return number.String()
	}
	return parsed
}

func parseOperator(dto *exprJsonDto) (Expr, error) {
	args, err := parseArgs(dto.Args)
	if err != nil {
		return nil, errors.Wrapf(err, "operator %q", dto.Op)
	}
	if unary := UnaryOperator(dto.Op); unary.IsValid() {
		if len(args) != 1 {
			return nil, errors.Errorf("operator %q expects exactly 1 argument but received %d", dto.Op, len(args))
		}
		return UnaryExpr{Op: unary, Operand: args[0]}, nil
	}
	binary := BinaryOperator(dto.Op)
	if !binary.IsValid() {
		return nil, errors.Errorf("operator %q is not supported", dto.Op)
	}
	return foldBinaryArgs(binary, args)
}

// foldBinaryArgs turns the args list into a left-folded chain. Associative operators accept two
// or more arguments; the rest require exactly two, so a subtle "subtract with three args" never
// silently reorders arithmetic.
func foldBinaryArgs(op BinaryOperator, args []Expr) (Expr, error) {
	associative := op == OpAdd || op == OpMultiply || op == OpAnd || op == OpOr
	if len(args) < 2 || (!associative && len(args) != 2) {
		return nil, errors.Errorf("operator %q expects exactly 2 arguments but received %d", op, len(args))
	}
	return foldBinary(op, args[0], args[1], args[2:]), nil
}

func parseFunctionCall(dto *exprJsonDto) (Expr, error) {
	args, err := parseArgs(dto.Args)
	if err != nil {
		return nil, errors.Wrapf(err, "function %q", dto.Function)
	}
	return FunctionExpr{Name: dto.Function, Args: args}, nil
}

func parseArgs(dtos []exprJsonDto) ([]Expr, error) {
	args := make([]Expr, len(dtos))
	for i := range dtos {
		arg, err := parseExprDto(&dtos[i])
		if err != nil {
			return nil, err
		}
		args[i] = arg
	}
	return args, nil
}

func parseCase(dto *caseJsonDto) (Expr, error) {
	if len(dto.When) == 0 || dto.Else == nil {
		return nil, errors.New("case requires at least one when branch and an else")
	}
	builder := Case()
	for i := range dto.When {
		branch := &dto.When[i]
		if branch.If == nil || branch.Then == nil {
			return nil, errors.New("case branch requires both if and then")
		}
		cond, err := parseExprDto(branch.If)
		if err != nil {
			return nil, err
		}
		then, err := parseExprDto(branch.Then)
		if err != nil {
			return nil, err
		}
		builder.When(cond, then)
	}
	fallback, err := parseExprDto(dto.Else)
	if err != nil {
		return nil, err
	}
	return builder.Else(fallback), nil
}
