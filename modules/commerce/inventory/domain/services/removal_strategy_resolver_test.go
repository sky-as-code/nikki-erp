package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

type fakeLocation struct {
	strategy string
	parentId string
}

type fakeStrategyReader struct {
	categories map[string]string
	locations  map[string]fakeLocation
	reads      int
}

func (this *fakeStrategyReader) CategoryStrategy(_ corectx.Context, categoryId string) (string, error) {
	return this.categories[categoryId], nil
}

func (this *fakeStrategyReader) LocationStrategy(_ corectx.Context, locationId string) (string, string, error) {
	this.reads++
	location := this.locations[locationId]
	return location.strategy, location.parentId, nil
}

func newFakeStrategyReader() *fakeStrategyReader {
	return &fakeStrategyReader{
		categories: map[string]string{"forced": "lifo", "plain": ""},
		locations: map[string]fakeLocation{
			"root":  {strategy: "fefo"},
			"mid":   {parentId: "root"},
			"leaf":  {parentId: "mid"},
			"own":   {strategy: "lifo", parentId: "root"},
			"bare":  {},
			"loopA": {parentId: "loopB"},
			"loopB": {parentId: "loopA"},
		},
	}
}

func TestResolveRemovalStrategyOrder(t *testing.T) {
	cases := []struct {
		name, category, location, want string
	}{
		{"category forces over location", "forced", "root", "lifo"},
		{"location's own strategy", "plain", "own", "lifo"},
		{"nearest ancestor", "plain", "leaf", "fefo"},
		{"no category id", "", "leaf", "fefo"},
		{"nothing configured is fifo", "plain", "bare", models.InventoryLocationRemovalStrategyFifo},
		{"no location is fifo", "", "", models.InventoryLocationRemovalStrategyFifo},
	}
	for _, testCase := range cases {
		got, err := resolveRemovalStrategy(nil, newFakeStrategyReader(), testCase.category, testCase.location)
		require.NoError(t, err, testCase.name)
		assert.Equal(t, testCase.want, got, testCase.name)
	}
}

func TestResolveRemovalStrategyStopsOnACycle(t *testing.T) {
	reader := newFakeStrategyReader()
	got, err := resolveRemovalStrategy(nil, reader, "", "loopA")
	require.NoError(t, err)
	assert.Equal(t, models.InventoryLocationRemovalStrategyFifo, got)
	assert.Equal(t, 2, reader.reads)
}

func datedQuant(id string, incoming *time.Time) LockedQuant {
	quant := lockedQuant(id, "10", "0")
	quant.IncomingDate = incoming
	return quant
}

func quantIds(quants []LockedQuant) []string {
	ids := make([]string, 0, len(quants))
	for _, quant := range quants {
		ids = append(ids, string(quant.Id))
	}
	return ids
}

func TestOrderQuantsForRemoval(t *testing.T) {
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(24 * time.Hour)
	// The lock returns oldest first, nulls last.
	locked := []LockedQuant{datedQuant("old", &older), datedQuant("new", &newer), datedQuant("undated", nil)}

	fifoOrder := []string{"old", "new", "undated"}
	for _, strategy := range []string{"fifo", "fefo", "closest", "least_packages", ""} {
		assert.Equal(t, fifoOrder, quantIds(OrderQuantsForRemoval(locked, strategy)), strategy)
	}
	assert.Equal(t, []string{"new", "old", "undated"}, quantIds(OrderQuantsForRemoval(locked, "lifo")))
	assert.Equal(t, fifoOrder, quantIds(locked), "the caller's slice keeps the lock order")
}
