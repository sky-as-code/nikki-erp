package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// NewPurchaseOrderDomainService derives the order service from the engine's default one.
//
// The lifecycle operations — confirm, approve, cancel and the rest — live on this type and nowhere
// else, which is why the module must install it: the action callbacks reach them by type-asserting
// the engine's service, and without this every one of them fails at the assertion.
func NewPurchaseOrderDomainService(
	base drif.DynamicResourceService, references *OrderReferenceValidator,
	products *ProductLineValidator, pricer *LinePricer,
) *PurchaseOrderDomainServiceImpl {
	return &PurchaseOrderDomainServiceImpl{
		DynamicResourceService: base,
		references:             references,
		products:               products,
		pricer:                 pricer,
	}
}

type PurchaseOrderDomainServiceImpl struct {
	drif.DynamicResourceService

	// references validates the vendor and currency and defaults the currency from the vendor. It
	// is nil in tests that exercise only the lifecycle rules, which need no ports.
	references *OrderReferenceValidator

	// products and pricer are what Reprice needs (section 30): the first re-resolves each line's
	// product to obtain its template, the second asks the vendor's price list for the current
	// number. Both nil in the same lifecycle tests, where repricing reports that it examined
	// nothing rather than failing.
	products *ProductLineValidator
	pricer   *LinePricer
}

var _ drif.DynamicResourceService = (*PurchaseOrderDomainServiceImpl)(nil)

// Create stamps the fields a client may not choose.
//
// Five things happen that a plain CRUD create would not:
//
//   - The code is generated. It identifies the order on paperwork and to the vendor, so letting a
//     client pick it would let two orders collide, or one impersonate another's reference.
//   - The status is forced to RFQ. An order that could be created `purchase_order` would be a
//     commitment with no confirmation behind it, and one created `cancelled` would be a document
//     that never existed. Every order starts as a request for quotation and earns the rest.
//   - The three totals are zeroed. They are a summary of lines that do not exist yet.
//   - is_locked and vendor_acknowledged start false: both are things that happen TO an order later,
//     and neither is true of one that has just been typed in.
//   - approval_required starts false. Whether approval is needed is decided at confirm time
//     against the org's configuration and the order's total, neither of which is knowable now.
//
// Client values for these are overwritten rather than rejected: they are not part of the request's
// meaning, and a client echoing a record back should not fail for carrying them.
func (this *PurchaseOrderDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	code, err := generateOrderCode()
	if err != nil {
		return nil, err
	}

	prepared := make(dmodel.DynamicFields, len(params)+7)
	for key, value := range params {
		prepared[key] = value
	}

	prepared[models.PurchaseOrderFieldCode] = code
	prepared[models.PurchaseOrderFieldStatus] = string(models.PurchaseOrderStatusRfq)
	prepared[models.PurchaseOrderFieldIsLocked] = false
	prepared[models.PurchaseOrderFieldVendorAcknowledged] = false
	prepared[models.PurchaseOrderFieldApprovalRequired] = false
	StampOrderTotalsForCreate(prepared)

	if this.references != nil {
		vErrs := ft.NewClientErrors()
		if err := this.references.PrepareOrder(ctx, prepared, vErrs); err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *vErrs}, nil
		}
	}

	return this.DynamicResourceService.Create(ctx, prepared)
}

// generateOrderCode mints the order's human-facing reference.
//
// The prefix is "PO" for every order regardless of status, and deliberately so: an RFQ and the
// purchase order it becomes are one record, so a code that encoded the status would have to change
// when the status did — and the vendor is holding a document quoting the old one. See PUR-R1.
//
// The suffix is a ULID rather than a per-org counter. A counter needs a sequence or a locked
// read-modify-write to stay gap-free under concurrency, and BR 25's numbering scheme is not in
// this phase; a ULID is unique without coordination and still sorts by creation time.
func generateOrderCode() (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "generateOrderCode")
	}
	return "PO-" + *id, nil
}
