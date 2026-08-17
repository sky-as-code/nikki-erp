package baserepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

const virtualRepoSchemaName = "test_baserepo_virt"

// virtualRepo builds a repository over a schema carrying one physical column, one virtual scalar
// and one edge, which is the minimum needed to tell the three apart.
func virtualRepo(t *testing.T) *BaseDynamicRepositoryImpl {
	t.Helper()
	registry := dmodel.GetSchemaRegistry()
	if registry.Get(virtualRepoSchemaName) == nil {
		require.NoError(t, dmodel.RegisterSchemaB(
			dmodel.DefineModel(virtualRepoSchemaName+"_peer").
				TableName("test_baserepo_virt_peers").
				ShouldBuildDb().
				Field(dmodel.DefineField().Name("id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
				Field(dmodel.DefineField().Name("name").
					DataType(dmodel.FieldDataTypeString(0, 200)))))

		require.NoError(t, dmodel.RegisterSchemaB(
			dmodel.DefineModel(virtualRepoSchemaName).
				TableName("test_baserepo_virts").
				ShouldBuildDb().
				Field(dmodel.DefineField().Name("id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate().PrimaryKey()).
				Field(dmodel.DefineField().Name("peer_id").
					DataType(dmodel.FieldDataTypeUlid()).RequiredForCreate()).
				Field(dmodel.DefineField().Name("sku").
					DataType(dmodel.FieldDataTypeString(1, 100)).RequiredForCreate().Unique()).
				Field(dmodel.DefineField().Name("peer_name").
					DataType(dmodel.FieldDataTypeString(0, 200)).
					Computed(false, computed.Related("peer.name"))).
				EdgeTo(dmodel.Edge("peer").ManyToOne(
					virtualRepoSchemaName+"_peer",
					dmodel.DynamicFields{"peer_id": "id"}))))

		// Relations carry no foreign-key pairs until they are finalized, and the edge-fetch plan
		// reads those pairs to decide which columns the main SELECT needs.
		require.NoError(t, registry.FinalizeRelations())
	}

	return &BaseDynamicRepositoryImpl{schema: registry.Get(virtualRepoSchemaName)}
}

// The single most important guard in this change.
//
// hasNestedOrEdgeColumns decides whether a request takes the per-row hydration path, which costs
// one follow-up query per row. A virtual scalar is filled by its service from a single batched
// read, so it must NOT trigger that path. Widening the check to IsVirtual would silently turn
// every request selecting a virtual field into an N+1 -- invisible in behaviour, fatal in cost.
func TestVirtual_ScalarDoesNotTriggerNestedHydration(t *testing.T) {
	repo := virtualRepo(t)

	assert.False(t, repo.hasNestedOrEdgeColumns([]string{"id", "sku", "peer_name"}),
		"a virtual scalar must stay on the plain, single-query path")
	assert.True(t, repo.hasNestedOrEdgeColumns([]string{"id", "peer"}),
		"a real edge still needs hydration")
	assert.True(t, repo.hasNestedOrEdgeColumns([]string{"id", "peer.name"}),
		"a dotted edge path still needs hydration")
	assert.False(t, repo.hasNestedOrEdgeColumns([]string{"id", "sku"}),
		"plain columns need no hydration")
}

// A virtual field is a legal thing to select, so GetOne must not reject it as unknown.
func TestVirtual_AcceptedInGetOneFields(t *testing.T) {
	repo := virtualRepo(t)

	vErr := repo.validateGetOneColumnsAndFilter(
		[]string{"id", "sku", "peer_name"},
		dmodel.DynamicFields{"id": "01J000000000000000000000"},
	)

	assert.Nil(t, vErr)
}

// A GetOne filter must resolve to a primary or unique key, which a virtual field can never be.
// Reporting it as "unavailable" rather than "unknown" tells the caller the field exists but
// cannot identify a record.
func TestVirtual_RejectedInGetOneFilter(t *testing.T) {
	repo := virtualRepo(t)

	vErr := repo.validateGetOneColumnsAndFilter(
		[]string{"id"},
		dmodel.DynamicFields{"peer_name": "Classic T-Shirt"},
	)

	require.NotNil(t, vErr)
	assert.Contains(t, string(vErr.Key), "err_virtual_field_unavailable")
}

func TestVirtual_UnknownFieldStillRejectedInGetOneFields(t *testing.T) {
	repo := virtualRepo(t)

	vErr := repo.validateGetOneColumnsAndFilter(
		[]string{"id", "no_such_field"},
		dmodel.DynamicFields{"id": "01J000000000000000000000"},
	)

	require.NotNil(t, vErr)
	assert.Contains(t, string(vErr.Key), "err_unknown_schema_field")
}

// A bare edge expands to what a client may read from the destination, so a virtual scalar on the
// far side comes along too.
func TestVirtual_ReadableFieldNamesIncludesVirtual(t *testing.T) {
	repo := virtualRepo(t)

	assert.Equal(t,
		[]string{"id", "peer_id", "sku", "peer_name"},
		readableFieldNames(repo.schema))
}
