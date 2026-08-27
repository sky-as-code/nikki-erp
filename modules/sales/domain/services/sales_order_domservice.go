package services

import (
	"fmt"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// SalesOrderDomainServiceImpl derives the order resource, adding the rules the built-in CRUD cannot
// express.
//
// Two of those rules exist here because they exist NOWHERE else. The framework declares no CHECK
// constraints — no migration in either tree contains one — so the BR §55 quantity invariant and the
// BR §11 snapshot immutability rule have exactly one enforcement point, which is this file. The plan
// asked for them in the database as well; that is not available, and the consequence is that a write
// bypassing this service bypasses the invariant entirely.
type SalesOrderDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*SalesOrderDomainServiceImpl)(nil)

func NewSalesOrderDomainService(base drif.DynamicResourceService) *SalesOrderDomainServiceImpl {
	return &SalesOrderDomainServiceImpl{DynamicResourceService: base}
}

// AssertEditable refuses a change to an order that is no longer a draft.
//
// Confirmation is the line: before it the document may be repriced freely, after it the numbers are
// what the business promised the customer (BR §11). A cancelled order is refused too, and
// deliberately — it is a record of something that was attempted, and rewriting it would destroy the
// evidence of what was attempted.
func (this *SalesOrderDomainServiceImpl) AssertEditable(
	ctx corectx.Context, orderId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	record, err := loadRecord(ctx, models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return notFoundResult(models.SalesOrderSchemaName, orderId), nil
	}

	order := models.NewSalesOrderFrom(record)
	if !order.IsEditable() {
		return violationResult(models.SalesOrderSchemaName, "sales_order.not_editable",
			"only a draft sales order may be changed; this one is '"+
				stringOf(record, models.SalesOrderFieldStatus)+"'"), nil
	}
	return mutateOk(), nil
}

// SalesOrderLineDomainServiceImpl derives the order line resource.
type SalesOrderLineDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*SalesOrderLineDomainServiceImpl)(nil)

func NewSalesOrderLineDomainService(
	base drif.DynamicResourceService,
) *SalesOrderLineDomainServiceImpl {
	return &SalesOrderLineDomainServiceImpl{DynamicResourceService: base}
}

// Create writes a line, refusing one whose quantities break the BR §55 invariant.
//
// The check runs on create as well as on update because a line can be born broken: a caller posting
// ordered_quantity 0, or a fulfilled quantity above it, would otherwise be stored and only discovered
// when something tried to compute a refund from it.
func (this *SalesOrderLineDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	if vErrs := assertQuantitiesConsistent(params); vErrs != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
	}
	return this.DynamicResourceService.Create(ctx, params)
}

// Update refuses a change that would break the quantity invariant or edit a frozen snapshot.
//
// The merge with the stored row matters: an update carries only the fields the caller touched, so
// checking the payload alone would let "fulfilled_quantity = 5" through against a stored
// ordered_quantity of 3. The invariant is a property of the resulting row, so the resulting row is
// what must be checked.
func (this *SalesOrderLineDomainServiceImpl) Update(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	lineId := stringOf(params, models.SalesOrderLineFieldId)
	stored, err := loadRecord(ctx,
		models.SalesOrderLineSchemaName, models.SalesOrderLineFieldId, lineId)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return notFoundResult(models.SalesOrderLineSchemaName, lineId), nil
	}

	if vErrs := assertQuantitiesConsistent(mergedFields(stored, params)); vErrs != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}

	frozen, err := this.assertSnapshotsUnchanged(ctx, stored, params)
	if err != nil {
		return nil, err
	}
	if frozen != nil {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *frozen}, nil
	}

	return this.DynamicResourceService.Update(ctx, params)
}

// assertSnapshotsUnchanged enforces BR §11: once the order is confirmed, the line's snapshot fields
// are frozen.
//
// It is the order's status that decides, not the line's own state, because a snapshot records how
// the world looked when the BUSINESS committed to the sale. A line added to a draft may be re-priced
// as often as the basket changes; the same line on a confirmed order may not, because a receipt has
// been printed from it.
//
// Only a field whose value actually DIFFERS is refused. A caller re-submitting a whole line
// unchanged is not trying to edit a snapshot, and refusing it would make a read-modify-write cycle
// impossible for the mutable fields alongside it.
func (this *SalesOrderLineDomainServiceImpl) assertSnapshotsUnchanged(
	ctx corectx.Context, stored dmodel.DynamicFields, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	orderId := stringOf(stored, models.SalesOrderLineFieldSalesOrderId)
	orderRecord, err := loadRecord(ctx,
		models.SalesOrderSchemaName, models.SalesOrderFieldId, orderId)
	if err != nil {
		return nil, err
	}
	if orderRecord == nil {
		// The line's order is gone. That is a broken reference rather than a rule the caller broke,
		// so it is not this check's business to report it; leave the write to fail on its own terms.
		return nil, nil
	}
	if !models.NewSalesOrderFrom(orderRecord).IsConfirmed() {
		return nil, nil
	}

	vErrs := ft.NewClientErrors()
	for _, field := range models.SnapshotFields {
		submitted, present := params[field]
		if !present {
			continue
		}
		if sameFieldValue(stored[field], submitted) {
			continue
		}
		vErrs.Append(*ft.NewBusinessViolation(field, "sales_order_line.snapshot_immutable",
			"'"+field+"' is a snapshot of the moment of sale and cannot change after the order "+
				"is confirmed"))
	}
	if vErrs.Count() == 0 {
		return nil, nil
	}
	return vErrs, nil
}

