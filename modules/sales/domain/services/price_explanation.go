package services

import (
	"sort"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The price explanation (BR 87.9).
//
// Given an order, answer "why does this line cost this" as the ordered chain of steps that produced
// the number:
//
//	Base 12,000 / Conditional promotion -2,000 / Final 10,000
//
// **It adds no storage and computes no prices.** Every figure is read back from
// sales_order_adjustments, which the pricing engine already wrote. That is the whole reason the
// engine returns a LIST of adjustments rather than a total (BR 13): a total alone could never be
// explained, because discounts do not commute and the same three of them in a different order give a
// different answer.
//
// The consequence worth stating: this reads history, it does not re-derive it. An explanation of a
// confirmed sale stays true even after the pricelist, the promotion or the tax rate that produced it
// have all changed — which is what an operator answering a customer complaint actually needs.

// PriceExplanation is the full account of one order's pricing.
type PriceExplanation struct {
	SalesOrderId string

	// Lines is one chain per order line, in line-number order.
	Lines []LinePriceExplanation

	// OrderSteps are the adjustments that belong to the order rather than to any one line - an
	// order-level voucher before it was allocated down, and the final monetary rounding. They are
	// kept separate rather than distributed across the lines because an operator asking "where did
	// this 1,234 go" needs to see the step as it was applied, not six fragments of it.
	OrderSteps []PriceStep

	Subtotal      decimal.Decimal
	DiscountTotal decimal.Decimal
	TaxTotal      decimal.Decimal
	GrandTotal    decimal.Decimal
}

// LinePriceExplanation is the chain for one line.
type LinePriceExplanation struct {
	SalesOrderLineId string
	LineNumber       int32

	ProductCode string
	ProductName string

	Quantity decimal.Decimal

	// BaseAmount is where the chain starts: the line's gross, before any adjustment.
	BaseAmount decimal.Decimal

	// Steps are the adjustments that touched this line, in the order they were applied.
	Steps []PriceStep

	// NetAmount, TaxAmount and FinalAmount are where it ends. They are read from the line rather
	// than recomputed from the steps, so that an explanation whose arithmetic disagrees with the
	// stored line is VISIBLE rather than silently papered over - see StepsReconcile.
	NetAmount   decimal.Decimal
	TaxAmount   decimal.Decimal
	FinalAmount decimal.Decimal
}

// PriceStep is one link in the chain.
type PriceStep struct {
	Sequence int32

	// Type is the adjustment type: combo_price, conditional_price, percentage_discount,
	// fixed_discount, voucher, manual_discount, rounding.
	Type string

	// SourceType and SourceId name what caused it - the promotion program, the combo, the pricelist.
	// Carried so a screen can link through to the campaign rather than only naming it.
	SourceType string
	SourceId   string

	Description string

	// BaseAmount is what this step was calculated FROM, which depends on where in the sequence it
	// fell: ten percent applied third operates on a different base than the same ten percent applied
	// first. Storing it is what makes the chain readable without replaying the engine.
	BaseAmount decimal.Decimal

	// Amount is SIGNED - negative for a discount, positive for a clawback or a rounding increase.
	Amount decimal.Decimal
}

// StepsReconcile reports whether a line's steps actually account for the difference between its base
// and its net.
//
// Worth having as an explicit answer rather than an assumption. If it is ever false, either an
// adjustment was written without its line, or a line was repriced without its adjustments being
// rewritten - both are real bugs, and both would otherwise show up as an explanation that quietly
// does not add up on a screen an operator is using to answer a customer.
func (this LinePriceExplanation) StepsReconcile() bool {
	running := this.BaseAmount
	for _, step := range this.Steps {
		running = running.Add(step.Amount)
	}
	return running.Equal(this.NetAmount)
}

// ExplainOrderPrice builds the explanation for one order.
//
// Returns nil when the order does not exist, rather than an error: a caller naming a record that is
// not there has made a mistake it can correct.
func ExplainOrderPrice(
	ctx corectx.Context, orderId string,
) (*PriceExplanation, error) {
	orderRecord, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil || orderRecord == nil {
		return nil, err
	}
	order := models.NewSalesOrderFrom(orderRecord)

	lineRecords, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}
	adjustmentRecords, err := searchBy(ctx,
		models.SalesOrderAdjustmentSchemaName,
		models.SalesOrderAdjustmentFieldSalesOrderId, orderId)
	if err != nil {
		return nil, err
	}

	stepsByLine, orderSteps := groupSteps(adjustmentRecords)

	explanation := &PriceExplanation{
		SalesOrderId:  orderId,
		OrderSteps:    orderSteps,
		Subtotal:      decimalOrZero(order.GetSubtotal()),
		DiscountTotal: decimalOrZero(order.GetDiscountTotal()),
		TaxTotal:      decimalOrZero(order.GetTaxTotal()),
		GrandTotal:    decimalOrZero(order.GetGrandTotal()),
	}

	for _, lineRecord := range lineRecords {
		lineId := stringOf(lineRecord, models.SalesOrderLineFieldId)
		explanation.Lines = append(explanation.Lines, LinePriceExplanation{
			SalesOrderLineId: lineId,
			LineNumber:       int32Of(lineRecord, models.SalesOrderLineFieldLineNumber),
			ProductCode:      stringOf(lineRecord, models.SalesOrderLineFieldProductCodeSnapshot),
			ProductName:      stringOf(lineRecord, models.SalesOrderLineFieldProductNameSnapshot),
			Quantity:         decimalOf(lineRecord, models.SalesOrderLineFieldOrderedQuantity),
			BaseAmount:       decimalOf(lineRecord, models.SalesOrderLineFieldGrossAmount),
			Steps:            stepsByLine[lineId],
			NetAmount:        decimalOf(lineRecord, models.SalesOrderLineFieldNetAmount),
			TaxAmount:        decimalOf(lineRecord, models.SalesOrderLineFieldTaxAmount),
			FinalAmount:      decimalOf(lineRecord, models.SalesOrderLineFieldFinalAmount),
		})
	}

	// Line number order, so the explanation reads down the receipt. The repository returns rows in
	// whatever order it likes, and an explanation whose lines shuffle between two reads of the same
	// unchanged order is one nobody can check against a printed receipt.
	sort.SliceStable(explanation.Lines, func(i, j int) bool {
		return explanation.Lines[i].LineNumber < explanation.Lines[j].LineNumber
	})
	return explanation, nil
}

