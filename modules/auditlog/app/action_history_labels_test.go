package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
)

func historyWithEdges() []itAction.HistoryItem {
	return []itAction.HistoryItem{
		{Action: "link", Changes: []itAction.HistoryChange{{Field: "roles", New: []any{"r1", "r2"}}}},
		{Action: "unlink", Changes: []itAction.HistoryChange{{Field: "roles", Old: []any{"r2"}}}},
		{Action: "update", Changes: []itAction.HistoryChange{
			{Field: "category_id", Old: "c1", New: "c2"},
			{Field: "name", Old: "a", New: "b"},
		}},
	}
}

func TestCollectEdgeIdsGroupsOnePageByTargetSchema(t *testing.T) {
	targets := map[string]string{"roles": "iam_role", "category_id": "essential_uomcat"}

	ids := collectEdgeIds(targets, historyWithEdges())

	assert.Equal(t, map[string][]model.Id{
		"iam_role":         {"r1", "r2"},
		"essential_uomcat": {"c1", "c2"},
	}, ids)
}

func TestApplyEdgeLabelsLabelsOnlyEdgesAndResolvedIds(t *testing.T) {
	items := historyWithEdges()
	targets := map[string]string{"roles": "iam_role", "category_id": "essential_uomcat"}
	labels := map[string]map[string]string{
		"iam_role":         {"r1": "Role one"},
		"essential_uomcat": {"c1": "Weight", "c2": "Length"},
	}

	applyEdgeLabels(targets, labels, items)

	assert.Equal(t, map[string]string{"r1": "Role one"}, items[0].Changes[0].Labels,
		"an unreadable or deleted target is left out")
	assert.Nil(t, items[1].Changes[0].Labels, "no resolvable id, no labels")
	assert.Equal(t, map[string]string{"c1": "Weight", "c2": "Length"}, items[2].Changes[0].Labels)
	assert.Nil(t, items[2].Changes[1].Labels, "a plain field carries no labels")
	assert.Equal(t, []any{"r1", "r2"}, items[0].Changes[0].New, "values stay raw ids")
}

func TestEdgeTargetsOfWithoutSchemaIsEmpty(t *testing.T) {
	assert.Empty(t, edgeTargetsOf(nil, historyWithEdges()))
	assert.Equal(t, []string{"a", "b"}, idsOf([]any{"a", 3, "b"}))
	assert.Nil(t, idsOf(nil))
}
