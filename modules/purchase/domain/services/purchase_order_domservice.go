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

// NewPurchaseOrderDomainService derives the order service from the engine's default one. The
// module must install it: the lifecycle operations live only on this type, and the action callbacks
// reach them by type-asserting the engine's service.
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

	// Validates the vendor and currency and defaults the currency from the vendor. Nil in tests
	// that exercise only the lifecycle rules.
	references *OrderReferenceValidator

	// What Reprice needs: products re-resolves each line's product for its template, pricer asks
	// the vendor's price list for the current number. Both nil in lifecycle tests, where repricing
	// reports it examined nothing rather than failing.
	products *ProductLineValidator
	pricer   *LinePricer
}

var _ drif.DynamicResourceService = (*PurchaseOrderDomainServiceImpl)(nil)

// Create stamps the fields a client may not choose: a generated code, status forced to RFQ, zeroed
// totals, and is_locked, vendor_acknowledged and approval_required set false. Approval is decided
// at confirm time from the org configuration and the order total. Client values for these are
// overwritten rather than rejected, so echoing a record back does not fail.
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

// generateOrderCode mints the order's human-facing reference. The prefix is "PO" whatever the
// status: an RFQ and the purchase order it becomes are one record, and a status-encoding code would
// have to change under a vendor already holding the old one. The suffix is a ULID rather than a
// per-org counter, being unique without coordination while still sorting by creation time.
func generateOrderCode() (string, error) {
	id, err := model.NewId()
	if err != nil {
		return "", errors.Wrap(err, "generateOrderCode")
	}
	return "PO-" + *id, nil
}
