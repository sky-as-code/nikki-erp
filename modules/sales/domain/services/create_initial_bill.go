package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The initial bill: the settlement unit a confirmation produces.
//
// Until it exists an order is a sale nobody can pay - a kiosk confirms, then has nothing to take
// money against. It allocates the WHOLE order, so at the moment of confirmation the order has
// exactly one bill worth exactly the order. A split or a merge afterwards redistributes that value
// between bills; neither changes what confirmation produced, which is why the order keeps pointing
// at this bill through initial_bill_id even once it has been superseded.
//
// Called inside the confirm transaction and never outside one: an order that says `confirmed` with
// no bill is the state CR §5.2 exists to make impossible.

// CreateInitialBill writes the bill and its line allocations for a freshly confirmed order, and
// answers the new bill's id.
func CreateInitialBill(
	ctx corectx.Context, order dmodel.DynamicFields, policy SalesPolicy,
) (string, error) {
	orderId := stringOf(order, models.SalesOrderFieldId)

	lines, err := searchBy(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldSalesOrderId, orderId)
	if err != nil {
		return "", err
	}
	if len(lines) == 0 {
		// assertConfirmable already refused this; reaching here means the lines vanished between that
		// check and this write, and billing nothing would produce a bill for an empty sale.
		return "", errors.New("order '" + orderId + "' has no lines to bill")
	}

	billId, err := model.NewId()
	if err != nil {
		return "", err
	}

	orgId := stringOf(order, basemodel.FieldOrgId)
	allocations, totals := initialAllocations(order, lines, policy)

	err = insertBill(ctx, string(*billId), orderId, orgId,
		stringOf(order, models.SalesOrderFieldCurrencyCode),
		initialBillNumberOf(string(*billId)), totals)
	if err != nil {
		return "", err
	}
	if err := insertBillAllocations(ctx, string(*billId), orgId, allocations); err != nil {
		return "", err
	}
	return string(*billId), nil
}

// ConfirmedOrderView is the order and the bill a confirmation produced, each as the whole record
// rather than an id. A till showing what the customer owes, and a kiosk about to ask for money,
// both need the figures in the same answer that told them the sale is on.
//
// The records are the same shape the read API returns, deliberately: the module has no hand-written
// detail DTO, and inventing one here would be a second definition of an order to keep in step.
type ConfirmedOrderView struct {
	Order dmodel.DynamicFields
	Bill  dmodel.DynamicFields
}

// LoadConfirmedOrderView reads back the order and its initial bill after a confirmation.
//
// Read after the transaction commits, not inside it: these are for the caller to look at, and
// reading them within the transaction would lengthen the work under the confirm lock for data the
// business rules do not need.
func LoadConfirmedOrderView(
	ctx corectx.Context, orderId, billId string,
) (*ConfirmedOrderView, error) {
	order, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}

	view := &ConfirmedOrderView{Order: order}
	if billId == "" {
		return view, nil
	}

	bill, err := loadRecord(ctx, models.SalesBillSchemaName, models.SalesBillFieldId, billId)
	if err != nil {
		return nil, err
	}
	view.Bill = bill
	return view, nil
}

// initialBillNumberOf names the first bill after its own id, the convention create_order uses for
// the order number. It is unique because the id is, needs no counter nobody would own, and still
// lets billNumberOf derive BILL-<id>-1 when the bill is later split.
func initialBillNumberOf(billId string) string {
	return "BILL-" + billId
}

// initialAllocations calculates the lines and totals before the parent bill is inserted.
//
// The three money columns are ALLOCATED, not copied. Pricing rounds the grand total once, as a
// whole (pricing/engine.go step 8), so the line amounts can sum to a hundredth either side of it;
// copying them would leave the bill disagreeing with the order it settles and fail the balance
// check. Allocate spreads each order total across the lines in proportion to what they already
// hold and gives the residual to one line deterministically, so the allocations sum to the order
// EXACTLY.
func initialAllocations(
	order dmodel.DynamicFields,
	lines []dmodel.DynamicFields,
	policy SalesPolicy,
) ([]billAllocation, *billTotals) {
	inputs := make([]AllocationInput, 0, len(lines))
	for index, line := range lines {
		inputs = append(inputs, AllocationInput{
			Key:       stringOf(line, models.SalesOrderLineFieldId),
			Reference: decimalOf(line, models.SalesOrderLineFieldFinalAmount),
			Tiebreak:  int32(index),
		})
	}

	net := AllocateAcrossBills(
		decimalOf(order, models.SalesOrderFieldSubtotal), inputs, policy.RoundingScale)
	tax := AllocateAcrossBills(
		decimalOf(order, models.SalesOrderFieldTaxTotal), inputs, policy.RoundingScale)
	total := AllocateAcrossBills(
		decimalOf(order, models.SalesOrderFieldGrandTotal), inputs, policy.RoundingScale)

	allocations := make([]billAllocation, 0, len(lines))
	totals := &billTotals{}
	for _, line := range lines {
		lineId := stringOf(line, models.SalesOrderLineFieldId)

		// Every line gets a row, including one allocated zero: the initial bill is the statement of
		// what was sold, and a free item omitted from it would make the bill disagree with the order
		// about what the customer bought.
		allocations = append(allocations, billAllocation{
			lineId: lineId, quantity: decimalOf(line, models.SalesOrderLineFieldOrderedQuantity),
			net: net[lineId], tax: tax[lineId], total: total[lineId],
		})
		totals.add(net[lineId], tax[lineId], total[lineId])
	}
	return allocations, totals
}
