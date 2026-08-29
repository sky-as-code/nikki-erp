package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// DoScrap executes a draft scrap, removing the goods from usable stock: validate the document, lock
// the source balance and check what it can spare, generate the movement through
// ApplyCorrectionMovement, mark the document done.
//
// It is not left for a user to validate afterwards: a half-executed scrap would leave goods that
// are neither usable nor written off.
func (this *StockScrapDomainServiceImpl) DoScrap(
	ctx corectx.Context, scrapId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	var result *dyn.OpResult[dyn.MutateResultData]

	err := withQuantTransaction(ctx, func(tranxCtx corectx.Context) error {
		scrap, vErrs, err := this.loadScrap(tranxCtx, scrapId)
		if err != nil {
			return err
		}
		if vErrs.Count() > 0 {
			result = &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
			return nil
		}

		outcome, err := executeScrap(tranxCtx, *scrap)
		if err != nil {
			return err
		}
		result = outcome
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// executeScrap is the body of Do Scrap, once the transaction is open.
func executeScrap(
	ctx corectx.Context, scrap models.StockScrap,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	if derefString(scrap.GetStatus()) != models.StockScrapStatusDraft {
		return &dyn.OpResult[dyn.MutateResultData]{
			ClientErrors: *scrapViolation(
				"stock_scrap.not_draft", "only a draft scrap can be executed"),
		}, nil
	}

	quantity := orZero(scrap.GetQuantity())
	if quantity.LessThanOrEqual(decimal.Zero) {
		return &dyn.OpResult[dyn.MutateResultData]{
			ClientErrors: *scrapViolation(
				"stock_scrap.quantity_not_positive",
				"a scrap must remove a quantity greater than zero"),
		}, nil
	}

	if vErrs, err := assertScrappableStock(ctx, scrap, quantity); err != nil {
		return nil, err
	} else if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	moveId, vErrs, err := generateScrapMovement(ctx, scrap, quantity)
	if err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	if err := closeScrap(ctx, scrap, moveId); err != nil {
		return nil, err
	}
	return mutateOk(), nil
}

// assertScrappableStock locks the source balance and refuses to scrap what is not free. It checks
// Available(), not OnHand: reserved stock is promised to a transfer, and scrapping it would leave
// that transfer unable to ship. The lock is held for the rest of the transaction, so the movement
// applies against the same balance this check approved.
func assertScrappableStock(
	ctx corectx.Context, scrap models.StockScrap, quantity decimal.Decimal,
) (*ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	engine, err := engineFor(models.StockQuantSchemaName)
	if err != nil {
		return vErrs, err
	}

	locked, err := LockQuantsForUpdate(ctx, engine.ResourceRepository().GetBaseRepo(), QuantLockKey{
		OrgId:            model.Id(derefString(scrap.GetOrgId())),
		ProductVariantId: model.Id(derefString(scrap.GetProductVariantId())),
		LocationId:       model.Id(derefString(scrap.GetSourceLocationId())),
	})
	if err != nil {
		return vErrs, err
	}

	available := decimal.Zero
	for _, row := range locked {
		if !matchesScrapDimension(scrap, row) {
			continue
		}
		available = available.Add(row.Available())
	}

	if available.LessThan(quantity) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.StockScrapSchemaName, "stock_scrap.insufficient_stock",
			"only "+available.String()+" is available to scrap, which is less than the "+
				quantity.String()+" requested"))
	}
	return vErrs, nil
}

// matchesScrapDimension keeps the availability sum to the balances the scrap names. An empty
// lot/package/owner means "not tracked" and matches every row; a named one matches only its own,
// because scrapping lot A must not draw on lot B's stock.
func matchesScrapDimension(scrap models.StockScrap, row LockedQuant) bool {
	if lot := derefString(scrap.GetLotRef()); lot != "" && lot != row.LotRef {
		return false
	}
	if pkg := derefString(scrap.GetPackageRef()); pkg != "" && pkg != row.PackageRef {
		return false
	}
	if owner := derefString(scrap.GetOwnerRef()); owner != "" && owner != row.OwnerRef {
		return false
	}
	return true
}

