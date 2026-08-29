package services

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// The product stock summary from the rollup side: available is derived not stored, a template total
// is the sum of its variants, and a place holding several lots of one product counts once.

const (
	testVariantAId  = "01VARIANTA0000000000000000"
	testVariantBId  = "01VARIANTB0000000000000000"
	testLocationAId = "01LOCATIONA000000000000000"
	testLocationBId = "01LOCATIONB000000000000000"
	testWarehouseId = "01WAREHOUSE000000000000000"
)

// stubRowRepository stands in for one schema's repository, returning the rows it was given. It must
// apply the graph's status conditions: the real repository filters in SQL, so a stub returning
// everything would let a test pass while the production query excluded the rows under test.
type stubRowRepository struct {
	drif.DynamicResourceRepository

	rows []dmodel.DynamicFields
}

func (this *stubRowRepository) Search(
	_ corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	rows := filterRowsByStatus(this.rows, param.Graph)
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data: dyn.PagedResultData[dmodel.DynamicFields]{
			Items: rows,
			Total: len(rows),
		},
		HasData: true,
	}, nil
}

// filterRowsByStatus applies whatever status restriction the graph carries. Only status is honoured;
// the variant and location conditions merely select rows the fixture already scopes by hand.
func filterRowsByStatus(
	rows []dmodel.DynamicFields, graph *dmodel.SearchGraph,
) []dmodel.DynamicFields {
	excluded, included := statusConditionsOf(graph)
	if len(excluded) == 0 && len(included) == 0 {
		return rows
	}

	kept := make([]dmodel.DynamicFields, 0, len(rows))
	for _, row := range rows {
		status, _ := row[models.StockMoveFieldStatus].(string)
		if excluded[status] {
			continue
		}
		if len(included) > 0 && !included[status] {
			continue
		}
		kept = append(kept, row)
	}
	return kept
}

// statusConditionsOf reads the status restrictions out of a graph. Its conditions are unexported
// but it marshals to JSON, so it is inspected through that.
func statusConditionsOf(graph *dmodel.SearchGraph) (map[string]bool, map[string]bool) {
	excluded, included := map[string]bool{}, map[string]bool{}
	if graph == nil {
		return excluded, included
	}

	encoded, err := json.Marshal(graph)
	if err != nil {
		return excluded, included
	}
	rendered := string(encoded)

	statusField := `"` + models.StockMoveFieldStatus + `"`
	if !strings.Contains(rendered, statusField) {
		return excluded, included
	}

	negated := strings.Contains(rendered, string(dmodel.NotIn))
	for _, status := range allMoveStatuses() {
		if !strings.Contains(rendered, `"`+status+`"`) {
			continue
		}
		if negated {
			excluded[status] = true
		} else {
			included[status] = true
		}
	}
	return excluded, included
}

func allMoveStatuses() []string {
	return []string{
		models.StockMoveStatusDraft,
		models.StockMoveStatusWaiting,
		models.StockMoveStatusConfirmed,
		models.StockMoveStatusPartiallyAvailable,
		models.StockMoveStatusAssigned,
		models.StockMoveStatusDone,
		models.StockMoveStatusCancelled,
	}
}

// useSchemaEngines routes engineFor to a repository per schema, restoring the original afterwards
// so one test's substitution cannot leak into the next.
func useSchemaEngines(t *testing.T, rowsBySchema map[string][]dmodel.DynamicFields) {
	t.Helper()

	original := engineFor
	t.Cleanup(func() { engineFor = original })
	engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
		return &stubEngine{repo: &stubRowRepository{rows: rowsBySchema[schemaName]}}, nil
	}
}

func quantRow(variantId, locationId string, onHand, reserved int64) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.StockQuantFieldProductVariantId: variantId,
		models.StockQuantFieldLocationId:       locationId,
		models.StockQuantFieldOnHandQuantity:   decimal.NewFromInt(onHand),
		models.StockQuantFieldReservedQuantity: decimal.NewFromInt(reserved),
	}
}

func internalLocationRow(id string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.InventoryLocationFieldId:            id,
		models.InventoryLocationFieldLocationUsage: models.InventoryLocationUsageInternal,
		models.InventoryLocationFieldStatus:        models.InventoryLocationStatusActive,
		models.InventoryLocationFieldWarehouseId:   testWarehouseId,
	}
}

func summariseVariants(t *testing.T, variantIds ...string) map[string]itStock.VariantStockSummary {
	t.Helper()

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetVariantSummaries(
		callerContext(), itStock.GetVariantSummariesQuery{VariantIds: variantIds})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result.Data.Summaries
}

// Available is on-hand minus reserved, computed on read and never stored, so nothing can persist a
// value that disagrees with the two it comes from.
func TestVariantSummaryDerivesAvailableFromOnHandAndReserved(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName:        {quantRow(testVariantAId, testLocationAId, 120, 20)},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	summaries := summariseVariants(t, testVariantAId)

	summary := summaries[testVariantAId]
	assert.Equal(t, "120", summary.OnHand.String())
	assert.Equal(t, "20", summary.Reserved.String())
	assert.Equal(t, "100", summary.Available.String(),
		"available is on-hand minus reserved, not a third stored number")
}

