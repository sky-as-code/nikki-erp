package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
	itInvoicing "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external/invoicing"
	itFiscal "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/fiscal"
)

// The fiscal request and billing instruction surfaces.
//
// The invoicing port is optional: no e-invoicing adapter ships in every deployment, and the
// operation tolerates its absence rather than refusing -- a sale is still a sale when the fiscal
// document cannot be issued yet.

type SalesFiscalRequestApplicationServiceImpl struct {
	composable.CrudApplicationService

	invoicingProvider itInvoicing.InvoicingExtService
	effectiveSettings itExt.EffectiveSettingsExtService
	partyPort         itExt.PartyExtService
}

func NewSalesFiscalRequestApplicationService(
	base composable.CrudApplicationService,
	invoicing itInvoicing.InvoicingExtService,
	settings itExt.EffectiveSettingsExtService,
	parties itExt.PartyExtService,
) itFiscal.SalesFiscalRequestApplicationService {
	return &SalesFiscalRequestApplicationServiceImpl{
		CrudApplicationService: base,
		invoicingProvider:      invoicing,
		effectiveSettings:      settings,
		partyPort:              parties,
	}
}

// RequestInvoice is collection-level: it creates the fiscal request.
func (this *SalesFiscalRequestApplicationServiceImpl) RequestInvoice(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, PermissionRequestInvoice, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runRequestInvoice(ctx, cmd)
}

type SalesBillingInstructionApplicationServiceImpl struct {
	composable.CrudApplicationService

	partyPort         itExt.PartyExtService
	effectiveSettings itExt.EffectiveSettingsExtService
}

func NewSalesBillingInstructionApplicationService(
	base composable.CrudApplicationService,
	parties itExt.PartyExtService,
	settings itExt.EffectiveSettingsExtService,
) itFiscal.SalesBillingInstructionApplicationService {
	return &SalesBillingInstructionApplicationServiceImpl{
		CrudApplicationService: base,
		partyPort:              parties,
		effectiveSettings:      settings,
	}
}

// Creating an instruction is collection-level; the other four address one.
func (this *SalesBillingInstructionApplicationServiceImpl) CreateInstruction(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if _, cErrs := this.AssertAction(ctx, PermissionCreateBillingInstruction, cmd); cErrs != nil {
		return anyFailure(cErrs, nil)
	}
	return this.runCreateBillingInstruction(ctx, cmd)
}

func (this *SalesBillingInstructionApplicationServiceImpl) UpdateInstruction(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionUpdateBillingInstruction, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runUpdateBillingInstruction(ctx, cmd)
}

func (this *SalesBillingInstructionApplicationServiceImpl) MarkReady(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionMarkBillingReady, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runMarkBillingReady(ctx, cmd)
}

func (this *SalesBillingInstructionApplicationServiceImpl) RevertToDraft(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionRevertBillingToDraft, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runRevertBillingToDraft(ctx, cmd)
}

func (this *SalesBillingInstructionApplicationServiceImpl) Cancel(
	ctx corectx.Context, cmd itFiscal.FiscalActionCommand,
) (*dyn.OpResult[any], error) {
	if cErrs, err := assertRecordAction(this, ctx, PermissionCancelBillingInstruction, cmd); cErrs != nil || err != nil {
		return anyFailure(cErrs, err)
	}
	return this.runCancelBillingInstruction(ctx, cmd)
}

func NewSalesBillingIssuanceAttemptApplicationService(base composable.CrudApplicationService) itFiscal.SalesBillingIssuanceAttemptApplicationService {
	return &SalesBillingIssuanceAttemptApplicationServiceImpl{CrudApplicationService: base}
}

type SalesBillingIssuanceAttemptApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// Parameter names the moved bodies read.
const (
	paramSalesOrderId            = "sales_order_id"
	paramTaxId                   = "tax_id"
	paramLegalName               = "legal_name"
	paramBillingAddress          = "billing_address"
	paramBillingEmail            = "billing_email"
	paramBillingSource           = "source"
	paramFetchLatestPartyDetails = "fetch_latest_party_details"
)

// The moved action bodies follow.

