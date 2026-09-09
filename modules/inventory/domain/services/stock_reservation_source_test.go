package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

// A hold taken for another module is found again by its source pair, so a request missing either
// half is refused before anything is written: the transfer would exist, hold stock, and be
// unreachable by the only key its owner has.
func TestSourceReservationRequiresItsDemandKey(t *testing.T) {
	wellFormed := itStock.SourceReservationRequest{
		SourceType:      "sales_fulfillment",
		SourceId:        "01ABC",
		OrgId:           "01ORG",
		LocationId:      "01LOC",
		OperationTypeId: "01OPT",
		Items: []itStock.SourceReservationItem{{
			ProductVariantId: "01VAR",
			Quantity:         decimal.NewFromInt(1),
		}},
	}
	assert.Nil(t, assertSourceReservationWellFormed(wellFormed))

	missingType := wellFormed
	missingType.SourceType = ""
	assert.NotNil(t, assertSourceReservationWellFormed(missingType),
		"a hold with no source type could never be found by its owner")

	missingId := wellFormed
	missingId.SourceId = ""
	assert.NotNil(t, assertSourceReservationWellFormed(missingId),
		"a hold with no source id could never be found by its owner")
}

// The location is the caller's decision and Inventory has no basis to guess one, so it is required.
// Without it the transfer would inherit the operation type's default and hold stock in a warehouse
// nobody promised the customer.
func TestSourceReservationRequiresALocation(t *testing.T) {
	request := itStock.SourceReservationRequest{
		SourceType:      "sales_fulfillment",
		SourceId:        "01ABC",
		OperationTypeId: "01OPT",
		Items: []itStock.SourceReservationItem{{
			ProductVariantId: "01VAR",
			Quantity:         decimal.NewFromInt(1),
		}},
	}
	assert.NotNil(t, assertSourceReservationWellFormed(request))
}

// An empty hold would be a document that reserves nothing and reports success, so it is refused for
// the same reason CreateWithMoves refuses an empty transfer.
func TestSourceReservationRefusesNoItems(t *testing.T) {
	request := itStock.SourceReservationRequest{
		SourceType:      "sales_fulfillment",
		SourceId:        "01ABC",
		LocationId:      "01LOC",
		OperationTypeId: "01OPT",
	}
	result := assertSourceReservationWellFormed(request)
	assert.NotNil(t, result)
	assert.Positive(t, result.ClientErrors.Count())
}

// A refusal must carry ClientErrors and no Go error, which is what makes the REST layer answer 400
// rather than 500: choosing a location that cannot supply the goods is the caller's to fix.
func TestSourceReservationRefusalIsAClientError(t *testing.T) {
	result := assertSourceReservationWellFormed(itStock.SourceReservationRequest{})
	assert.NotNil(t, result)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Empty(t, result.InventoryReference, "a refused hold reserves nothing")
	assert.False(t, result.FullyReserved)
}

// A reallocation must name where the stock is going. It deliberately does NOT name where it is
// coming from: the origin is read from the demand's current holds, so the request cannot disagree
// with reality about what is being moved.
func TestReallocationRequiresADestination(t *testing.T) {
	wellFormed := itStock.ReservationReallocationRequest{
		SourceType:      "sales_fulfillment",
		SourceId:        "01ABC",
		ToLocationId:    "01LOC2",
		OperationTypeId: "01OPT",
		Items: []itStock.SourceReservationItem{{
			ProductVariantId: "01VAR",
			Quantity:         decimal.NewFromInt(1),
		}},
	}
	assert.Nil(t, assertReallocationWellFormed(wellFormed))

	noDestination := wellFormed
	noDestination.ToLocationId = ""
	assert.NotNil(t, assertReallocationWellFormed(noDestination))
}

// The deadlock guard. Two reallocations moving stock between the same pair of locations in opposite
// directions must take the row locks in the SAME order, or each holds half of what the other needs
// and neither finishes. The ordering is by location then variant, so it cannot depend on which way
// the caller happens to be moving the goods.
func TestBothSidesAreLockedInOneDeterministicOrder(t *testing.T) {
	forward := []QuantLockKey{
		{OrgId: "01ORG", ProductVariantId: "VAR-B", LocationId: "LOC-2"},
		{OrgId: "01ORG", ProductVariantId: "VAR-A", LocationId: "LOC-1"},
		{OrgId: "01ORG", ProductVariantId: "VAR-A", LocationId: "LOC-2"},
	}
	// The same three rows, discovered in the order the opposite move would find them.
	backward := []QuantLockKey{
		{OrgId: "01ORG", ProductVariantId: "VAR-A", LocationId: "LOC-1"},
		{OrgId: "01ORG", ProductVariantId: "VAR-A", LocationId: "LOC-2"},
		{OrgId: "01ORG", ProductVariantId: "VAR-B", LocationId: "LOC-2"},
	}

	sortQuantLockKeys(forward)
	sortQuantLockKeys(backward)

	assert.Equal(t, forward, backward,
		"two reallocations over the same rows must queue, not deadlock: both orderings must agree")
	assert.Equal(t, "LOC-1", forward[0].LocationId, "ordered by location first")
	assert.Equal(t, "VAR-A", forward[1].ProductVariantId, "then by variant within a location")
}

func wellFormedReallocation() itStock.ReservationReallocationRequest {
	return itStock.ReservationReallocationRequest{
		SourceType:      "sales_fulfillment",
		SourceId:        "01ABC",
		ToLocationId:    "01LOC2",
		OperationTypeId: "01OPT",
		Items: []itStock.SourceReservationItem{{
			ProductVariantId: "01VAR",
			Quantity:         decimal.NewFromInt(1),
		}},
	}
}

// The remaining branches of the reallocation guard. Each of these would otherwise fail deep inside a
// transaction that has already taken row locks at both locations, which is a far more expensive way
// to learn the request was never answerable.
func TestReallocationRequiresItsDemandOperationTypeAndItems(t *testing.T) {
	cases := map[string]func(*itStock.ReservationReallocationRequest){
		"no source type":    func(r *itStock.ReservationReallocationRequest) { r.SourceType = "" },
		"no source id":      func(r *itStock.ReservationReallocationRequest) { r.SourceId = "" },
		"no operation type": func(r *itStock.ReservationReallocationRequest) { r.OperationTypeId = "" },
		"no items":          func(r *itStock.ReservationReallocationRequest) { r.Items = nil },
	}

	for name, break_ := range cases {
		request := wellFormedReallocation()
		break_(&request)

		refusal := assertReallocationWellFormed(request)

		assert.NotNil(t, refusal, "a reallocation with %s must be refused", name)
	}
}

// A destination that cannot supply the whole quantity is a REFUSAL, not a fault: the customer keeps
// the hold they already had, and the caller is told why in terms it can show a user. The transaction
// rolls back around this, so the message is the only trace the attempt leaves.
func TestAReallocationShortageIsARefusalNamingTheDestination(t *testing.T) {
	vErrs := reallocationShortage("01LOC-DEST")

	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "stock_transfer.reallocation_insufficient_stock", vErrs[0].Key)
	assert.Contains(t, vErrs[0].Message, "01LOC-DEST",
		"the refusal must name the location that could not supply the stock")
}
