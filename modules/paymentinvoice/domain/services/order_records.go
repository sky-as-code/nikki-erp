package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// createRecord writes one record through a resource's own service and returns it as stored.
//
// It goes through the service rather than the repository so that the id and the audit columns are
// generated the way every other write generates them, and so the fields are validated against the
// schema. The record comes back with its generated id, which the caller needs in order to link
// the transaction to the order it belongs to.
//
// A validation failure here is not a client error: every field written by the payment flow is
// composed by this module, so a rejection means this code built a record its own schema refuses.
func createRecord(
	ctx corectx.Context, schemaName string, fields dmodel.DynamicFields,
) (dmodel.DynamicFields, error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return nil, err
	}

	result, err := engine.ResourceService().Create(ctx, fields)
	if err != nil {
		return nil, errors.Wrapf(err, "createRecord(%s)", schemaName)
	}
	if result == nil || result.ClientErrors.Count() > 0 {
		return nil, errors.Errorf(
			"createRecord(%s): the record this module composed was rejected by its own schema: %v",
			schemaName, result.ClientErrors)
	}
	if !result.HasData {
		return nil, errors.Errorf("createRecord(%s): no record was returned", schemaName)
	}
	return result.Data, nil
}

// loadActiveMethod fetches the payment method an order names, and refuses one withdrawn from use.
//
// The method may be named by id or by code, and the two are looked up the same way because both
// columns are unique. The code exists for callers migrating off the standalone service, which was
// only ever told "momo", "vietqr" or "mpos"; the id is what a picker on the REST surface holds.
//
// Every failure here is the caller's: they named no method, one that does not exist, or one that
// is no longer taking payments. None of them is a reason to answer 500.
func (this *OrderDomainService) loadActiveMethod(
	ctx corectx.Context, cmd CreatePaymentCommand, vErrs *ft.ClientErrors,
) (*models.PaymentMethod, error) {
	key, named := methodLookupKey(cmd)
	if !named {
		appendFieldViolation(vErrs, models.OrderFieldPaymentMethodId,
			"paymentinvoice.payment_method_required", "no payment method was given")
		return nil, nil
	}

	engine, err := engineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "loadActiveMethod")
	}
	if found == nil || !found.HasData {
		appendFieldViolation(vErrs, models.OrderFieldPaymentMethodId,
			"paymentinvoice.payment_method_not_found",
			"no payment method matching "+describeMethodKey(cmd))
		return nil, nil
	}

	method := models.NewPaymentMethodFrom(found.Data)
	if active := method.GetIsActive(); active == nil || !*active {
		appendFieldViolation(vErrs, models.OrderFieldPaymentMethodId,
			"paymentinvoice.payment_method_inactive",
			"payment method '"+derefString(method.GetCode())+"' is not taking payments")
		return nil, nil
	}
	return method, nil
}

// methodLookupKey turns whichever identifier the caller gave into the unique key to fetch by.
//
// The id wins when both are present. They name the same row when the caller is consistent, and
// preferring the id keeps the behaviour of the existing REST callers exactly as it was.
func methodLookupKey(cmd CreatePaymentCommand) (dmodel.DynamicFields, bool) {
	if cmd.PaymentMethodId != "" {
		return dmodel.DynamicFields{models.PaymentMethodFieldId: cmd.PaymentMethodId}, true
	}
	if cmd.PaymentMethodCode != "" {
		return dmodel.DynamicFields{models.PaymentMethodFieldCode: cmd.PaymentMethodCode}, true
	}
	return nil, false
}

// describeMethodKey names what the caller asked for, so a "not found" says which of the two
// identifiers was wrong rather than only that something was.
func describeMethodKey(cmd CreatePaymentCommand) string {
	if cmd.PaymentMethodId != "" {
		return "id '" + cmd.PaymentMethodId + "'"
	}
	return "code '" + cmd.PaymentMethodCode + "'"
}
