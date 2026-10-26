package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// quantRowRepository adds a lookup by id to the etag-honouring stub.
type quantRowRepository struct {
	moveWriteRepository
}

func (this *quantRowRepository) FindByKeys(
	_ corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	for _, row := range this.rows {
		if row[models.StockQuantFieldId] == keys[models.StockQuantFieldId] {
			return &dyn.OpResult[dmodel.DynamicFields]{Data: cloneFields(row), HasData: true}, nil
		}
	}
	return &dyn.OpResult[dmodel.DynamicFields]{}, nil
}

// The adjustment movement rewrites the balance before the count is cleared, so the clearing write
// must carry the row's current etag; with the one read before the movement it matched nothing and
// the balance stayed on the worklist.
func TestApplyingACountClearsItOnTheRewrittenRow(t *testing.T) {
	stored := dmodel.DynamicFields{
		models.StockQuantFieldId:               "qt-1",
		models.StockQuantFieldCountQuantitySet: true,
		models.StockQuantFieldCountedQuantity:  dec(t, "10"),
		models.StockQuantFieldCountSnapshotQty: dec(t, "10"),
		basemodel.FieldEtag:                    "etag-after-movement",
	}
	repo := &quantRowRepository{moveWriteRepository{rows: []dmodel.DynamicFields{stored}}}
	original := repoFor
	t.Cleanup(func() { repoFor = original })
	repoFor = func(string) (composable.CrudRepository, error) { return repo, nil }

	staleRead := cloneFields(stored)
	staleRead[basemodel.FieldEtag] = "etag-before-movement"
	result, err := applyCountToQuant(callerContext(), *models.NewStockQuantFrom(staleRead),
		LockedQuant{Id: "qt-1", OnHand: dec(t, "10")})

	require.NoError(t, err)
	require.Zero(t, result.ClientErrors.Count())
	assert.Equal(t, false, repo.rows[0][models.StockQuantFieldCountQuantitySet])
	assert.NotNil(t, repo.rows[0][models.StockQuantFieldLastCountDate])
}
