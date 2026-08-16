package services

import (
	"sort"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// Where one variant's stock physically sits, by warehouse and by location.
//
// Both are groupings of the same quants: by-warehouse is by-location rolled up one level. They are
// two methods because the Product UI shows a warehouse total first and drills into the locations
// underneath (CR §9.1, §9.2), and doing the rollup here keeps the caller from re-deriving it.
//
// Suspended warehouses and locations appear in both. Suspension governs whether something may be
// chosen for a *new* operation, not whether the stock already there exists — hiding it would make
// the product page disagree with the shelf (CR §9.4, §9.5, AC-PROD-INT-017/018, TS-PROD-05). The
// status travels with each row so the UI can badge it.

// quantPlacement is one variant's balance at one location, before it is grouped.
type quantPlacement struct {
	LocationId string
	OnHand     decimal.Decimal
	Reserved   decimal.Decimal
}

// GetStockByWarehouse groups a variant's stock by the warehouse holding it.
func (this *StockQuantDomainServiceImpl) GetStockByWarehouse(
	ctx corectx.Context, query itStock.GetStockByWarehouseQuery,
) (*itStock.GetStockByWarehouseResult, error) {
	if query.VariantId == "" {
		return nil, errors.New("GetStockByWarehouse requires a variant id")
	}

	placements, err := this.readVariantPlacements(ctx, query.VariantId)
	if err != nil {
		return nil, err
	}

	locationIds := make([]string, 0, len(placements))
	for locationId := range placements {
		locationIds = append(locationIds, locationId)
	}
	locations, err := this.loadLocationDetails(ctx, locationIds)
	if err != nil {
		return nil, err
	}

	// Grouped by warehouse id, with "" its own group: stock at a vendor, customer, transit or
	// loss location belongs to no warehouse and must not be attributed to one.
	grouped := map[string]*itStock.WarehouseStockRow{}
	for locationId, placement := range placements {
		warehouseId := locations[locationId].WarehouseId
		row, seen := grouped[warehouseId]
		if !seen {
			row = &itStock.WarehouseStockRow{}
			grouped[warehouseId] = row
		}
		row.OnHand = row.OnHand.Add(placement.OnHand)
		row.Reserved = row.Reserved.Add(placement.Reserved)
	}

	warehouseIds := make([]string, 0, len(grouped))
	for warehouseId := range grouped {
		if warehouseId != "" {
			warehouseIds = append(warehouseIds, warehouseId)
		}
	}
	warehouses, err := this.loadWarehouseDetails(ctx, warehouseIds)
	if err != nil {
		return nil, err
	}

	rows := make([]itStock.WarehouseStockRow, 0, len(grouped))
	for warehouseId, row := range grouped {
		row.Available = row.OnHand.Sub(row.Reserved)
		if warehouseId != "" {
			id := model.Id(warehouseId)
			row.WarehouseId = &id
			detail := warehouses[warehouseId]
			row.WarehouseCode = detail.Code
			row.WarehouseName = detail.Name
			row.WarehouseStatus = detail.Status
		}
		rows = append(rows, *row)
	}

	sortWarehouseRows(rows)

	return &itStock.GetStockByWarehouseResult{
		HasData: true,
		Data:    itStock.GetStockByWarehouseResultData{Rows: rows},
	}, nil
}

// GetStockByLocation lists a variant's stock per location, optionally within one warehouse.
func (this *StockQuantDomainServiceImpl) GetStockByLocation(
	ctx corectx.Context, query itStock.GetStockByLocationQuery,
) (*itStock.GetStockByLocationResult, error) {
	if query.VariantId == "" {
		return nil, errors.New("GetStockByLocation requires a variant id")
	}

	placements, err := this.readVariantPlacements(ctx, query.VariantId)
	if err != nil {
		return nil, err
	}

	locationIds := make([]string, 0, len(placements))
	for locationId := range placements {
		locationIds = append(locationIds, locationId)
	}
	locations, err := this.loadLocationDetails(ctx, locationIds)
	if err != nil {
		return nil, err
	}

	rows := make([]itStock.LocationStockRow, 0, len(placements))
	for locationId, placement := range placements {
		detail := locations[locationId]
		if query.WarehouseId != "" && detail.WarehouseId != query.WarehouseId {
			continue
		}

		row := itStock.LocationStockRow{
			LocationId:     model.Id(locationId),
			LocationCode:   detail.Code,
			LocationName:   detail.Name,
			LocationStatus: detail.Status,
			OnHand:         placement.OnHand,
			Reserved:       placement.Reserved,
			Available:      placement.OnHand.Sub(placement.Reserved),
		}
		if detail.WarehouseId != "" {
			id := model.Id(detail.WarehouseId)
			row.WarehouseId = &id
		}
		rows = append(rows, row)
	}

	sortLocationRows(rows)

	return &itStock.GetStockByLocationResult{
		HasData: true,
		Data:    itStock.GetStockByLocationResultData{Rows: rows},
	}, nil
}

// readVariantPlacements totals one variant's quants per location.
//
// Several quants commonly share a location — one per lot, package and owner combination — so they
// are summed rather than listed: a product page shows how much is in a place, not how many rows
// the database uses to record it.
//
// Locations holding nothing are dropped. A zeroed quant marks somewhere stock used to be, and
// listing it would pad the page with empty rows.
func (this *StockQuantDomainServiceImpl) readVariantPlacements(
	ctx corectx.Context, variantId string,
) (map[string]quantPlacement, error) {
	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.StockQuantFieldProductVariantId, dmodel.Equals, variantId),
	)

	placements := map[string]quantPlacement{}
	for page := 0; page < maxSummaryQuantPages; page++ {
		found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  summaryScanPageSize,
		})
		if err != nil {
			return nil, errors.Wrap(err, "readVariantPlacements")
		}
		if found == nil || !found.HasData || len(found.Data.Items) == 0 {
			break
		}

		for _, row := range found.Data.Items {
			quant := models.NewStockQuantFrom(row)
			locationId := derefId(quant.GetLocationId())
			if locationId == "" {
				continue
			}

			placement := placements[locationId]
			placement.LocationId = locationId
			placement.OnHand = placement.OnHand.Add(derefDecimal(quant.GetOnHandQuantity()))
			placement.Reserved = placement.Reserved.Add(derefDecimal(quant.GetReservedQuantity()))
			placements[locationId] = placement
		}

		if len(found.Data.Items) < summaryScanPageSize {
			break
		}
	}

	for locationId, placement := range placements {
		if placement.OnHand.IsZero() && placement.Reserved.IsZero() {
			delete(placements, locationId)
		}
	}
	return placements, nil
}

// sortWarehouseRows orders by descending on-hand, then by code.
//
// Map iteration is random in Go, so without this the same product would list its warehouses in a
// different order on every page load. Largest holding first is what a reader wants; the code
// breaks ties so the order is stable.
func sortWarehouseRows(rows []itStock.WarehouseStockRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		if cmp := rows[j].OnHand.Cmp(rows[i].OnHand); cmp != 0 {
			return cmp < 0
		}
		return rows[i].WarehouseCode < rows[j].WarehouseCode
	})
}

func sortLocationRows(rows []itStock.LocationStockRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		if cmp := rows[j].OnHand.Cmp(rows[i].OnHand); cmp != 0 {
			return cmp < 0
		}
		return rows[i].LocationCode < rows[j].LocationCode
	})
}
