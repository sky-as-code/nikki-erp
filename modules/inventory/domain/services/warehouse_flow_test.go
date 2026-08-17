package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// The three-step paths are named in the change request as TS-09 and AC-CR-LOC-028/029, so they are
// asserted verbatim rather than derived.
func TestResolveIncomingFlow(t *testing.T) {
	assert.Equal(t, []FlowLeg{
		{FromCode: FlowEndpointVendor, ToCode: "Stock"},
	}, ResolveIncomingFlow(models.WarehouseFlowOneStep))

	assert.Equal(t, []FlowLeg{
		{FromCode: FlowEndpointVendor, ToCode: "Input"},
		{FromCode: "Input", ToCode: "Stock"},
	}, ResolveIncomingFlow(models.WarehouseFlowTwoStep))

	assert.Equal(t, []FlowLeg{
		{FromCode: FlowEndpointVendor, ToCode: "Input"},
		{FromCode: "Input", ToCode: "Quality Control"},
		{FromCode: "Quality Control", ToCode: "Stock"},
	}, ResolveIncomingFlow(models.WarehouseFlowThreeStep))
}

func TestResolveOutgoingFlow(t *testing.T) {
	assert.Equal(t, []FlowLeg{
		{FromCode: "Stock", ToCode: FlowEndpointCustomer},
	}, ResolveOutgoingFlow(models.WarehouseFlowOneStep))

	assert.Equal(t, []FlowLeg{
		{FromCode: "Stock", ToCode: "Output"},
		{FromCode: "Output", ToCode: FlowEndpointCustomer},
	}, ResolveOutgoingFlow(models.WarehouseFlowTwoStep))

	assert.Equal(t, []FlowLeg{
		{FromCode: "Stock", ToCode: "Packing"},
		{FromCode: "Packing", ToCode: "Output"},
		{FromCode: "Output", ToCode: FlowEndpointCustomer},
	}, ResolveOutgoingFlow(models.WarehouseFlowThreeStep))
}

// Every path starts and ends where it should, whatever the flow: goods come from a vendor and end
// in stock, or start in stock and end at a customer.
func TestFlowPathsAreWellFormed(t *testing.T) {
	flows := []string{
		models.WarehouseFlowOneStep,
		models.WarehouseFlowTwoStep,
		models.WarehouseFlowThreeStep,
	}

	for _, flow := range flows {
		incoming := ResolveIncomingFlow(flow)
		assert.Equal(t, FlowEndpointVendor, incoming[0].FromCode, flow)
		assert.Equal(t, "Stock", incoming[len(incoming)-1].ToCode, flow)
		assertLegsConnect(t, incoming, flow)

		outgoing := ResolveOutgoingFlow(flow)
		assert.Equal(t, "Stock", outgoing[0].FromCode, flow)
		assert.Equal(t, FlowEndpointCustomer, outgoing[len(outgoing)-1].ToCode, flow)
		assertLegsConnect(t, outgoing, flow)
	}
}

// A gap between two legs would mean goods teleporting, so each leg must start where the last ended.
func assertLegsConnect(t *testing.T, legs []FlowLeg, flow string) {
	t.Helper()
	for i := 1; i < len(legs); i++ {
		assert.Equalf(t, legs[i-1].ToCode, legs[i].FromCode,
			"%s: leg %d starts somewhere leg %d did not end", flow, i, i-1)
	}
}

// A warehouse always has Stock; the rest follow from the flows. Three-step both ways is the
// example given in the requirement's location-hierarchy sample.
func TestRequiredSystemLocations(t *testing.T) {
	oneStep := requiredSystemLocations(models.WarehouseFlowOneStep, models.WarehouseFlowOneStep)
	assert.Equal(t, []systemLocationSpec{
		{Code: "Stock", Purpose: models.InventoryLocationPurposeStorage},
	}, oneStep, "a one-step warehouse needs nowhere to stop")

	full := requiredSystemLocations(models.WarehouseFlowThreeStep, models.WarehouseFlowThreeStep)
	codes := make([]string, 0, len(full))
	for _, spec := range full {
		codes = append(codes, spec.Code)
	}
	assert.Equal(t, []string{"Stock", "Input", "Quality Control", "Packing", "Output"}, codes)
}

// Reducing a flow leaves locations behind. They are reported so the caller can suspend them; the
// history that passed through them has to keep resolving, so they are never deleted.
func TestObsoleteSystemLocations(t *testing.T) {
	dropped := obsoleteSystemLocations(
		models.WarehouseFlowThreeStep, models.WarehouseFlowOneStep, false)
	assert.Equal(t, []string{"Input", "Quality Control"}, specCodes(dropped))

	narrowed := obsoleteSystemLocations(
		models.WarehouseFlowThreeStep, models.WarehouseFlowTwoStep, false)
	assert.Equal(t, []string{"Quality Control"}, specCodes(narrowed),
		"Input is still needed at two steps, so it is not obsolete")

	outgoing := obsoleteSystemLocations(
		models.WarehouseFlowThreeStep, models.WarehouseFlowTwoStep, true)
	assert.Equal(t, []string{"Packing"}, specCodes(outgoing))

	widened := obsoleteSystemLocations(
		models.WarehouseFlowOneStep, models.WarehouseFlowThreeStep, false)
	assert.Empty(t, widened, "adding steps makes nothing obsolete")
}

func specCodes(specs []systemLocationSpec) []string {
	codes := make([]string, 0, len(specs))
	for _, spec := range specs {
		codes = append(codes, spec.Code)
	}
	return codes
}

func TestIsKnownFlow(t *testing.T) {
	assert.True(t, isKnownFlow(models.WarehouseFlowOneStep))
	assert.True(t, isKnownFlow(models.WarehouseFlowThreeStep))
	assert.False(t, isKnownFlow("four_step"))
	assert.False(t, isKnownFlow(""))
}
