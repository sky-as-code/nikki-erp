package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Shared reads and small helpers behind the product stock summary. A quant records where stock is
// but not what kind of place that is, so locations are read once per summary call and cached in a
// map, never once per quant.

// locationScanPageSize is how many locations are read at a time when classifying them.
const locationScanPageSize = 500

// maxLocationScanPages bounds the location scans, so a page render cannot trigger an unbounded
// read.
const maxLocationScanPages = 20

// internalLocationIds returns the locations that hold stock we own. Forecast direction depends on
// it: without the distinction an internal transfer would count as both an arrival and a departure.
func (this *StockQuantDomainServiceImpl) internalLocationIds(
	ctx corectx.Context,
) (map[string]bool, error) {
	return this.locationIdsWithUsage(ctx, models.InventoryLocationUsageInternal)
}

// transitLocationIds returns the locations representing stock that has left somewhere and not yet
// arrived anywhere. In-transit quantity is derived from balances sitting at these, never stored as
// a counter of its own.
func (this *StockQuantDomainServiceImpl) transitLocationIds(
	ctx corectx.Context,
) (map[string]bool, error) {
	return this.locationIdsWithUsage(ctx, models.InventoryLocationUsageTransit)
}

func (this *StockQuantDomainServiceImpl) locationIdsWithUsage(
	ctx corectx.Context, usage string,
) (map[string]bool, error) {
	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldLocationUsage, dmodel.Equals, usage),
	)

	found := map[string]bool{}
	for page := 0; page < maxLocationScanPages; page++ {
		result, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  locationScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "locationIdsWithUsage")
		}
		if result == nil || !result.HasData || len(result.Data.Items) == 0 {
			break
		}

		for _, row := range result.Data.Items {
			location := models.NewInventoryLocationFrom(row)
			if id := derefId(location.GetId()); id != "" {
				found[id] = true
			}
		}

		if len(result.Data.Items) < locationScanPageSize {
			break
		}
	}
	return found, nil
}

// locationDetail is the display metadata a stock row needs about the place it sits in. Status is
// carried because a suspended location keeps its stock and keeps being shown, and the UI needs it
// to badge the row.
type locationDetail struct {
	Id          string
	Code        string
	Name        string
	Status      string
	WarehouseId string
}

// loadLocationDetails reads the given locations in one batched search.
func (this *StockQuantDomainServiceImpl) loadLocationDetails(
	ctx corectx.Context, locationIds []string,
) (map[string]locationDetail, error) {
	details := map[string]locationDetail{}
	if len(locationIds) == 0 {
		return details, nil
	}

	engine, err := engineFor(models.InventoryLocationSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.InventoryLocationFieldId, dmodel.In, toAnySlice(locationIds)...),
	)

	for page := 0; page < maxLocationScanPages; page++ {
		result, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  locationScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "loadLocationDetails")
		}
		if result == nil || !result.HasData || len(result.Data.Items) == 0 {
			break
		}

		for _, row := range result.Data.Items {
			location := models.NewInventoryLocationFrom(row)
			id := derefId(location.GetId())
			if id == "" {
				continue
			}
			details[id] = locationDetail{
				Id:          id,
				Code:        derefString(location.GetCode()),
				Name:        langJsonToString(location.GetName()),
				Status:      derefString(location.GetStatus()),
				WarehouseId: derefId(location.GetWarehouseId()),
			}
		}

		if len(result.Data.Items) < locationScanPageSize {
			break
		}
	}
	return details, nil
}

// warehouseIdsOfLocations maps each location to the warehouse holding it. An empty string means no
// warehouse, which is legitimate: vendor, customer, transit and inventory-loss locations sit
// outside one.
func (this *StockQuantDomainServiceImpl) warehouseIdsOfLocations(
	ctx corectx.Context, locationIds []string,
) (map[string]string, error) {
	details, err := this.loadLocationDetails(ctx, locationIds)
	if err != nil {
		return nil, err
	}

	warehouseOf := make(map[string]string, len(details))
	for id, detail := range details {
		warehouseOf[id] = detail.WarehouseId
	}
	return warehouseOf, nil
}

// warehouseDetail is the display metadata a by-warehouse row needs.
type warehouseDetail struct {
	Code   string
	Name   string
	Status string
}

// loadWarehouseDetails reads the given warehouses in one batched search.
func (this *StockQuantDomainServiceImpl) loadWarehouseDetails(
	ctx corectx.Context, warehouseIds []string,
) (map[string]warehouseDetail, error) {
	details := map[string]warehouseDetail{}
	if len(warehouseIds) == 0 {
		return details, nil
	}

	engine, err := engineFor(models.WarehouseSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.WarehouseFieldId, dmodel.In, toAnySlice(warehouseIds)...),
	)

	result, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  locationScanPageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadWarehouseDetails")
	}
	if result == nil || !result.HasData {
		return details, nil
	}

	for _, row := range result.Data.Items {
		warehouse := models.NewWarehouseFrom(row)
		id := derefId(warehouse.GetId())
		if id == "" {
			continue
		}
		details[id] = warehouseDetail{
			Code:   derefString(warehouse.GetCode()),
			Name:   langJsonToString(warehouse.GetName()),
			Status: derefString(warehouse.GetStatus()),
		}
	}
	return details, nil
}

// forecastIgnoredMoveStatuses are the move states that contribute nothing to a forecast: cancelled
// moves never happen, and drafts are not commitments. Done moves are deliberately absent — they are
// read in the same pass for the last movement date and skipped for forecast at the point of use,
// their effect already being inside on-hand.
func forecastIgnoredMoveStatuses() []string {
	return []string{models.StockMoveStatusDraft, models.StockMoveStatusCancelled}
}

// dedupeNonEmpty removes blanks and repeats while preserving order, so a caller passing the same id
// twice does not have its stock counted twice.
func dedupeNonEmpty(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func toAnySlice(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}

func mapKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}
