package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeAndNode_KeepsTheOrderOfAnOrderOnlyGraph(t *testing.T) {
	requested := NewSearchGraph().OrderBy("name", Desc)

	merged := MergeAndCondition(requested, "org_id", Equals, "org_1")

	assert.Equal(t, NewSearchOrder("name", Desc), merged.GetOrder())
	assert.Len(t, merged.GetAnd(), 1)
}

func TestAndDecodedGraph_KeepsConditionAndOrder(t *testing.T) {
	orgNode := *NewSearchNode().NewCondition("org_id", Equals, "org_1")
	raw := map[string]any{"if": []any{"name", "=", "x"}, "order": []any{[]any{"name", "asc"}}}

	merged, err := AndDecodedGraph(raw, orgNode)

	require.NoError(t, err)
	assert.Len(t, merged.GetAnd(), 2, "the caller's condition and the org condition")
	assert.Equal(t, NewSearchOrder("name", Asc), merged.GetOrder())
}

func TestAndDecodedGraph_AcceptsAnAlreadyBuiltGraph(t *testing.T) {
	orgNode := *NewSearchNode().NewCondition("org_id", Equals, "org_1")
	built := NewSearchGraph().NewCondition("name", Equals, "x").OrderBy("name", Desc)

	merged, err := AndDecodedGraph(built, orgNode)

	require.NoError(t, err)
	assert.Len(t, merged.GetAnd(), 2)
	assert.Equal(t, NewSearchOrder("name", Desc), merged.GetOrder())
}
