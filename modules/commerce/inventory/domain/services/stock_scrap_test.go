package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func TestScrapQuantityMustBePositive(t *testing.T) {
	// The column allows zero, but a scrap of nothing asserts a movement that never happened, so the
	// rule lives here.
	for _, quantity := range []string{"0", "-1"} {
		vErrs := assertScrapQuantityPositive(dmodel.DynamicFields{
			models.StockScrapFieldQuantity: quantity,
		})

		assert.Equal(t, 1, vErrs.Count(), "quantity %s should be refused", quantity)
	}
}

func TestScrapQuantityAcceptsAPositive(t *testing.T) {
	vErrs := assertScrapQuantityPositive(dmodel.DynamicFields{
		models.StockScrapFieldQuantity: "2.5",
	})

	assert.Equal(t, 0, vErrs.Count())
}

func TestScrapQuantityIsOnlyCheckedWhenSupplied(t *testing.T) {
	// An update that changes only the note carries no quantity, and must not be refused for a
	// field it never mentioned.
	assert.Equal(t, 0, assertScrapQuantityPositive(dmodel.DynamicFields{}).Count())
}

func TestScrapDimensionMatchesEveryRowWhenUntracked(t *testing.T) {
	// An empty lot on the document means "not tracked", so the whole location's balance is fair
	// game. Treating '' as a lot to match would find nothing and make every untracked scrap fail.
	scrap := models.NewStockScrapFrom(dmodel.DynamicFields{})

	assert.True(t, matchesScrapDimension(*scrap, LockedQuant{LotRef: "LOT-A"}))
	assert.True(t, matchesScrapDimension(*scrap, LockedQuant{LotRef: ""}))
}

func TestScrapDimensionNarrowsToTheNamedLot(t *testing.T) {
	// Scrapping lot A must not draw down lot B's stock: they are different goods that happen to
	// share a location.
	scrap := models.NewStockScrapFrom(dmodel.DynamicFields{
		models.StockScrapFieldLotRef: "LOT-A",
	})

	assert.True(t, matchesScrapDimension(*scrap, LockedQuant{LotRef: "LOT-A"}))
	assert.False(t, matchesScrapDimension(*scrap, LockedQuant{LotRef: "LOT-B"}))
}

func TestScrapDimensionNarrowsOnEveryNamedAxis(t *testing.T) {
	scrap := models.NewStockScrapFrom(dmodel.DynamicFields{
		models.StockScrapFieldLotRef:     "LOT-A",
		models.StockScrapFieldPackageRef: "PKG-1",
		models.StockScrapFieldOwnerRef:   "OWN-9",
	})

	assert.True(t, matchesScrapDimension(*scrap, LockedQuant{
		LotRef: "LOT-A", PackageRef: "PKG-1", OwnerRef: "OWN-9",
	}))
	// One axis differing is enough to exclude the row.
	assert.False(t, matchesScrapDimension(*scrap, LockedQuant{
		LotRef: "LOT-A", PackageRef: "PKG-2", OwnerRef: "OWN-9",
	}))
}

func TestGeneratedScrapNumbersAreDistinct(t *testing.T) {
	first, err := generateScrapNumber()
	assert.NoError(t, err)
	second, err := generateScrapNumber()
	assert.NoError(t, err)

	assert.NotEqual(t, first, second)
	assert.Contains(t, first, "SCR-")
}
