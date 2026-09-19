package services

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// CheckWarehouseAvailability reports H, R, A and the shortage for one request, as of now. It takes
// no lock and writes nothing, so its answer is advisory: the reserve that follows recomputes the
// same figures under the guard and may still refuse.
func (this *StockReservationDomainServiceImpl) CheckWarehouseAvailability(
	ctx corectx.Context, query itStock.CheckWarehouseAvailabilityQuery,
) (*itStock.WarehouseAvailabilityReport, error) {
	vErrs := ft.NewClientErrors()
	if query.OrgId == "" || query.WarehouseId == "" || query.ProductVariantId == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
			"org_id, warehouse_id and product_variant_id are required"))
		return &itStock.WarehouseAvailabilityReport{ClientErrors: *vErrs}, nil
	}
	if query.Requested.IsNegative() {
		vErrs.Append(*ft.NewBusinessViolation(models.StockReservationSchemaName, ReasonReservationRequestMalformed,
			"requested must not be negative"))
		return &itStock.WarehouseAvailabilityReport{ClientErrors: *vErrs}, nil
	}

	now := time.Now().UTC()
	availability, err := ComputeWarehouseAvailability(ctx, GuardKey{
		OrgId: query.OrgId, WarehouseId: query.WarehouseId, ProductVariantId: query.ProductVariantId,
	}, now)
	if err != nil {
		return nil, err
	}
	return &itStock.WarehouseAvailabilityReport{
		EligibleOnHand:    availability.EligibleOnHand,
		EffectiveReserved: availability.EffectiveReserved,
		Available:         availability.Available,
		Shortage:          availability.Shortage(query.Requested),
		AsOf:              availability.AsOf,
	}, nil
}
