package model

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"
)

func (this *SearchGraph) IsEmpty() bool {
	return this == nil ||
		(this.GetCondition() == nil && len(this.GetAnd()) == 0 && len(this.GetOr()) == 0)
}

// MergeAndNode ANDs the leading nodes above a caller's graph, so the caller's own shape - a
// top-level OR included - can only narrow the result, never widen it past them.
func MergeAndNode(requested *SearchGraph, leading ...SearchNode) *SearchGraph {
	if requested == nil {
		requested = NewSearchGraph()
	}

	merged := NewSearchGraph()
	if requested.IsEmpty() {
		merged.And(leading...)
	} else {
		merged.And(append(leading, *requested.ToSearchNode())...)
	}
	// A SearchNode carries no order, so rebuilding through one would drop the caller's sort.
	// An order-only graph is "empty" as a predicate but still names the sort the caller wants.
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

// AndDecodedGraph narrows an already-decoded graph value (the map a request binder produced, or a
// *SearchGraph) by ANDing the trailing nodes after it. The value is re-marshalled through
// SearchGraph, which owns the graph's JSON shape, so the caller's order survives the rebuild.
func AndDecodedGraph(raw any, trailing ...SearchNode) (*SearchGraph, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, errors.Wrap(err, "AndDecodedGraph.Marshal")
	}
	requested := NewSearchGraph()
	if err := json.Unmarshal(encoded, requested); err != nil {
		return nil, errors.Wrap(err, "AndDecodedGraph.Unmarshal")
	}
	merged := NewSearchGraph()
	if requested.IsEmpty() {
		merged.And(trailing...)
	} else {
		merged.And(append([]SearchNode{*requested.ToSearchNode()}, trailing...)...)
	}
	if order := requested.GetOrder(); len(order) > 0 {
		merged.Order(order)
	}
	return merged, nil
}
