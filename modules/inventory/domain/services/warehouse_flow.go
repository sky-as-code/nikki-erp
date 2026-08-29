package services

import (
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The warehouse flow topology: which locations a flow needs and what ordered path goods take
// through them. Pure functions over the flow setting; nothing here reads or writes stock, and a
// flow change provisions locations while leaving every existing quant and transfer alone.

// systemLocationSpec describes one location a warehouse creates for itself.
type systemLocationSpec struct {
	Code    string
	Purpose string
}

// The codes are the segment names that appear in a location path, e.g. 'MAIN/Quality Control'.
const (
	warehouseStockLocationCode   = "Stock"
	warehouseInputLocationCode   = "Input"
	warehouseQualityLocationCode = "Quality Control"
	warehousePackingLocationCode = "Packing"
	warehouseOutputLocationCode  = "Output"
)

// requiredSystemLocations lists what a warehouse must have for its two flows. Stock is always
// present, being where goods live; the rest depend on how many stops the flows declare.
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

// FlowLeg is one hop of a movement plan. Endpoints outside the warehouse are named rather than
// resolved to ids: which vendor or customer location applies depends on the transaction, and the
// movement engine decides that.
type FlowLeg struct {
	FromCode string
	ToCode   string
}

// The two endpoints outside the warehouse: locations in the shared master with no warehouse of
// their own, referred to by usage rather than by path.
const (
	FlowEndpointVendor   = "vendor"
	FlowEndpointCustomer = "customer"
)

// ResolveIncomingFlow returns the ordered path goods take from a vendor into stock. A pure read of
// configuration: it moves nothing, and the movement engine builds the transfer from the legs.
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

// obsoleteSystemLocations lists the stops a flow change leaves unused, so the caller can suspend
// them. Never delete them: the moves that passed through still name them, and removing one would
// break the history explaining where stock went.
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
