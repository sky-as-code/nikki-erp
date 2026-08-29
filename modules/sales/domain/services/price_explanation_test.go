package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// Grouping and ordering the adjustment chain, which is where an ordering bug would hide.
// ExplainOrderPrice itself reads the repository and is exercised live.

func adjustment(sequence int32, lineId, adjustmentType, amount string) dmodel.DynamicFields {
	record := dmodel.DynamicFields{
		models.SalesOrderAdjustmentFieldSequence:         sequence,
		models.SalesOrderAdjustmentFieldAdjustmentType:   adjustmentType,
		models.SalesOrderAdjustmentFieldAdjustmentAmount: amount,
	}
	if lineId != "" {
		record[models.SalesOrderAdjustmentFieldSalesOrderLineId] = lineId
	}
	return record
}

// The chain must come back in sequence order whatever order the rows arrived in: discounts do not
// commute, so a chain shown out of order describes a different calculation than the one that ran.
func TestStepsAreOrderedBySequenceNotByRowOrder(t *testing.T) {
	records := []dmodel.DynamicFields{
		adjustment(3, "L1", "voucher", "-500"),
		adjustment(1, "L1", "conditional_price", "-2000"),
		adjustment(2, "L1", "percentage_discount", "-300"),
	}

	byLine, _ := groupSteps(records)

	steps := byLine["L1"]
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	for index, want := range []int32{1, 2, 3} {
		if steps[index].Sequence != want {
			t.Errorf("step %d has sequence %d, want %d", index, steps[index].Sequence, want)
		}
	}
}

// An adjustment with no line belongs to the order; distributing those across the lines would show an
// operator six fragments of one step.
func TestOrderLevelStepsAreSeparatedFromLineSteps(t *testing.T) {
	records := []dmodel.DynamicFields{
		adjustment(1, "L1", "conditional_price", "-2000"),
		adjustment(2, "", "rounding", "-3"),
		adjustment(3, "L2", "voucher", "-500"),
	}

	byLine, orderSteps := groupSteps(records)

	if len(orderSteps) != 1 || orderSteps[0].Type != "rounding" {
		t.Fatalf("the line-less adjustment must be an order step, got %+v", orderSteps)
	}
	if len(byLine["L1"]) != 1 || len(byLine["L2"]) != 1 {
		t.Errorf("each line must keep its own step, got %+v", byLine)
	}
	if _, present := byLine[""]; present {
		t.Error("an order-level step must not be filed under an empty line id")
	}
}

// Order-level steps are ordered among themselves too.
func TestOrderLevelStepsAreAlsoOrdered(t *testing.T) {
	records := []dmodel.DynamicFields{
		adjustment(9, "", "rounding", "-3"),
		adjustment(4, "", "voucher", "-1000"),
	}

	_, orderSteps := groupSteps(records)

	if len(orderSteps) != 2 || orderSteps[0].Sequence != 4 || orderSteps[1].Sequence != 9 {
		t.Fatalf("order steps must be sequence-ordered, got %+v", orderSteps)
	}
}

// A sequence arriving through jsonb is a float64; read as anything else it becomes zero and every
// step sorts equal, destroying the ordering the explanation depends on.
func TestSequenceIsReadFromEveryNumericShape(t *testing.T) {
	for name, value := range map[string]any{
		"int32":   int32(7),
		"int64":   int64(7),
		"int":     7,
		"float64": float64(7),
	} {
		t.Run(name, func(t *testing.T) {
			byLine, _ := groupSteps([]dmodel.DynamicFields{{
				models.SalesOrderAdjustmentFieldSalesOrderLineId: "L1",
				models.SalesOrderAdjustmentFieldSequence:         value,
			}})
			if got := byLine["L1"][0].Sequence; got != 7 {
				t.Errorf("sequence from %s = %d, want 7", name, got)
			}
		})
	}
}

// Reading an amount unsigned would turn every discount into a surcharge.
func TestStepAmountsKeepTheirSign(t *testing.T) {
	byLine, _ := groupSteps([]dmodel.DynamicFields{
		adjustment(1, "L1", "percentage_discount", "-2000"),
		adjustment(2, "L1", "rounding", "3"),
	})

	steps := byLine["L1"]
	if !steps[0].Amount.IsNegative() {
		t.Errorf("a discount must stay negative, got %s", steps[0].Amount)
	}
	if !steps[1].Amount.IsPositive() {
		t.Errorf("a positive rounding must stay positive, got %s", steps[1].Amount)
	}
}

// Base plus every step must equal net.
func TestStepsReconcileWhenTheChainAddsUp(t *testing.T) {
	line := LinePriceExplanation{
		BaseAmount: dec("12000"),
		Steps: []PriceStep{
			{Sequence: 1, Amount: dec("-2000")},
		},
		NetAmount: dec("10000"),
	}

	if !line.StepsReconcile() {
		t.Error("12,000 less 2,000 is 10,000; the chain must reconcile")
	}
}

// And it must fail when it does not add up, or a customer is shown a discount nobody can account for.
func TestStepsDoNotReconcileWhenAStepIsMissing(t *testing.T) {
	line := LinePriceExplanation{
		BaseAmount: dec("12000"),
		Steps:      []PriceStep{{Sequence: 1, Amount: dec("-2000")}},
		NetAmount:  dec("9000"),
	}

	if line.StepsReconcile() {
		t.Error("12,000 less 2,000 is not 9,000; the mismatch must be reported")
	}
}

// A line with no adjustments reconciles trivially: its net is its base.
func TestALineWithNoStepsReconciles(t *testing.T) {
	line := LinePriceExplanation{
		BaseAmount: dec("5000"),
		NetAmount:  dec("5000"),
	}

	if !line.StepsReconcile() {
		t.Error("an unadjusted line must reconcile: its net is its base")
	}
}

// The worked example, end to end: Base 12,000 / Conditional −2,000 / Final 10,000.
func TestTheWorkedExampleFromTheRequirement(t *testing.T) {
	byLine, _ := groupSteps([]dmodel.DynamicFields{
		adjustment(1, "L1", "conditional_price", "-2000"),
	})

	line := LinePriceExplanation{
		BaseAmount: dec("12000"),
		Steps:      byLine["L1"],
		NetAmount:  dec("10000"),
	}

	if !line.StepsReconcile() {
		t.Fatal("the requirement's own example must reconcile")
	}
	if line.Steps[0].Type != "conditional_price" {
		t.Errorf("step type = %q, want conditional_price", line.Steps[0].Type)
	}
}