// generateScrapMovement moves the goods to the org's scrap location. The direction never varies,
// unlike an adjustment's: goods always leave usable stock, which is what writes them off rather
// than merely moving them.
func generateScrapMovement(
	ctx corectx.Context, scrap models.StockScrap, quantity decimal.Decimal,
) (string, *ft.ClientErrors, error) {
	orgId := derefString(scrap.GetOrgId())

	// The document's scrap location comes from a client, so it is checked to be a scrap location
	// rather than trusted.
	destination, err := loadScrapDestination(ctx, scrap, orgId)
	if err != nil {
		return "", nil, err
	}
	if destination == "" {
		return "", scrapViolation(
			"stock_scrap.no_scrap_location",
			"this organisation has no scrap location for the goods to be written off to"), nil
	}

	result, vErrs, err := ApplyCorrectionMovement(ctx, CorrectionRequest{
		OrgId:                 orgId,
		ProductVariantId:      derefString(scrap.GetProductVariantId()),
		Quantity:              quantity,
		SourceLocationId:      derefString(scrap.GetSourceLocationId()),
		DestinationLocationId: destination,
		LotRef:                derefString(scrap.GetLotRef()),
		PackageRef:            derefString(scrap.GetPackageRef()),
		OwnerRef:              derefString(scrap.GetOwnerRef()),
		OriginReference:       "scrap:" + derefString(scrap.GetScrapNumber()),
		// False: a scrap is a write-off, not a count correction, and mixing them would make adjustment
		// figures include losses that were never a counting discrepancy.
		IsInventoryAdjustment: false,
	})
	if err != nil || (vErrs != nil && vErrs.Count() > 0) {
		return "", vErrs, err
	}
	return result.MoveId, ft.NewClientErrors(), nil
}

// loadScrapDestination resolves where the goods are written off to: the document's own
// scrap_location_id when it names a real scrap location, otherwise the org's seeded one. A location
// of the wrong type is ignored, since scrapping into an internal location would silently move
// usable stock instead of writing it off.
func loadScrapDestination(
	ctx corectx.Context, scrap models.StockScrap, orgId string,
) (string, error) {
	named := derefString(scrap.GetScrapLocationId())
	if named != "" {
		engine, err := engineFor(models.InventoryLocationSchemaName)
		if err != nil {
			return "", err
		}
		found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
			models.InventoryLocationFieldId: named,
		})
		if err != nil {
			return "", errors.Wrap(err, "loadScrapDestination")
		}
		if found != nil && found.HasData {
			location := models.NewInventoryLocationFrom(found.Data)
			if derefString(location.GetLocationUsage()) == models.InventoryLocationUsageScrap {
				return named, nil
			}
		}
	}

	fallback, err := FindLocationByType(ctx, orgId, models.InventoryLocationUsageScrap)
	if err != nil || fallback == nil {
		return "", err
	}
	return derefString(fallback.GetId()), nil
}

// closeScrap marks the document done and records the movement it generated. All three fields are
// written together: a done scrap with no move id is a write-off nobody could trace.
func closeScrap(ctx corectx.Context, scrap models.StockScrap, moveId string) error {
	engine, err := engineFor(models.StockScrapSchemaName)
	if err != nil {
		return err
	}

	_, err = engine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.StockScrapFieldId:          derefString(scrap.GetId()),
		models.StockScrapFieldStatus:      models.StockScrapStatusDone,
		models.StockScrapFieldMoveId:      moveId,
		models.StockScrapFieldCompletedAt: time.Now().UTC(),
		basemodel.FieldEtag:               derefString(scrap.GetEtag()),
	})
	return errors.Wrap(err, "closeScrap")
}
