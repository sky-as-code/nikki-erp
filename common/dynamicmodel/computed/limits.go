package computed

// Limits are the guardrails that keep a schema author from declaring a computation the engine
// cannot evaluate safely. All shape limits are enforced once, at schema finalize time;
// MaxComputedFieldsPerRequest is checked per read request by the eval planner.
type Limits struct {
	// MaxExpressionNestingDepth bounds how deep an expression tree may nest.
	MaxExpressionNestingDepth int
	// MaxComputedDependencyDepth bounds a computed-field-depends-on-computed-field chain.
	MaxComputedDependencyDepth int
	// MaxRelatedPathDepth bounds the edge-chain length of a related path. The current phase
	// evaluates a single forward to-one hop, so values above 1 are reserved for later phases.
	MaxRelatedPathDepth int
	// MaxComputedFieldsPerRequest bounds how many computed fields one read may evaluate.
	MaxComputedFieldsPerRequest int
}

// DefaultLimits returns the standard guardrails. Override with SetLimits at application start.
func DefaultLimits() Limits {
	return Limits{
		MaxExpressionNestingDepth:   10,
		MaxComputedDependencyDepth:  5,
		MaxRelatedPathDepth:         1,
		MaxComputedFieldsPerRequest: 15,
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
