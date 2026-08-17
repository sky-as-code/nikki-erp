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
// Both failures are the caller's: they named a method that does not exist, or one that is no
// longer taking payments. Neither is a reason to answer 500.
func (this *OrderDomainService) loadActiveMethod(
	ctx corectx.Context, methodId string, vErrs *ft.ClientErrors,
) (*models.PaymentMethod, error) {
	if methodId == "" {
		appendFieldViolation(vErrs, models.OrderFieldPaymentMethodId,
			"paymentinvoice.payment_method_required", "no payment method was given")
		return nil, nil
	}

	engine, err := engineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentMethodFieldId: methodId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadActiveMethod")
	}
	if found == nil || !found.HasData {
		appendFieldViolation(vErrs, models.OrderFieldPaymentMethodId,
			"paymentinvoice.payment_method_not_found",
			"no payment method with id '"+methodId+"'")
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
