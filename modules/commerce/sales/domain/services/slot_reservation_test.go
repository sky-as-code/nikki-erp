package services

import (
	"testing"

	"github.com/shopspring/decimal"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The guard of slot-level reservation: stock for a vending sale is held at the individual slot, and
// several slots in one sale means several holds. What these protect is the ability to consume the
// slots that dispensed and release only the one that jammed.

const (
	targetLocation = "01TARGETLOCATION0000000000"
	slotA01        = "01SLOTLOCATIONA010000000AA"
	slotB01        = "01SLOTLOCATIONB010000000BB"
)

func linesAt(locationIds ...string) []itExt.FulfillmentLine {
	lines := make([]itExt.FulfillmentLine, 0, len(locationIds))
	for index, locationId := range locationIds {
		lines = append(lines, itExt.FulfillmentLine{
			ProductVariantId: "variant-" + string(rune('A'+index)),
			Quantity:         decimal.NewFromInt(1),
			SourceLocationId: locationId,
		})
	}
	return lines
}

// A line naming no location is every non-vending sale there is. It must still produce exactly one
// hold, at the fulfillment's own target — otherwise this change breaks warehouse and eCommerce.
func TestLinesWithoutASlotBecomeOneHoldAtTheTarget(t *testing.T) {
	groups := groupItemsByLocation(targetLocation, linesAt("", ""), []string{"item-1", "item-2"})

	if len(groups) != 1 {
		t.Fatalf("expected one group for an unslotted target, got %d", len(groups))
	}
	if groups[0].LocationId != targetLocation {
		t.Fatalf("expected the group at the target location, got %q", groups[0].LocationId)
	}
	if len(groups[0].Items) != 2 {
		t.Fatalf("expected both items in the one group, got %d", len(groups[0].Items))
	}
}

// Two slots are two holds. One call could not express them: Inventory records one location per
// reservation, so a single hold would silently be taken at whichever slot was named first.
func TestItemsAtDifferentSlotsBecomeSeparateHolds(t *testing.T) {
	groups := groupItemsByLocation(targetLocation,
		linesAt(slotA01, slotB01), []string{"item-1", "item-2"})

	if len(groups) != 2 {
		t.Fatalf("expected one group per slot, got %d", len(groups))
	}
	if groups[0].LocationId != slotA01 || groups[1].LocationId != slotB01 {
		t.Fatalf("groups lost the order the lines were given: %q then %q",
			groups[0].LocationId, groups[1].LocationId)
	}
	for _, group := range groups {
		if len(group.Items) != 1 || len(group.ItemIds) != 1 {
			t.Fatalf("a slot holding one line reported %d items", len(group.Items))
		}
	}
}

// Two lines from the SAME slot share one hold. Splitting them would take two reservations against
// one quant and make the second look like a replay of the first.
func TestItemsAtTheSameSlotShareOneHold(t *testing.T) {
	groups := groupItemsByLocation(targetLocation,
		linesAt(slotA01, slotA01), []string{"item-1", "item-2"})

	if len(groups) != 1 {
		t.Fatalf("expected one group for one slot, got %d", len(groups))
	}
	if len(groups[0].Items) != 2 {
		t.Fatalf("expected both lines in the slot's group, got %d", len(groups[0].Items))
	}
}

// A hold at the fulfillment's own target keeps the bare id, so nothing about the holds already taken
// by the existing non-slotted flows changes.
func TestTheTargetLocationKeepsTheBareSourceId(t *testing.T) {
	if got := reservationSourceId("ful-1", targetLocation, targetLocation); got != "ful-1" {
		t.Fatalf("expected the bare fulfillment id at the target, got %q", got)
	}
	if got := reservationSourceId("ful-1", "", targetLocation); got != "ful-1" {
		t.Fatalf("expected the bare fulfillment id for an unslotted line, got %q", got)
	}
}

// Each slot gets a source id of its own. Reusing the bare id across slots would make Inventory treat
// the second slot as an idempotent replay of the first and hold nothing for it.
func TestEachSlotGetsItsOwnSourceId(t *testing.T) {
	first := reservationSourceId("ful-1", slotA01, targetLocation)
	second := reservationSourceId("ful-1", slotB01, targetLocation)

	if first == second {
		t.Fatal("two slots produced the same source id, so one hold would replay over the other")
	}
	if first != "ful-1"+reservationSourceSeparator+slotA01 {
		t.Fatalf("unexpected slot source id %q", first)
	}
}