// groupSteps splits the adjustments into per-line chains and the order-level chain, each ordered by
// sequence.
//
// Sequence is what makes the chain replayable and is unique per order, so ordering by it gives one
// total order across the whole document - a line's steps keep their relative positions even though
// they are shown separately.
func groupSteps(
	adjustmentRecords []dmodel.DynamicFields,
) (byLine map[string][]PriceStep, orderSteps []PriceStep) {
	byLine = make(map[string][]PriceStep, len(adjustmentRecords))

	for _, record := range adjustmentRecords {
		step := PriceStep{
			Sequence:    int32Of(record, models.SalesOrderAdjustmentFieldSequence),
			Type:        stringOf(record, models.SalesOrderAdjustmentFieldAdjustmentType),
			SourceType:  stringOf(record, models.SalesOrderAdjustmentFieldSourceType),
			SourceId:    stringOf(record, models.SalesOrderAdjustmentFieldSourceId),
			Description: stringOf(record, models.SalesOrderAdjustmentFieldDescription),
			BaseAmount:  decimalOf(record, models.SalesOrderAdjustmentFieldBaseAmount),
			Amount:      decimalOf(record, models.SalesOrderAdjustmentFieldAdjustmentAmount),
		}

		lineId := stringOf(record, models.SalesOrderAdjustmentFieldSalesOrderLineId)
		if lineId == "" {
			orderSteps = append(orderSteps, step)
			continue
		}
		byLine[lineId] = append(byLine[lineId], step)
	}

	for lineId := range byLine {
		sortSteps(byLine[lineId])
	}
	sortSteps(orderSteps)
	return byLine, orderSteps
}

func sortSteps(steps []PriceStep) {
	sort.SliceStable(steps, func(i, j int) bool {
		return steps[i].Sequence < steps[j].Sequence
	})
}

// decimalOrZero reads a nullable decimal accessor, treating absence as zero.
func decimalOrZero(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}
