package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// NewSalesVoucherApplicationService serves voucher administration (BR §9.2). A client never
// supplies source_order_id: a next-order coupon is issued only by its source order's
// confirmation, so a hand-written one would fake history.
func NewSalesVoucherApplicationService(base dynamicresource.CrudApplicationService) dynamicresource.CrudApplicationService {
	return &SalesVoucherApplicationServiceImpl{
		CrudApplicationService: base,
		voucherSvc:             base.DomainService().(itPromotion.SalesVoucherDomainService),
	}
}

type SalesVoucherApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
	voucherSvc itPromotion.SalesVoucherDomainService
}

var _ itPromotion.SalesVoucherApplicationService = (*SalesVoucherApplicationServiceImpl)(nil)

func (this *SalesVoucherApplicationServiceImpl) Create(
	ctx corectx.Context, cmd dynamicresource.CreateCommand,
) (*dynamicresource.CreateResult, error) {
	for _, field := range []string{
		models.SalesVoucherFieldSourceOrderId, models.SalesVoucherFieldVoucherType,
		models.SalesVoucherFieldUsedCount, models.SalesVoucherFieldIssuedAt,
	} {
		delete(cmd, field)
	}
	return this.CrudApplicationService.Create(ctx, cmd)
}

// GenerateBatch is collection-level: the org comes from the body and is checked like a create.
func (this *SalesVoucherApplicationServiceImpl) GenerateBatch(
	ctx corectx.Context, cmd itPromotion.GenerateVouchersCommand,
) (*itPromotion.GenerateVouchersResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionGenerateVouchers, cmd); cErrs != nil {
		return &itPromotion.GenerateVouchersResult{ClientErrors: *cErrs}, nil
	}
	delete(cmd, models.SalesVoucherFieldSourceOrderId)
	return this.voucherSvc.GenerateBatch(ctx, cmd)
}

func (this *SalesVoucherApplicationServiceImpl) Revoke(
	ctx corectx.Context, cmd itPromotion.LifecycleCommand,
) (*itPromotion.RecordResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionRevokeVoucher, cmd); cErrs != nil || err != nil {
		return recordFailure(cErrs, err)
	}
	return this.voucherSvc.Revoke(ctx, readStringParam(cmd, paramRecordId),
		readStringParam(cmd, itPromotion.FieldVoucherNote))
}
