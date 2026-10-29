package services

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
)

// The warehouse flow topology: which locations a flow needs and what ordered path goods take
// through them. Pure functions over the flow setting; nothing here reads or writes stock, and a
// flow change provisions locations while leaving every existing quant and transfer alone.

// systemLocationSpec describes one stop of a warehouse flow, a location the warehouse creates for
// itself under its root. Field is the warehouse FK that names it; Name is its initial name.
type systemLocationSpec struct {
	Name  string
	Field string
}

// The stops' names, which the flow legs also use to name them. "Stock" is the managed child of the
// root where goods live; it exists whatever the flows.
const (
	warehouseStockLocationCode   = "Stock"
	warehouseInputLocationCode   = "Input"
	warehouseQualityLocationCode = "Quality Control"
	warehousePackingLocationCode = "Packing"
	warehouseOutputLocationCode  = "Output"
)

var (
	inputStop   = systemLocationSpec{Name: warehouseInputLocationCode, Field: models.WarehouseFieldReceivingLocationId}
	qualityStop = systemLocationSpec{Name: warehouseQualityLocationCode, Field: models.WarehouseFieldQualityLocationId}
	packingStop = systemLocationSpec{Name: warehousePackingLocationCode, Field: models.WarehouseFieldPackingLocationId}
	outputStop  = systemLocationSpec{Name: warehouseOutputLocationCode, Field: models.WarehouseFieldOutputLocationId}
)

// incomingFlowLocations lists the stops goods make on the way in, beyond Stock itself.
func incomingFlowLocations(flow string) []systemLocationSpec {
	switch flow {
	case models.WarehouseFlowTwoStep:
		return []systemLocationSpec{inputStop}
	case models.WarehouseFlowThreeStep:
		return []systemLocationSpec{inputStop, qualityStop}
	default:
		return nil
	}
}

// outgoingFlowLocations lists the stops goods make on the way out, beyond Stock itself.
func outgoingFlowLocations(flow string) []systemLocationSpec {
	switch flow {
	case models.WarehouseFlowTwoStep:
		return []systemLocationSpec{outputStop}
	case models.WarehouseFlowThreeStep:
		return []systemLocationSpec{packingStop, outputStop}
	default:
		return nil
	}
}

// stopField names the warehouse FK holding a stop named in a flow leg: "Stock" is the stock
// location, any other stop is looked up in the three-step flows, which name every stop there is.
func stopField(stopName string) string {
	if stopName == warehouseStockLocationCode {
		return models.WarehouseFieldStockLocationId
	}
	for _, spec := range []systemLocationSpec{inputStop, qualityStop, packingStop, outputStop} {
		if spec.Name == stopName {
			return spec.Field
		}
	}
	return ""
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
		legs = append(legs, FlowLeg{FromCode: from, ToCode: stop.Name})
		from = stop.Name
	}
	return append(legs, FlowLeg{FromCode: from, ToCode: warehouseStockLocationCode})
}

// ResolveOutgoingFlow returns the ordered path goods take from stock out to a customer.
func ResolveOutgoingFlow(flow string) []FlowLeg {
	stops := outgoingFlowLocations(flow)

	from := warehouseStockLocationCode
	legs := make([]FlowLeg, 0, len(stops)+1)
	for _, stop := range stops {
		legs = append(legs, FlowLeg{FromCode: from, ToCode: stop.Name})
		from = stop.Name
	}
	return append(legs, FlowLeg{FromCode: from, ToCode: FlowEndpointCustomer})
}

// obsoleteSystemLocations lists the stops a flow change leaves unused, so the caller can archive
// them. Never delete them: the moves that passed through still name them, and removing one would
// break the history explaining where stock went.
func obsoleteSystemLocations(previousFlow string, nextFlow string, outgoing bool) []systemLocationSpec {
	lister := incomingFlowLocations
	if outgoing {
		lister = outgoingFlowLocations
	}

	stillNeeded := map[string]bool{}
	for _, spec := range lister(nextFlow) {
		stillNeeded[spec.Name] = true
	}

	obsolete := make([]systemLocationSpec, 0)
	for _, spec := range lister(previousFlow) {
		if !stillNeeded[spec.Name] {
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
