package app

import (
	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
	it "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/paymentmethod"
)

// maxPaymentMethods bounds the listing. Payment methods are configuration a human maintains, so a
// deployment has tens of them and never thousands; the bound exists so that a consumer calling this
// on every checkout cannot be handed an unbounded read.
const maxPaymentMethods = 500

// PaymentMethodApplicationServiceImpl answers the payment-method port.
//
// It holds the gateway registry because one of the three usability gates is "does this build ship
// the adapter this row names" — a question only the registry can answer, and the reason a consumer
// cannot derive usability from the row alone.
type PaymentMethodApplicationServiceImpl struct {
	registry *itGateway.Registry
}

func NewPaymentMethodApplicationServiceImpl(
	registry *itGateway.Registry,
) it.PaymentMethodAppService {
	return &PaymentMethodApplicationServiceImpl{registry: registry}
}

// ListPaymentMethods answers every method this deployment knows about, each already judged.
//
// No permission is asserted. This is a module-to-module port reading configuration that is not
// tenant data and carries no secret — the adapter credentials live in config, not in the fields
// exposed here — and the consumers are back-end services acting on behalf of a request whose own
// authorization has already run against their resource. Adding a check here would mean every module
// wanting to show a payment chooser needed a grant on paymentinvoice resources, which is a coupling
// the port exists to avoid.
func (this *PaymentMethodApplicationServiceImpl) ListPaymentMethods(
	ctx corectx.Context, query it.ListPaymentMethodsQuery,
) (*it.ListPaymentMethodsResult, error) {
	engine, err := services.EngineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: &dmodel.SearchGraph{},
		Page:  0,
		Size:  maxPaymentMethods,
	})
	if err != nil {
		return nil, errors.Wrap(err, "ListPaymentMethods")
	}
	if found.ClientErrors.Count() > 0 {
		return &it.ListPaymentMethodsResult{ClientErrors: found.ClientErrors}, nil
	}
	if !found.HasData {
		return &it.ListPaymentMethodsResult{HasData: true, Data: []it.PaymentMethodData{}}, nil
	}

	methods := make([]it.PaymentMethodData, 0, len(found.Data.Items))
	for _, record := range found.Data.Items {
		// The amount is not known here, so the bounds gate is not applied: a listing says whether a
		// method may ever be offered, not whether one particular payment would pass.
		data := this.judge(record, nil)
		if query.UsableOnly && !data.IsUsable {
			continue
		}
		methods = append(methods, data)
	}
	return &it.ListPaymentMethodsResult{HasData: true, Data: methods}, nil
}

// AssertUsable is the single place the three gates are applied together.
//
// It exists so that a consumer never re-implements them. Each is subtle in its own way: is_active
// treats nil as false, the adapter gate depends on the running build rather than on the data, and
// the upper amount bound is exclusive while the lower is inclusive. Any consumer reproducing all
// three would be correct only until this module changed one.
func (this *PaymentMethodApplicationServiceImpl) AssertUsable(
	ctx corectx.Context, query it.AssertUsableQuery,
) (*it.AssertUsableResult, error) {
	if query.PaymentMethodId == "" {
		return usableRejection("paymentinvoice.payment_method_required",
			"a payment method id is required"), nil
	}

	engine, err := services.EngineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentMethodFieldId: query.PaymentMethodId,
	})
	if err != nil {
		return nil, errors.Wrap(err, "AssertUsable")
	}
	if found == nil || !found.HasData {
		return usableRejection("paymentinvoice.payment_method_not_found",
			"no payment method with id "+quoted(query.PaymentMethodId)), nil
	}

	data := this.judge(found.Data, query.Amount)
	if !data.IsUsable {
		// The refusal still carries the method, so the caller can name it in a message rather than
		// only report that something was refused.
		result := usableRejection("paymentinvoice.payment_method_"+data.UnusableReason,
			"payment method "+quoted(data.Code)+" is not usable: "+data.UnusableReason)
		result.Data = data
		return result, nil
	}
	return &it.AssertUsableResult{HasData: true, Data: data}, nil
}

// judge applies the gates in order of how permanent the obstacle is: archived (retired for good),
// inactive (paused by an administrator), gateway missing (this deployment cannot serve it), then
// the amount. The first that closes is the reason reported, because a caller shown "amount too
// large" for a method that is also archived would fix the wrong thing.
func (this *PaymentMethodApplicationServiceImpl) judge(
	record dmodel.DynamicFields, amount *decimal.Decimal,
) it.PaymentMethodData {
	method := models.NewPaymentMethodFrom(record)
	active := method.GetIsActive()
	isActive := active != nil && *active

	data := it.PaymentMethodData{
		Id:        derefString(method.GetId()),
		Code:      derefString(method.GetCode()),
		IsActive:  isActive,
		MinAmount: method.GetMinAmount(),
		MaxAmount: method.GetMaxAmount(),
	}
	if name := method.GetName(); name != nil {
		data.Name = *name
	}

	switch {
	case boolOf(record, basemodel.FieldIsArchived):
		data.UnusableReason = it.ReasonArchived
	case !isActive:
		// Mirrors order_records.go:76 — a nil is_active counts as inactive, never as a default of
		// true. A row half-written is not a row that may take money.
		data.UnusableReason = it.ReasonInactive
	case !this.hasAdapter(method):
		data.UnusableReason = it.ReasonGatewayUnavailable
	case amount != nil && !withinBounds(*amount, method):
		data.UnusableReason = it.ReasonAmountOutOfBounds
	}
	data.IsUsable = data.UnusableReason == ""
	return data
}

func (this *PaymentMethodApplicationServiceImpl) hasAdapter(method *models.PaymentMethod) bool {
	if this.registry == nil {
		return false
	}
	_, exists := this.registry.Get(derefString(method.GetAdapterCode()))
	return exists
}

// withinBounds reproduces assertAmountWithinMethodBounds (order_domservice.go:341) exactly,
// including its asymmetry: the minimum is inclusive and the maximum is EXCLUSIVE. That asymmetry is
// a deliberate bug-compatibility choice with the service this module supersedes, so it is copied
// rather than corrected — a port answering "usable" for an amount that CreatePayment then refused
// would be worse than no port at all.
func withinBounds(amount decimal.Decimal, method *models.PaymentMethod) bool {
	if !amount.IsPositive() {
		return false
	}
	if minAmount := method.GetMinAmount(); minAmount != nil && amount.LessThan(*minAmount) {
		return false
	}
	if maxAmount := method.GetMaxAmount(); maxAmount != nil && amount.GreaterThanOrEqual(*maxAmount) {
		return false
	}
	return true
}

func usableRejection(key, message string) *it.AssertUsableResult {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(models.PaymentMethodSchemaName, key, message))
	return &it.AssertUsableResult{ClientErrors: *vErrs}
}

func quoted(value string) string {
	return "'" + value + "'"
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolOf(record dmodel.DynamicFields, field string) bool {
	if record == nil {
		return false
	}
	value, ok := record[field]
	if !ok || value == nil {
		return false
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return false
}