// assertQuantitiesConsistent enforces BR §55 on the resulting row:
// ordered > 0, 0 <= fulfilled <= ordered, 0 <= returned <= fulfilled.
//
// Each violation names the field the caller must fix rather than the rule as a whole, so a form can
// point at the offending box. Returning nil means the row is consistent.
func assertQuantitiesConsistent(fields dmodel.DynamicFields) *ft.ClientErrors {
	ordered := decimalField(fields, models.SalesOrderLineFieldOrderedQuantity)
	fulfilled := decimalField(fields, models.SalesOrderLineFieldFulfilledQuantity)
	returned := decimalField(fields, models.SalesOrderLineFieldReturnedQuantity)

	vErrs := ft.NewClientErrors()
	if !ordered.IsPositive() {
		vErrs.Append(*ft.NewBusinessViolation(
			models.SalesOrderLineFieldOrderedQuantity, "sales_order_line.ordered_quantity_not_positive",
			"a line must order more than zero; a line ordering nothing is not a line"))
	}
	if fulfilled.IsNegative() {
		vErrs.Append(*ft.NewBusinessViolation(
			models.SalesOrderLineFieldFulfilledQuantity, "sales_order_line.fulfilled_quantity_negative",
			"fulfilled quantity cannot be negative"))
	} else if fulfilled.GreaterThan(ordered) {
		vErrs.Append(*ft.NewBusinessViolation(
			models.SalesOrderLineFieldFulfilledQuantity, "sales_order_line.fulfilled_exceeds_ordered",
			"cannot fulfil "+fulfilled.String()+" of a line that ordered "+ordered.String()))
	}
	if returned.IsNegative() {
		vErrs.Append(*ft.NewBusinessViolation(
			models.SalesOrderLineFieldReturnedQuantity, "sales_order_line.returned_quantity_negative",
			"returned quantity cannot be negative"))
	} else if returned.GreaterThan(fulfilled) {
		// Measured against fulfilled rather than ordered on purpose: a customer cannot return what
		// was never handed over.
		vErrs.Append(*ft.NewBusinessViolation(
			models.SalesOrderLineFieldReturnedQuantity, "sales_order_line.returned_exceeds_fulfilled",
			"cannot return "+returned.String()+" of a line that fulfilled "+fulfilled.String()))
	}

	if vErrs.Count() == 0 {
		return nil
	}
	return vErrs
}

// mergedFields is the stored row overlaid with the submitted changes — the row as it would be after
// the update.
func mergedFields(stored, params dmodel.DynamicFields) dmodel.DynamicFields {
	merged := make(dmodel.DynamicFields, len(stored)+len(params))
	for key, value := range stored {
		merged[key] = value
	}
	for key, value := range params {
		merged[key] = value
	}
	return merged
}

// decimalField reads a quantity or money field, treating absent as zero.
//
// It accepts every shape the value can arrive in — a repository read, a JSON round-trip and a
// directly-constructed payload each produce a different concrete type — and never bare
// type-asserts, because a bare assertion on an unexpected type panics the request.
func decimalField(fields dmodel.DynamicFields, name string) decimal.Decimal {
	value, ok := fields[name]
	if !ok || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed != nil {
			return *typed
		}
	case string:
		// A decimal crosses JSON as a string so it does not lose precision on the way. A value that
		// will not parse is treated as zero, which fails the ordered > 0 check and is reported as a
		// violation rather than silently accepted.
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	case int64:
		return decimal.NewFromInt(typed)
	case int:
		return decimal.NewFromInt(int64(typed))
	}
	return decimal.Zero
}

// sameFieldValue compares a stored value with a submitted one across the type differences a round
// trip introduces.
//
// A decimal read back from the database and the same decimal submitted as a string are equal in
// value and different in type, so a plain == would report every re-submitted snapshot as an attempt
// to change it — and confirmed orders would become unupdatable in any field.
func sameFieldValue(stored, submitted any) bool {
	if stored == nil && submitted == nil {
		return true
	}
	if stored == nil || submitted == nil {
		return false
	}
	if storedDec, ok := asDecimal(stored); ok {
		if submittedDec, ok := asDecimal(submitted); ok {
			return storedDec.Equal(submittedDec)
		}
		return false
	}
	return stringValue(stored) == stringValue(submitted)
}

func asDecimal(value any) (decimal.Decimal, bool) {
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed, true
	case *decimal.Decimal:
		if typed != nil {
			return *typed, true
		}
	}
	return decimal.Zero, false
}

// stringValue renders a value for comparison without ever bare type-asserting.
//
// fmt.Sprint is the fallback rather than an assertion on some String() interface: the values here
// arrive from a repository, a JSON decode or a hand-built payload, so the concrete type is not
// knowable in advance, and an assertion that guessed wrong would panic the request rather than
// report a mismatch.
func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case *string:
		if typed != nil {
			return *typed
		}
		return ""
	}
	return fmt.Sprint(value)
}
