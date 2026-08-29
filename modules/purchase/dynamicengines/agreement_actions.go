package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
)

// The agreement's lifecycle operations. Archive and restore are deliberately absent: they go
// through the engine's built-in set_archived, because splitting them would let a role archive
// agreements it could not bring back.

const (
	PermissionClose     = "close"
	PermissionCreateRfq = "create_rfq"
)

const (
	ActionClose     = "close"
	ActionCreateRfq = "create_rfq"
)

func defineAgreementActions(engine drif.DynamicResourceEngine) error {
	if err := defineAgreementDeleteGuard(engine); err != nil {
		return err
	}
	return stdErr.Join(
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionConfirm,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/confirm",
			Permission:  PermissionConfirm,
			MainProcess: processAgreementConfirm,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionClose,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/close",
			Permission:  PermissionClose,
			MainProcess: processAgreementClose,
		}),
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCancel,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/cancel",
			Permission:  PermissionCancel,
			MainProcess: processAgreementCancel,
		}),
		// Raising an RFQ creates a purchase order, so it carries the order's create permission: a
		// role that may not create one by hand must not do it through an agreement.
		engine.DefineAction(drif.DynamicActionDefinition{
			ActionName:  ActionCreateRfq,
			ActionType:  drif.ActionTypeGeneric,
			RestPath:    ":id/create_rfq",
			Permission:  drif.PermissionCreate,
			MainProcess: processAgreementCreateRfq,
		}),
	)
}

func processAgreementConfirm(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := agreementServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Confirm(ctx, readAgreementId(input))
	return toMutateActionResult(result, err)
}

func processAgreementClose(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := agreementServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Close(ctx, readAgreementId(input))
	return toMutateActionResult(result, err)
}

func processAgreementCancel(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := agreementServiceOf(input)
	if err != nil {
		return nil, err
	}
	result, err := service.Cancel(
		ctx, readAgreementId(input), readStringParam(input.Params, paramReason))
	return toMutateActionResult(result, err)
}

// processAgreementCreateRfq returns the new order rather than an affected count so the caller has
// its id.
func processAgreementCreateRfq(ctx corectx.Context, input drif.ProcessInput) (*drif.ActionResult, error) {
	service, err := agreementServiceOf(input)
	if err != nil {
		return nil, err
	}
	orderService, err := orderServiceFromRegistry()
	if err != nil {
		return nil, err
	}

	result, err := service.CreateRfq(ctx, readAgreementId(input), orderService)
	if err != nil {
		return nil, err
	}
	out := &drif.ActionResult{
		ClientErrors: result.ClientErrors,
		HasData:      result.HasData,
	}
	if result.HasData {
		out.Data = result.Data
	}
	return out, nil
}

func agreementServiceOf(input drif.ProcessInput) (*services.PurchaseAgreementDomainServiceImpl, error) {
	service, ok := input.ResourceService.(*services.PurchaseAgreementDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the purchase agreement engine is not running the derived agreement service; " +
				"PurchaseModule.Init must install it with SetResourceService")
	}
	return service, nil
}

// orderServiceFromRegistry reaches the order's derived service, so create_rfq goes through the same
// rules as a hand-made order: the generated code, the forced RFQ status, the vendor and currency
// checks.
func orderServiceFromRegistry() (*services.PurchaseOrderDomainServiceImpl, error) {
	engine, err := services.EngineFor(models.PurchaseOrderSchemaName)
	if err != nil {
		return nil, err
	}
	service, ok := engine.ResourceService().(*services.PurchaseOrderDomainServiceImpl)
	if !ok {
		return nil, errors.New(
			"the purchase order engine is not running the derived order service; " +
				"PurchaseModule.Init must install it before the agreement actions run")
	}
	return service, nil
}

func readAgreementId(input drif.ProcessInput) string {
	return readStringParam(input.Params, paramOrderId)
}
