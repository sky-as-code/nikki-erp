package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"

	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
)

// NewPurchaseOrderDomainService derives the order service from the onion's default one. The
// lifecycle operations live only on this type; the order application service reaches them by
// type-asserting its domain service at construction.
func NewPurchaseOrderDomainService(
	base dynamicresource.CrudDomainService, references *OrderReferenceValidator,
	products *ProductLineValidator, pricer *LinePricer, numbering *CodeNumbering,
) *PurchaseOrderDomainServiceImpl {
	return &PurchaseOrderDomainServiceImpl{
		CrudDomainService: base,
		references:        references,
		products:          products,
		pricer:            pricer,
		numbering:         numbering,
	}
}

type PurchaseOrderDomainServiceImpl struct {
	dynamicresource.CrudDomainService

	// Validates the vendor and currency and defaults the currency from the vendor. Nil in tests
	// that exercise only the lifecycle rules.
	references *OrderReferenceValidator

	// What Reprice needs: products re-resolves each line's product for its template, pricer asks
	// the vendor's price list for the current number. Both nil in lifecycle tests, where repricing
	// reports it examined nothing rather than failing.
	products *ProductLineValidator
	pricer   *LinePricer

	// Issues the order code from the org's purchase.order sequence.
	numbering *CodeNumbering
}

// Create stamps the fields a client may not choose: a sequence-issued code, status forced to RFQ, zeroed
// totals, and is_locked, vendor_acknowledged and approval_required set false. Approval is decided
// at confirm time from the org configuration and the order total. Client values for these are
// overwritten rather than rejected, so echoing a record back does not fail.
func (this *PurchaseOrderDomainServiceImpl) Create(
	ctx corectx.Context, params dmodel.DynamicFields, options ...dynamicresource.CreateOptions,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	prepared := make(dmodel.DynamicFields, len(params)+7)
	for key, value := range params {
		prepared[key] = value
	}

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

	return createWithCode(ctx, this.CrudDomainService, this.numbering, SeqOrderCode, prepared, options...)
}
