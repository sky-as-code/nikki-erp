package services

import (
	"sort"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The price explanation: why a line costs what it does, as the ordered chain of steps that produced
// the number.
//
// It adds no storage and computes no prices — every figure is read back from
// sales_order_adjustments, which the pricing engine already wrote. This reads history rather than
// re-deriving it, so an explanation of a confirmed sale stays true after the pricelist, promotion or
// tax rate that produced it have changed.

// PriceExplanation is the full account of one order's pricing.
type PriceExplanation struct {
	SalesOrderId string

	// Lines is one chain per order line, in line-number order.
	Lines []LinePriceExplanation

	// OrderSteps are the adjustments that belong to the order rather than to any one line — an
	// order-level voucher before allocation, and the final monetary rounding. Kept separate so an
	// operator sees the step as it was applied, not six fragments of it.
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

	// BaseAmount is the line's gross, before any adjustment.
	BaseAmount decimal.Decimal

	// Steps are the adjustments that touched this line, in the order they were applied.
	Steps []PriceStep

	// Read from the line rather than recomputed from the steps, so an explanation whose arithmetic
	// disagrees with the stored line is visible — see StepsReconcile.
	NetAmount   decimal.Decimal
	TaxAmount   decimal.Decimal
	FinalAmount decimal.Decimal
}

// PriceStep is one link in the chain.
type PriceStep struct {
	Sequence int32

	// Type is one of combo_price, conditional_price, percentage_discount, fixed_discount, voucher,
	// manual_discount, rounding.
	Type string

	// SourceType and SourceId name what caused it, so a screen can link through to the campaign.
	SourceType string
	SourceId   string

	Description string

	// BaseAmount is what this step was calculated from, which depends on where in the sequence it
	// fell; storing it makes the chain readable without replaying the engine.
	BaseAmount decimal.Decimal

	// Amount is signed: negative for a discount, positive for a clawback or a rounding increase.
	Amount decimal.Decimal
}

// StepsReconcile reports whether a line's steps account for the difference between its base and its
// net. False means an adjustment was written without its line, or a line was repriced without its
// adjustments being rewritten — both real bugs that would otherwise surface as an explanation that
// quietly does not add up.
func (this LinePriceExplanation) StepsReconcile() bool {
	running := this.BaseAmount
	for _, step := range this.Steps {
		running = running.Add(step.Amount)
	}
	return running.Equal(this.NetAmount)
}

// ExplainOrderPrice builds the explanation for one order, returning nil when the order does not
// exist rather than an error.
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

	// Line number order, so the explanation reads down the receipt: the repository returns rows in
	// whatever order it likes.
	sort.SliceStable(explanation.Lines, func(i, j int) bool {
		return explanation.Lines[i].LineNumber < explanation.Lines[j].LineNumber
	})
	return explanation, nil
}

// groupSteps splits the adjustments into per-line chains and the order-level chain, each ordered by
// sequence. Sequence is unique per order, so it gives one total order across the whole document and
// a line's steps keep their relative positions even when shown separately.
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
