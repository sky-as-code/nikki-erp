package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itAction "github.com/sky-as-code/nikki-erp/modules/auditlog/interfaces/action"
)

func TestToHistoryItemMasksUnreadableFields(t *testing.T) {
	row := dmodel.DynamicFields{
		"id": "a1", "action": "update", "actor_type": "user", "actor_id": "u1",
		"actor_display_name": "Old name", "request_id": "req_1",
		"before":     map[string]any{"name": "old", "salary": 1},
		"after":      map[string]any{"name": "new", "salary": 2},
		"actor_user": map[string]any{"display_name": "Current name", "avatar_file_id": "f1"},
	}

	item := toHistoryItem(row, map[string]bool{"salary": true})

	assert.Equal(t, []itAction.HistoryChange{
		{Field: "name", Old: "old", New: "new"},
		{Field: "salary"},
	}, item.Changes)
	assert.Equal(t, []string{"salary"}, item.MaskedFields)
	assert.Equal(t, "Current name", *item.Actor.DisplayName)
	assert.Equal(t, "f1", *item.Actor.AvatarFileId)
	assert.Equal(t, "req_1", *item.RequestId)
}

func TestToHistoryActorFallsBackToTheSnapshot(t *testing.T) {
	actor := toHistoryActor(dmodel.DynamicFields{"actor_type": "service", "actor_display_name": "itstorage"})

	assert.Equal(t, "service", actor.Type)
	assert.Equal(t, "itstorage", *actor.DisplayName)
	assert.Nil(t, actor.Id)
	assert.Nil(t, actor.AvatarFileId)
}

func TestParseHistoryRequestIsZeroBasedAndClamped(t *testing.T) {
	request, cErrs := parseHistoryRequest(dmodel.DynamicFields{
		"resource_code": "purchase_order", "resource_id": "p1", "size": "1000",
	})

	require.Nil(t, cErrs)
	assert.Equal(t, 0, request.page)
	assert.Equal(t, 100, request.size)

	_, cErrs = parseHistoryRequest(dmodel.DynamicFields{"page": "-1"})
	require.NotNil(t, cErrs)
	assert.Equal(t, 3, cErrs.Count())
}

func TestBuildActionRowFallsBackToSystem(t *testing.T) {
	row, err := buildActionRow(itAuditEntry(), actorSnapshot{kind: "system"}, metaWithoutSource())

	require.NoError(t, err)
	assert.Equal(t, "system", row["source"])
	assert.Equal(t, map[string]any{}, row["before"])
	assert.Equal(t, map[string]any{"name": "n"}, row["after"])
	_, hasActor := row["actor_id"]
	assert.False(t, hasActor)
}
