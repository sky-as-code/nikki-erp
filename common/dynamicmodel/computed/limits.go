package computed

// Limits are the guardrails that keep a schema author from declaring a computation the engine
// cannot evaluate safely. All shape limits are enforced once, at schema finalize time;
// MaxComputedFieldsPerRequest is checked per read request by the eval planner.
type Limits struct {
	// MaxExpressionNestingDepth bounds how deep an expression tree may nest.
	MaxExpressionNestingDepth int
	// MaxComputedDependencyDepth bounds a computed-field-depends-on-computed-field chain.
	MaxComputedDependencyDepth int
	// MaxRelatedPathDepth bounds the edge-chain length of a related path after a derived leaf
	// has been flattened into it ("template.uom.name" is two). The first hop is a batched read;
	// the rest is a nested field path on that read, which the repository projects in the same
	// statement, so the cost of a longer chain is one more join, not one more query.
	MaxRelatedPathDepth int
	// MaxComputedFieldsPerRequest bounds how many computed fields one read may evaluate. Related
	// fields cost one batched read per edge however many of them a schema declares, so a schema
	// that flattens a dozen template_* fields onto a variant stays well inside this bound.
	MaxComputedFieldsPerRequest int
	// MaxFilterNestingDepth bounds how deep an SQL-kind filter (and/or tree) may nest.
	MaxFilterNestingDepth int
	// MaxSqlComputedFieldsPerRequest bounds how many correlated subqueries one read may project.
	MaxSqlComputedFieldsPerRequest int
}

// DefaultLimits returns the standard guardrails. Override with SetLimits at application start.
func DefaultLimits() Limits {
	return Limits{
		MaxExpressionNestingDepth:      10,
		MaxComputedDependencyDepth:     5,
		MaxRelatedPathDepth:            2,
		MaxComputedFieldsPerRequest:    30,
		MaxFilterNestingDepth:          5,
		MaxSqlComputedFieldsPerRequest: 10,
	}
}

var activeLimits = DefaultLimits()

// SetLimits overrides the guardrails process-wide. Call before schemas are finalized.
func SetLimits(limits Limits) {
	activeLimits = limits
}

// ActiveLimits returns the guardrails currently in force.
func ActiveLimits() Limits {
	return activeLimits
}
