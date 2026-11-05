package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// A read-only promotion resource: a selector, written only through its owner's manage_* action, or
// a record of what a confirmed order received (application, allocation, issuance, attached voucher,
// reward choice), written only by the order flow (BR-L04: history has no archive or delete).
//
// Refused here, at the application layer, rather than through the onion's CrudActions: that list
// also binds the domain service, and the flows that legitimately write these rows reach it.

func NewSalesPromotionReadOnlyApplicationService(
	base dynamicresource.CrudApplicationService,
) itPromotion.SalesPromotionReadOnlyApplicationService {
	return &SalesPromotionReadOnlyApplicationServiceImpl{CrudApplicationService: base}
}

type SalesPromotionReadOnlyApplicationServiceImpl struct {
	dynamicresource.CrudApplicationService
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) refusal(action dynamicresource.CrudAction) ft.ClientErrors {
	return ft.ClientErrors{*ft.NewAnonymousBusinessViolation(ft.ErrorKey("err_action_not_supported"),
		"This resource does not support the requested action",
		map[string]any{"action": string(action), "resource": this.ResourceCode()})}
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) Create(
	_ corectx.Context, _ dynamicresource.CreateCommand,
) (*dynamicresource.CreateResult, error) {
	return &dynamicresource.CreateResult{ClientErrors: this.refusal(dynamicresource.CrudActionCreate)}, nil
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) Update(
	_ corectx.Context, _ dynamicresource.UpdateCommand,
) (*dynamicresource.MutateResult, error) {
	return &dynamicresource.MutateResult{ClientErrors: this.refusal(dynamicresource.CrudActionUpdate)}, nil
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) Delete(
	_ corectx.Context, _ dynamicresource.DeleteCommand,
) (*dynamicresource.MutateResult, error) {
	return &dynamicresource.MutateResult{ClientErrors: this.refusal(dynamicresource.CrudActionDelete)}, nil
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) SetArchived(
	_ corectx.Context, _ dynamicresource.SetArchivedCommand,
) (*dynamicresource.MutateResult, error) {
	return &dynamicresource.MutateResult{ClientErrors: this.refusal(dynamicresource.CrudActionSetArchived)}, nil
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) CreateBulk(
	_ corectx.Context, _ dynamicresource.BulkCreateCommand,
) (*dynamicresource.BulkCreateResult, error) {
	return &dynamicresource.BulkCreateResult{ClientErrors: this.refusal(dynamicresource.CrudActionBulkCreate)}, nil
}

func (this *SalesPromotionReadOnlyApplicationServiceImpl) Import(
	_ corectx.Context, _ dynamicresource.ImportCommand,
) (*dynamicresource.BulkCreateResult, error) {
	return &dynamicresource.BulkCreateResult{ClientErrors: this.refusal(dynamicresource.CrudActionImport)}, nil
}
