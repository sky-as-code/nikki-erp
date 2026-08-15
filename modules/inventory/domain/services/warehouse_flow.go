package services

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The warehouse flow topology: which locations a flow needs, and what ordered path goods take
// through them.
//
// This is configuration, expressed as pure functions over the flow setting. Nothing here reads or
// writes stock. The Stock movement engine asks what the path is and then creates the moves; a
// change to a flow provisions locations and leaves every existing quant and transfer alone.

// systemLocationSpec describes one location a warehouse creates for itself.
type systemLocationSpec struct {
	Code    string
	Purpose string
}

// The codes are the segment names that appear in a location path, so 'MAIN/Quality Control' reads
// the way a person would say it.
const (
	warehouseStockLocationCode   = "Stock"
	warehouseInputLocationCode   = "Input"
	warehouseQualityLocationCode = "Quality Control"
	warehousePackingLocationCode = "Packing"
	warehouseOutputLocationCode  = "Output"
)

// requiredSystemLocations lists what a warehouse must have for its two flows.
//
// Stock is always present: it is where goods live, and a warehouse without it could not hold
// anything. The rest depend on how many stops the flows declare.
func requiredSystemLocations(incomingFlow string, outgoingFlow string) []systemLocationSpec {
	specs := []systemLocationSpec{
		{Code: warehouseStockLocationCode, Purpose: models.InventoryLocationPurposeStorage},
	}
	specs = append(specs, incomingFlowLocations(incomingFlow)...)
	specs = append(specs, outgoingFlowLocations(outgoingFlow)...)
	return specs
}

// incomingFlowLocations lists the stops goods make on the way in, beyond Stock itself.
func incomingFlowLocations(flow string) []systemLocationSpec {
	switch flow {
	case models.WarehouseFlowTwoStep:
		return []systemLocationSpec{
			{Code: warehouseInputLocationCode, Purpose: models.InventoryLocationPurposeReceiving},
		}
	case models.WarehouseFlowThreeStep:
		return []systemLocationSpec{
			{Code: warehouseInputLocationCode, Purpose: models.InventoryLocationPurposeReceiving},
			{Code: warehouseQualityLocationCode, Purpose: models.InventoryLocationPurposeQuality},
		}
	default:
		return nil
	}
}

// outgoingFlowLocations lists the stops goods make on the way out, beyond Stock itself.
func outgoingFlowLocations(flow string) []systemLocationSpec {
	switch flow {
	case models.WarehouseFlowTwoStep:
		return []systemLocationSpec{
			{Code: warehouseOutputLocationCode, Purpose: models.InventoryLocationPurposeOutput},
		}
	case models.WarehouseFlowThreeStep:
		return []systemLocationSpec{
			{Code: warehousePackingLocationCode, Purpose: models.InventoryLocationPurposePacking},
			{Code: warehouseOutputLocationCode, Purpose: models.InventoryLocationPurposeOutput},
		}
	default:
		return nil
	}
}

// FlowLeg is one hop of a movement plan: goods leave one place and arrive at another.
//
// The endpoints outside the warehouse are named rather than resolved to ids here, because which
// vendor or customer location applies depends on the transaction, which is the movement engine's
// business and not the warehouse's.
type FlowLeg struct {
	FromCode string
	ToCode   string
}

// The two endpoints outside the warehouse. They are locations in the shared master with no
// warehouse of their own, which is why they are referred to by usage rather than by path.
const (
	FlowEndpointVendor   = "vendor"
	FlowEndpointCustomer = "customer"
)

// ResolveIncomingFlow returns the ordered path goods take from a vendor into stock.
//
// It is a pure read of configuration. Calling it moves nothing and creates nothing; the movement
// engine takes the legs and builds the transfer.
func ResolveIncomingFlow(flow string) []FlowLeg {
	stops := incomingFlowLocations(flow)

	from := FlowEndpointVendor
	legs := make([]FlowLeg, 0, len(stops)+1)
	for _, stop := range stops {
		legs = append(legs, FlowLeg{FromCode: from, ToCode: stop.Code})
		from = stop.Code
	}
	return append(legs, FlowLeg{FromCode: from, ToCode: warehouseStockLocationCode})
}

// ResolveOutgoingFlow returns the ordered path goods take from stock out to a customer.
func ResolveOutgoingFlow(flow string) []FlowLeg {
	stops := outgoingFlowLocations(flow)

	from := warehouseStockLocationCode
	legs := make([]FlowLeg, 0, len(stops)+1)
	for _, stop := range stops {
		legs = append(legs, FlowLeg{FromCode: from, ToCode: stop.Code})
		from = stop.Code
	}
	return append(legs, FlowLeg{FromCode: from, ToCode: FlowEndpointCustomer})
}

// obsoleteSystemLocations lists the stops a flow change leaves unused.
//
// They are reported so the caller can suspend them, never delete them: a location that once held
// goods is named by the moves that passed through it, and removing it would break the history that
// explains where stock went.
func obsoleteSystemLocations(previousFlow string, nextFlow string, outgoing bool) []systemLocationSpec {
	lister := incomingFlowLocations
	if outgoing {
		lister = outgoingFlowLocations
	}

	stillNeeded := map[string]bool{}
	for _, spec := range lister(nextFlow) {
		stillNeeded[spec.Code] = true
	}

	obsolete := make([]systemLocationSpec, 0)
	for _, spec := range lister(previousFlow) {
		if !stillNeeded[spec.Code] {
			obsolete = append(obsolete, spec)
		}
	}
	return obsolete
}

func isKnownFlow(flow string) bool {
	switch flow {
	case models.WarehouseFlowOneStep, models.WarehouseFlowTwoStep, models.WarehouseFlowThreeStep:
		return true
	default:
		return false
	}
}