func (this *SalesFiscalRequestApplicationServiceImpl) runRequestInvoice(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	result, vErrs, err := services.RequestInvoice(ctx, services.RequestInvoiceParams{
		SalesBillId:             readStringParam(params, "sales_bill_id"),
		Intent:                  readStringParam(params, "intent"),
		OriginalFiscalRequestId: readStringParam(params, "original_fiscal_request_id"),
		Reason:                  readStringParam(params, "reason"),
		IdempotencyKey:          readStringParam(params, "idempotency_key"),
		Buyer:                   readBuyerInfo(params),
	}, this.invoicingProvider)
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{
		HasData: true,
		Data: map[string]any{
			"fiscal_document_request_id": result.FiscalDocumentRequestId,
			"sales_bill_id":              result.SalesBillId,

			// Returned so a caller does not read a successful call as an issued document; an
			// in-flight request is pending.
			"status": result.Status,

			// Empty unless the provider confirmed. The only durable link to the document.
			"provider_reference": result.ProviderReference,
			"already_existed":    result.AlreadyExisted,
		},
	}, nil
}

// readBuyerInfo reads the buyer's fiscal identity, nested under `buyer` rather than flattened so
// the wire shape matches the stored snapshot.
func readBuyerInfo(params map[string]any) itInvoicing.BuyerInfo {
	nested, ok := params["buyer"].(map[string]any)
	if !ok {
		return itInvoicing.BuyerInfo{}
	}
	return itInvoicing.BuyerInfo{
		TaxCode:   readAnyString(nested, "tax_code"),
		LegalName: readAnyString(nested, "legal_name"),
		Address:   readAnyString(nested, "address"),
		Email:     readAnyString(nested, "email"),
	}
}

// readAnyString avoids a bare type assertion: a JSON round-trip can hand back a different concrete
// type, and a bare assertion panics the request.
func readAnyString(values map[string]any, field string) string {
	value, ok := values[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}

func (this *SalesBillingInstructionApplicationServiceImpl) runCreateBillingInstruction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	instructionId, vErrs, err := services.CreateBillingInstruction(ctx,
		services.CreateBillingInstructionParams{
			SalesOrderId:  readStringParam(params, paramSalesOrderId),
			BillToPartyId: readStringParam(params, paramBillToPartyId),
			Source:        readStringParam(params, paramBillingSource),
			Snapshot:      readBillingSnapshot(params),
		})
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}

	return &dyn.OpResult[any]{HasData: true, Data: map[string]any{
		"sales_billing_instruction_id": instructionId,
	}}, nil
}

func (this *SalesBillingInstructionApplicationServiceImpl) runUpdateBillingInstruction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.UpdateBillingInstructionSnapshot(ctx,
		readStringParam(params, paramRecordId), readBillingSnapshot(params))
	return billingActionResult(vErrs, err)
}

func (this *SalesBillingInstructionApplicationServiceImpl) runMarkBillingReady(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.MarkBillingInstructionReady(ctx,
		readStringParam(params, paramRecordId),
		readBoolParam(params, paramFetchLatestPartyDetails),
		this.partyPort)
	return billingActionResult(vErrs, err)
}

func (this *SalesBillingInstructionApplicationServiceImpl) runRevertBillingToDraft(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.RevertBillingInstructionToDraft(ctx,
		readStringParam(params, paramRecordId))
	return billingActionResult(vErrs, err)
}

func (this *SalesBillingInstructionApplicationServiceImpl) runCancelBillingInstruction(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[any], error) {
	vErrs, err := services.CancelBillingInstruction(ctx,
		readStringParam(params, paramRecordId))
	return billingActionResult(vErrs, err)
}

// readBillingSnapshot pulls the four confirmed fiscal fields off the request.
func readBillingSnapshot(params dmodel.DynamicFields) services.BillingSnapshot {
	return services.BillingSnapshot{
		TaxId:          readStringParam(params, paramTaxId),
		LegalName:      readStringParam(params, paramLegalName),
		BillingAddress: readStringParam(params, paramBillingAddress),
		BillingEmail:   readStringParam(params, paramBillingEmail),
	}
}

// billingActionResult is the shape every lifecycle action returns: the transition either happened or
// was refused for a reason the caller can act on.
func billingActionResult(vErrs *ft.ClientErrors, err error) (*dyn.OpResult[any], error) {
	if err != nil {
		return nil, err
	}
	if vErrs != nil {
		return &dyn.OpResult[any]{ClientErrors: *vErrs}, nil
	}
	return &dyn.OpResult[any]{HasData: true, Data: map[string]any{"updated": true}}, nil
}
