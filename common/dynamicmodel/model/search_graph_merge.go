package model

func (this *SearchGraph) IsEmpty() bool {
	return this == nil ||
		(this.GetCondition() == nil && len(this.GetAnd()) == 0 && len(this.GetOr()) == 0)
}

// MergeAndNode ANDs the leading nodes above a caller's graph, so the caller's own shape - a
// top-level OR included - can only narrow the result, never widen it past them.
func MergeAndNode(requested *SearchGraph, leading ...SearchNode) *SearchGraph {
	merged := NewSearchGraph()
	if requested.IsEmpty() {
		merged.And(leading...)
		return merged
	}

	merged.And(append(leading, *requested.ToSearchNode())...)
	// A SearchNode carries no order, so rebuilding through one would drop the caller's sort.
	if order := requested.GetOrder(); len(order) > 0 {
		merged.Order(order)
	}

	return merged
}

// MergeAndCondition is MergeAndNode for the common case of narrowing by a single field condition.
func MergeAndCondition(
	requested *SearchGraph, field string, operator Operator, values ...any,
) *SearchGraph {
	return MergeAndNode(requested, *NewSearchNode().NewCondition(field, operator, values...))
}
