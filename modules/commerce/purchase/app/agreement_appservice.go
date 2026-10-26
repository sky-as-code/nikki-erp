package app

import (
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/services"
	itAgreement "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/interfaces/agreement"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The agreement's authorized surface: the built-in CRUD plus confirm, close, cancel and
// create_rfq. Archive and restore are deliberately absent: they go through the built-in
// set_archived, because splitting them would let a role archive agreements it could not bring back.

type AgreementApplicationServiceImpl struct {
	composable.CrudApplicationService
	agreements *services.PurchaseAgreementDomainServiceImpl
}

// NewAgreementApplicationService needs the derived agreement service, which alone carries the
// lifecycle operations; any other wiring fails the boot.
func NewAgreementApplicationService(base composable.CrudApplicationService) itAgreement.AgreementCrudApplicationService {
	agreements, ok := base.DomainService().(*services.PurchaseAgreementDomainServiceImpl)
	if !ok {
		panic("the purchase agreement onion is not running the derived agreement domain service")
	}
	return &AgreementApplicationServiceImpl{CrudApplicationService: base, agreements: agreements}
}

func (this *AgreementApplicationServiceImpl) recordAction(
	ctx corectx.Context, permission string, cmd itAgreement.ActionCommand, run func() (*actionResult, error),
) (*actionResult, error) {
	if cErrs, err := assertRecordAction(this, ctx, permission, cmd); cErrs != nil || err != nil {
		return actionFailure(cErrs, err)
	}
	return run()
}

func (this *AgreementApplicationServiceImpl) ConfirmAction(
	ctx corectx.Context, cmd itAgreement.ActionCommand,
) (*itAgreement.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionConfirm, cmd, func() (*actionResult, error) {
		return toActionResult(this.agreements.Confirm(ctx, readStringParam(cmd, paramRecordId)))
	})
}

func (this *AgreementApplicationServiceImpl) CloseAction(
	ctx corectx.Context, cmd itAgreement.ActionCommand,
) (*itAgreement.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionClose, cmd, func() (*actionResult, error) {
		return toActionResult(this.agreements.Close(ctx, readStringParam(cmd, paramRecordId)))
	})
}

func (this *AgreementApplicationServiceImpl) CancelAction(
	ctx corectx.Context, cmd itAgreement.ActionCommand,
) (*itAgreement.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionCancel, cmd, func() (*actionResult, error) {
		return toActionResult(this.agreements.Cancel(
			ctx, readStringParam(cmd, paramRecordId), readStringParam(cmd, paramReason)))
	})
}

// CreateRfqAction raises a purchase order, so it carries "create" on the agreement resource, the
// code the legacy action asserted. It answers the new order rather than an affected count.
func (this *AgreementApplicationServiceImpl) CreateRfqAction(
	ctx corectx.Context, cmd itAgreement.ActionCommand,
) (*itAgreement.ActionResult, error) {
	return this.recordAction(ctx, constants.ActionCreate, cmd, func() (*actionResult, error) {
		return toActionResult(this.agreements.CreateRfq(ctx, readStringParam(cmd, paramRecordId)))
	})
}
