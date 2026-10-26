package external

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// SourceTypeSalesFulfillment is how Sales names itself to Inventory when it holds stock. It is a
// plain string on both sides, deliberately: Inventory stores it to attribute a hold and must never
// resolve it against a Sales table, which is what would make the two modules undeployable apart.
const SourceTypeSalesFulfillment = "sales_fulfillment"

// The reservation half of the adapter. It differs from the intent half in what it lets Inventory
// decide: those methods hand over a sales order and let an operation type's defaults choose the
// location, while these name the location outright, because a kiosk fulfillment is a promise about
// ONE machine and a hold anywhere else would not keep it.
//
// The refusal contract is the one fulfillment.go documents: Inventory reports a business refusal as
// ClientErrors with a nil Go error, and those become Accepted false carrying Inventory's own words.
// Only a genuine fault propagates as an error, and Sales retries that rather than telling a customer
// their goods are unavailable because a database was briefly unreachable.

func (this *fulfillmentAdapter) ReserveForFulfillment(
	ctx corectx.Context, request itExt.FulfillmentReservationRequest,
) (*itExt.FulfillmentReservationResponse, error) {
	operationTypeId, refusal, err := this.outgoingOperationType(ctx)
	if err != nil || refusal != nil {
		return refusal, err
	}

	result, err := this.transfers.ReserveForSource(ctx, itStock.SourceReservationRequest{
		SourceType:      SourceTypeSalesFulfillment,
		SourceId:        request.FulfillmentId,
		OrgId:           request.OrgId,
		LocationId:      request.LocationId,
		OperationTypeId: operationTypeId,
		OriginReference: request.FulfillmentId,
		ExpiresAt:       request.ExpiresAt,
		Items:           reservationItemsOf(request.Items),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reserving stock for fulfillment %s", request.FulfillmentId)
	}
	return reservationResponseOf(result), nil
}

func (this *fulfillmentAdapter) ReallocateReservation(
	ctx corectx.Context, request itExt.FulfillmentReservationRequest,
) (*itExt.FulfillmentReservationResponse, error) {
	operationTypeId, refusal, err := this.outgoingOperationType(ctx)
	if err != nil || refusal != nil {
		return refusal, err
	}

	// The origin location is not passed: Inventory reads it from the moves the hold already has, so
	// Sales cannot name a source that disagrees with what is actually claimed.
	result, err := this.transfers.ReallocateReservation(ctx, itStock.ReservationReallocationRequest{
		SourceType:      SourceTypeSalesFulfillment,
		SourceId:        request.FulfillmentId,
		OrgId:           request.OrgId,
		ToLocationId:    request.LocationId,
		OperationTypeId: operationTypeId,
		OriginReference: request.FulfillmentId,
		ExpiresAt:       request.ExpiresAt,
		Items:           reservationItemsOf(request.Items),
	})
	if err != nil {
		return nil, errors.Wrapf(err, "reallocating the reservation of fulfillment %s", request.FulfillmentId)
	}
	return reservationResponseOf(result), nil
}

// ReleaseFulfillmentReservation gives the held stock back. Releasing a hold that is already gone
// answers success: a cancel arriving after the expiry sweep has run must not fail, because both
// wanted the same end state and there is nothing left for a caller to do about it.
func (this *fulfillmentAdapter) ReleaseFulfillmentReservation(
	ctx corectx.Context, sourceId string,
) (*itExt.FulfillmentReservationResponse, error) {
	result, err := this.transfers.ReleaseReservationBySource(ctx, SourceTypeSalesFulfillment, sourceId)
	if err != nil {
		return nil, errors.Wrapf(err, "releasing the reservation of fulfillment %s", sourceId)
	}
	if result != nil && result.ClientErrors.Count() > 0 {
		return &itExt.FulfillmentReservationResponse{
			Accepted:      false,
			FailureReason: describeViolations(&result.ClientErrors),
		}, nil
	}
	return &itExt.FulfillmentReservationResponse{Accepted: true}, nil
}

// CheckAvailabilityByLocations answers per location without taking a lock, so the reply is true of
// the instant it was read and of no instant after it. Callers use it to shortlist kiosks for a
// customer to choose between; only a reservation secures anything, and this deliberately hands back
// no reference that could be mistaken for one.
func (this *fulfillmentAdapter) CheckAvailabilityByLocations(
	ctx corectx.Context, query itExt.AvailabilityQuery,
) (*itExt.AvailabilityResult, error) {
	if len(query.LocationIds) == 0 || len(query.Items) == 0 {
		return &itExt.AvailabilityResult{}, nil
	}

	available, err := this.availability.AvailableByLocations(ctx, itStock.LocationAvailabilityQuery{
		OrgId:       query.OrgId,
		LocationIds: query.LocationIds,
		VariantIds:  variantIdsOf(query.Items),
	})
	if err != nil {
		return nil, errors.Wrap(err, "reading stock availability by location")
	}

	return &itExt.AvailabilityResult{
		Locations: availabilityPerLocation(query, available),
	}, nil
}

// availabilityPerLocation answers for every location that was ASKED about, not only for those the
// query returned rows for. A location holding none of a wanted variant has no quant row at all, and
// omitting it would read to a caller as "not considered" rather than "cannot supply".
func availabilityPerLocation(
	query itExt.AvailabilityQuery, available map[string]map[string]decimal.Decimal,
) []itExt.LocationAvailability {
	locations := make([]itExt.LocationAvailability, 0, len(query.LocationIds))
	for _, locationId := range query.LocationIds {
		byVariant := available[locationId]

		shortages := make([]itExt.AvailabilityShortage, 0)
		for _, item := range query.Items {
			onHand := byVariant[item.ProductVariantId]
			if onHand.LessThan(item.Quantity) {
				shortages = append(shortages, itExt.AvailabilityShortage{
					ProductVariantId: item.ProductVariantId,
					Requested:        item.Quantity,
					Available:        onHand,
				})
			}
		}

		locations = append(locations, itExt.LocationAvailability{
			LocationId:    locationId,
			CanFulfillAll: len(shortages) == 0,
			Shortages:     shortages,
		})
	}
	return locations
}

// variantIdsOf de-duplicates, because one fulfillment may name the same variant on two lines and
// asking Inventory about it twice would widen the query for no extra answer.
func variantIdsOf(items []itExt.AvailabilityItem) []string {
	seen := make(map[string]struct{}, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if _, done := seen[item.ProductVariantId]; done {
			continue
		}
		seen[item.ProductVariantId] = struct{}{}
		ids = append(ids, item.ProductVariantId)
	}
	return ids
}

// outgoingOperationType resolves the type every reservation is raised against, answering a refusal
// rather than an error when the deployment has not configured one - the same reasoning as
// notConfigured: a 500 sends an operator hunting a software fault instead of a settings gap.
func (this *fulfillmentAdapter) outgoingOperationType(
	ctx corectx.Context,
) (string, *itExt.FulfillmentReservationResponse, error) {
	operationTypeId, err := this.operationTypes.OutgoingOperationTypeId(ctx)
	if err != nil {
		return "", nil, err
	}
	if operationTypeId == "" {
		return "", &itExt.FulfillmentReservationResponse{
			Accepted:      false,
			FailureReason: notConfigured("outgoing").FailureReason,
		}, nil
	}
	return operationTypeId, nil, nil
}

func reservationItemsOf(items []itExt.FulfillmentReservationItem) []itStock.SourceReservationItem {
	out := make([]itStock.SourceReservationItem, 0, len(items))
	for _, item := range items {
		out = append(out, itStock.SourceReservationItem{
			SourceItemId:     item.FulfillmentItemId,
			ProductVariantId: item.ProductVariantId,
			UomId:            item.UomId,
			Quantity:         item.Quantity,
		})
	}
	return out
}

// reservationResponseOf keeps FullyReserved and Accepted apart. Inventory accepting the request
// says it understood it; holding all of it is a different fact, and a caller that promised a
// customer a specific kiosk refuses on the second even though the first succeeded.
func reservationResponseOf(
	result *itStock.SourceReservationResult,
) *itExt.FulfillmentReservationResponse {
	if result == nil {
		return &itExt.FulfillmentReservationResponse{
			Accepted:      false,
			FailureReason: "inventory answered nothing about the reservation",
		}
	}
	if result.ClientErrors.Count() > 0 {
		return &itExt.FulfillmentReservationResponse{
			Accepted:           false,
			InventoryReference: result.InventoryReference,
			FailureReason:      describeViolations(&result.ClientErrors),
		}
	}
	return &itExt.FulfillmentReservationResponse{
		Accepted:           true,
		InventoryReference: result.InventoryReference,
		FullyReserved:      result.FullyReserved,
	}
}
