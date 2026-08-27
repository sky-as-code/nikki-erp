package computed

import (
	"strings"

	"go.bryk.io/pkg/errors"
)

// Chained authoring for the function kind, the sixth and only non-declarative one. Where the other
// kinds describe a value, this one names a Go function registered on the owning engine:
//
//	Computed(false, computed.GoFunction("inventory.effective_sales_tax_ids").DependsOn("sales_tax_mode"))
//
// The builder mirrors CaseBuilder: it accumulates and returns an Expr, so both authoring forms
// funnel through the same validate() chokepoint in NewDefinition.

// GoFunctionBuilder accumulates the optional parts of a function definition.
type GoFunctionBuilder struct {
	expr GoFunctionExpr
}

// GoFunction starts a function definition. Name must match a DefineComputedFieldFunction
// registration on the engine owning this schema, or the application fails to boot.
func GoFunction(name string) *GoFunctionBuilder {
	return &GoFunctionBuilder{expr: GoFunctionExpr{Name: name}}
}

// DependsOn names a same-schema field the computation reads, enabling real-time recompute from a
// form. The field must exist; resolution enforces it.
func (b *GoFunctionBuilder) DependsOn(field string) *GoFunctionBuilder {
	b.expr.DependsOn = field
	return b
}

// Build finishes the expression. NewDefinition also accepts the builder itself, so an inline
// GoFunction(...).DependsOn(...) needs no trailing Build() call.
func (b *GoFunctionBuilder) Build() Expr {
	return b.expr
}

// validate applies the structural rules both authoring forms share, matching how the SQL kinds
// validate in ast_sql.go. Whether the named function is actually registered is a boot-time check
// against the engine, not a structural one — see AssertComputedFunctionsDefined.
func (this GoFunctionExpr) validate() error {
	if strings.TrimSpace(this.Name) == "" {
		return errors.New("function requires the name of a registered computed-field function")
	}
	return nil
}

// parseFunctionJson decodes the function kind's JSON block, mirroring the SQL parsers in
// schema_parse_sql.go: build the node the chained constructor produces, then run the node's own
// validation immediately so a malformed block fails at parse time with the field name attached.
func parseFunctionJson(dto *definitionJsonDto, fieldName string) (Expr, error) {
	node := GoFunctionExpr{Name: dto.Function, DependsOn: dto.DependsOn}
	return node, errors.Wrapf(node.validate(), "computed field %q", fieldName)
}
