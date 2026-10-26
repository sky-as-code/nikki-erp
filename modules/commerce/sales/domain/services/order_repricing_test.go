package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
)

// The pure parts of repricing: how stored lines become engine input. RepriceOrder itself touches the
// repository and is exercised live.

func lineRecord(id string, number int32, lineType, quantity, unitPrice string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.SalesOrderLineFieldId:              id,
		models.SalesOrderLineFieldLineNumber:      number,
		models.SalesOrderLineFieldLineType:        lineType,
		models.SalesOrderLineFieldOrderedQuantity: quantity,
		models.SalesOrderLineFieldBaseUnitPrice:   unitPrice,
	}
}

// toLineInputs mirrors what buildPricingInput does to each record, without the repository read.
func toLineInputs(t *testing.T, records []dmodel.DynamicFields) []pricing.LineInput {
	t.Helper()
	lines := make([]pricing.LineInput, 0, len(records))
	for _, record := range records {
		if stringOf(record, models.SalesOrderLineFieldLineType) ==
			string(models.SalesOrderLineTypePromotionReward) {
			continue
		}
		lines = append(lines, pricing.LineInput{
			Key:                stringOf(record, models.SalesOrderLineFieldId),
			LineNumber:         int32Of(record, models.SalesOrderLineFieldLineNumber),
			Quantity:           decimalOf(record, models.SalesOrderLineFieldOrderedQuantity),
			CatalogueUnitPrice: decimalOf(record, models.SalesOrderLineFieldBaseUnitPrice),
		})
	}
	return lines
}

// A giveaway line is an output of pricing; feeding one back would give away another free item on
// every reprice, compounding on each edit.
func TestGiveawayLinesAreNotFedBackIntoTheEngine(t *testing.T) {
	records := []dmodel.DynamicFields{
		lineRecord("L1", 1, string(models.SalesOrderLineTypeProduct), "2", "10000"),
		lineRecord("L2", 2, string(models.SalesOrderLineTypePromotionReward), "1", "0"),
		lineRecord("L3", 3, string(models.SalesOrderLineTypeProduct), "1", "5000"),
	}

	lines := toLineInputs(t, records)

	if len(lines) != 2 {
		t.Fatalf("the giveaway line must be excluded, got %d lines", len(lines))
	}
	for _, line := range lines {
		if line.Key == "L2" {
			t.Error("the promotion_reward line was fed back into the engine")
		}
	}
}

// A combo parent is an input — unlike a giveaway, the engine needs it to expand components.
func TestComboLinesAreFedIn(t *testing.T) {
	records := []dmodel.DynamicFields{
		lineRecord("L1", 1, string(models.SalesOrderLineTypeCombo), "1", "48000"),
	}

	if lines := toLineInputs(t, records); len(lines) != 1 {
		t.Fatalf("a combo parent must be priced, got %d lines", len(lines))
	}
}

// An order with zero lines is a valid draft that prices to nothing; it is confirming one that is not
// allowed.
func TestAnEmptyBasketPricesToZero(t *testing.T) {
	result := pricing.Calculate(pricing.Input{
		Lines:   nil,
		Context: pricing.Context{CurrencyScale: 0},
	})

	if !result.GrandTotal.IsZero() || len(result.Lines) != 0 {
		t.Errorf("an empty draft must price to zero, got total %s over %d lines",
			result.GrandTotal, len(result.Lines))
	}
}

// Quantity and price must survive the numeric shapes a jsonb round trip produces; read as zero,
// every line would silently price to nothing.
func TestQuantityAndPriceSurviveJsonShapes(t *testing.T) {
	for name, value := range map[string]any{
		"string":  "2",
		"float64": float64(2),
	} {
		t.Run(name, func(t *testing.T) {
			record := dmodel.DynamicFields{
				models.SalesOrderLineFieldId:              "L1",
				models.SalesOrderLineFieldOrderedQuantity: value,
				models.SalesOrderLineFieldBaseUnitPrice:   value,
			}
			lines := toLineInputs(t, []dmodel.DynamicFields{record})
			if !lines[0].Quantity.Equal(dec("2")) {
				t.Errorf("quantity from %s = %s, want 2", name, lines[0].Quantity)
			}
			if !lines[0].CatalogueUnitPrice.Equal(dec("2")) {
				t.Errorf("price from %s = %s, want 2", name, lines[0].CatalogueUnitPrice)
			}
		})
	}
}

// Repricing the same basket twice must produce the same numbers: if the ordering of the stored lines
// leaked into the answer, an untouched draft would drift every time it was read.
func TestRepricingAnUnchangedBasketIsStable(t *testing.T) {
	build := func() pricing.Input {
		return pricing.Input{
			Lines: []pricing.LineInput{
				{Key: "L1", LineNumber: 1, Quantity: dec("3"), CatalogueUnitPrice: dec("1000")},
				{Key: "L2", LineNumber: 2, Quantity: dec("2"), CatalogueUnitPrice: dec("1500")},
			},
			Context: pricing.Context{CurrencyScale: 0},
		}
	}

	first := pricing.Calculate(build())
	for attempt := 0; attempt < 5; attempt++ {
		again := pricing.Calculate(build())
		if !again.GrandTotal.Equal(first.GrandTotal) {
			t.Fatalf("reprice %d gave %s, first gave %s",
				attempt, again.GrandTotal, first.GrandTotal)
		}
	}
}

// The adjustment chain renumbers from one on every reprice, which is why the old chain is deleted
// rather than amended: sequence is unique per order.
func TestAdjustmentSequencesRestartFromOne(t *testing.T) {
	result := pricing.Calculate(pricing.Input{
		Lines: []pricing.LineInput{
			{Key: "L1", LineNumber: 1, Quantity: dec("1"), CatalogueUnitPrice: dec("10000")},
		},
		Programs: []pricing.AppliedProgram{{
			ProgramId: "P1", ProgramName: "Ten off",
			Rewards: []pricing.RewardInput{
				{RewardId: "R1", Sequence: 1, Type: "percentage_discount", Value: dec("10")},
			},
		}},
		Context: pricing.Context{CurrencyScale: 0},
	})

	if len(result.Adjustments) == 0 {
		t.Fatal("a discount must produce an adjustment")
	}
	if result.Adjustments[0].Sequence != 1 {
		t.Errorf("the chain must start at sequence 1, got %d", result.Adjustments[0].Sequence)
	}
}
