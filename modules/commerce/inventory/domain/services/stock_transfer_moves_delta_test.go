package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	itStock "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/stock"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func TestLiftMovesOrRefuseRemovesTheDeltaFromTheCommand(t *testing.T) {
	cmd := dmodel.DynamicFields{
		"note": "x",
		"moves": map[string]any{
			"add": []any{map[string]any{"product_variant_id": "V1", "demand_quantity": "2.5"}},
		},
	}

	delta, cErrs := liftMovesOrRefuse(cmd)

	require.Empty(t, cErrs)
	require.NotNil(t, delta)
	assert.NotContains(t, cmd, itStock.TransferFieldMoves)
	require.Len(t, delta.Add, 1)
	assert.Equal(t, "2.5", delta.Add[0].DemandQuantity.String())
}

func TestLiftMovesOrRefuseRefusesAMalformedDelta(t *testing.T) {
	_, cErrs := liftMovesOrRefuse(dmodel.DynamicFields{"moves": []any{"not", "a", "delta"}})

	require.Len(t, cErrs, 1)
	assert.Equal(t, itStock.TransferFieldMoves, cErrs[0].Field)
}

func TestLiftMovesOrRefuseKeysErrorsByRow(t *testing.T) {
	_, cErrs := liftMovesOrRefuse(dmodel.DynamicFields{"moves": map[string]any{
		"add":    []any{map[string]any{"demand_quantity": 0}},
		"update": []any{map[string]any{"id": "M1", "etag": "e"}},
		"remove": []any{"M1"},
	}})

	fields := make([]string, 0, len(cErrs))
	for _, item := range cErrs {
		fields = append(fields, item.Field)
	}
	assert.ElementsMatch(t, []string{
		"moves.add[0].product_variant_id",
		"moves.add[0].demand_quantity",
		"moves.remove[0]",
	}, fields)
}

func TestReadTransferMovesDeltaPartsCountsAMalformedDeltaAsEveryPart(t *testing.T) {
	parts := itStock.ReadTransferMovesDeltaParts(dmodel.DynamicFields{"moves": "garbage"})

	assert.Equal(t, itStock.TransferMovesDeltaParts{Add: true, Update: true, Remove: true, Resolve: true}, parts)
}

type operationTypeRowsRepository struct {
	composable.CrudRepository
	rows  []dmodel.DynamicFields
	graph *dmodel.SearchGraph
}

func (this *operationTypeRowsRepository) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	this.graph = param.Graph
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		HasData: len(this.rows) > 0,
		Data:    dyn.PagedResultData[dmodel.DynamicFields]{Items: this.rows, Total: len(this.rows)},
	}, nil
}

func useOperationTypeRows(t *testing.T, rows ...dmodel.DynamicFields) *operationTypeRowsRepository {
	t.Helper()
	repo := &operationTypeRowsRepository{rows: rows}
	original := repoFor
	t.Cleanup(func() { repoFor = original })
	repoFor = func(string) (composable.CrudRepository, error) { return repo, nil }
	return repo
}

func TestResolveOperationTypeByCodePicksTheOrgsSingleActiveType(t *testing.T) {
	repo := useOperationTypeRows(t, dmodel.DynamicFields{"id": "01TYPE"})

	typeId, cErrs, err := resolveOperationTypeByCode(callerContext(),
		dmodel.DynamicFields{"operation_code": "incoming", "org_id": "01ORG"})

	require.NoError(t, err)
	require.Empty(t, cErrs)
	assert.Equal(t, "01TYPE", typeId)
	assertSearchedUnarchivedOnly(t, &graphCapturingRepository{graphs: []*dmodel.SearchGraph{repo.graph}})
}

func TestResolveOperationTypeByCodeRefusesNoneOrSeveral(t *testing.T) {
	for _, rows := range [][]dmodel.DynamicFields{nil, {{"id": "01A"}, {"id": "01B"}}} {
		useOperationTypeRows(t, rows...)

		typeId, cErrs, err := resolveOperationTypeByCode(callerContext(),
			dmodel.DynamicFields{"operation_code": "outgoing", "org_id": "01ORG"})

		require.NoError(t, err)
		assert.Empty(t, typeId)
		require.Len(t, cErrs, 1)
		assert.Equal(t, "operation_type_id", cErrs[0].Field)
		assert.Equal(t, "inventory:stock_transfer.operation_type_unresolved", cErrs[0].Key)
	}
}