// Reserved stock is part of on-hand rather than additional to it. Adding the two would double-count
// every promised unit, which is the mistake the port's documentation exists to prevent.
func TestVariantSummaryTreatsReservedAsPartOfOnHand(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName:        {quantRow(testVariantAId, testLocationAId, 10, 10)},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	summary := summariseVariants(t, testVariantAId)[testVariantAId]

	assert.Equal(t, "10", summary.OnHand.String())
	assert.Equal(t, "0", summary.Available.String(),
		"fully reserved stock is present but not available")
}

// A variant with no quants must come back zeroed rather than missing, so a caller never has to
// tell "no stock" apart from "not returned".
func TestVariantSummaryReportsVariantsWithNoStock(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName:        {quantRow(testVariantAId, testLocationAId, 5, 0)},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	summaries := summariseVariants(t, testVariantAId, testVariantBId)

	require.Contains(t, summaries, testVariantBId)
	assert.True(t, summaries[testVariantBId].OnHand.IsZero())
	assert.Equal(t, 0, summaries[testVariantBId].LocationCount)
}

// One product commonly has several quants in the same location — a lot, a package, an owner each
// make another row — and the page counts places, so the location counts once.
func TestVariantSummaryCountsALocationOnceAcrossSeveralQuants(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 4, 0),
			quantRow(testVariantAId, testLocationAId, 6, 0),
			quantRow(testVariantAId, testLocationBId, 5, 0),
		},
		models.InventoryLocationSchemaName: {
			internalLocationRow(testLocationAId),
			internalLocationRow(testLocationBId),
		},
	})

	summary := summariseVariants(t, testVariantAId)[testVariantAId]

	assert.Equal(t, "15", summary.OnHand.String(), "every quant contributes to the total")
	assert.Equal(t, 2, summary.LocationCount, "two places, three rows")
	assert.Equal(t, 1, summary.WarehouseCount, "both locations sit in the same warehouse")
}

// A zeroed quant marks somewhere stock used to be. Counting it would tell the reader the product
// is in a place it has already left.
func TestVariantSummaryIgnoresEmptiedLocationsInTheCount(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 7, 0),
			quantRow(testVariantAId, testLocationBId, 0, 0),
		},
		models.InventoryLocationSchemaName: {
			internalLocationRow(testLocationAId),
			internalLocationRow(testLocationBId),
		},
	})

	summary := summariseVariants(t, testVariantAId)[testVariantAId]

	assert.Equal(t, 1, summary.LocationCount, "an emptied location is not somewhere it is kept")
}

// The same id passed twice must not have its stock counted twice.
func TestVariantSummaryDedupesRequestedIds(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.StockQuantSchemaName:        {quantRow(testVariantAId, testLocationAId, 8, 0)},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	summaries := summariseVariants(t, testVariantAId, testVariantAId)

	assert.Len(t, summaries, 1)
	assert.Equal(t, "8", summaries[testVariantAId].OnHand.String())
}

// A template's total is the sum of its variants, and no quant is ever attributed to the template
// itself.
func TestTemplateSummaryIsTheSumOfItsVariants(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.ProductVariantSchemaName: {
			{
				models.ProductVariantFieldId:  testVariantAId,
				models.ProductVariantFieldSku: "SKU-A",
			},
			{
				models.ProductVariantFieldId:  testVariantBId,
				models.ProductVariantFieldSku: "SKU-B",
			},
		},
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 10, 2),
			quantRow(testVariantBId, testLocationAId, 20, 0),
		},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetTemplateSummary(
		callerContext(), itStock.GetTemplateSummaryQuery{TemplateId: testTemplateId})

	require.NoError(t, err)
	assert.Equal(t, "30", result.Data.Summary.OnHand.String())
	assert.Equal(t, "28", result.Data.Summary.Available.String())
	require.Len(t, result.Data.Variants, 2, "the breakdown accompanies the total")
}

// The aggregate exists for display. Location and warehouse counts are deliberately left off it,
// because two variants sharing an aisle would otherwise count that aisle twice.
func TestTemplateSummaryDoesNotSumPlaceCounts(t *testing.T) {
	useSchemaEngines(t, map[string][]dmodel.DynamicFields{
		models.ProductVariantSchemaName: {
			{models.ProductVariantFieldId: testVariantAId},
			{models.ProductVariantFieldId: testVariantBId},
		},
		models.StockQuantSchemaName: {
			quantRow(testVariantAId, testLocationAId, 1, 0),
			quantRow(testVariantBId, testLocationAId, 1, 0),
		},
		models.InventoryLocationSchemaName: {internalLocationRow(testLocationAId)},
	})

	service := &StockQuantDomainServiceImpl{}
	result, err := service.GetTemplateSummary(
		callerContext(), itStock.GetTemplateSummaryQuery{TemplateId: testTemplateId})

	require.NoError(t, err)
	assert.Equal(t, 0, result.Data.Summary.LocationCount,
		"summing place counts across variants would double-count a shared aisle")
	for _, row := range result.Data.Variants {
		assert.Equal(t, 1, row.Summary.LocationCount, "each variant keeps its own accurate count")
	}
}

// A template id is required. Falling back to "every variant" on an empty id would summarise the
// whole catalogue.
func TestTemplateSummaryRejectsAnEmptyTemplateId(t *testing.T) {
	service := &StockQuantDomainServiceImpl{}

	_, err := service.GetTemplateSummary(
		callerContext(), itStock.GetTemplateSummaryQuery{TemplateId: ""})

	assert.Error(t, err)
}
