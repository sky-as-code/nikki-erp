// Package services holds the rules of the Payment & Invoice module: what may be paid, what may be
// refunded, and what state an order is left in when a gateway answers.
//
// Nothing here knows which gateway is in play. An order names a payment method, the method names
// an adapter, and the adapter is fetched from the registry — so adding a gateway is a new adapter
// package and a new row, not an edit to this file.
package services

import (
	"time"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// NewOrderDomainService wires the order rules onto the gateway registry they select from and the
// profiles that say which merchant account each payment lands in.
func NewOrderDomainService(
	registry *itGateway.Registry, profiles *PaymentProfileDomainService,
) *OrderDomainService {
	return &OrderDomainService{registry: registry, profiles: profiles, now: time.Now}
}

// OrderDomainService takes payments and gives them back.
//
// It is not a derived resource service: the two operations here are not CRUD on an order, they
// are money moving. They are reached as engine actions, and the built-in create is left alone —
// an order that exists without a payment attempt behind it would claim a collection nobody asked
// for.
type OrderDomainService struct {
	registry *itGateway.Registry

	// profiles resolves the merchant credentials an order is collected with. Every step after
	// the payment is created reads them back through it, because a refund or a callback has to
	// use the same account the money moved through. See order_profile.go.
	profiles *PaymentProfileDomainService

	// now is injected so the order-code date prefix can be pinned in a test.
	now func() time.Time
}

// CreatePaymentCommand is what a caller asks for. Amount and the method are theirs; every
// identifier on the resulting order is generated here.
type CreatePaymentCommand struct {
	Source          string
	Amount          decimal.Decimal
	PaymentMethodId string
	Content         *string
	ReturnUrl       *string

	// PaymentProfileId names the merchant account to collect into. Optional: an order without one
	// is collected with the credentials in this deployment's configuration.
	PaymentProfileId string

	// Metadata is the method-specific input, uninterpreted. Only the selected adapter reads it.
	Metadata map[string]any
}

// CreatePaymentResult is what the payer needs in order to pay. Both URLs are empty for a card
// terminal, where the prompt is pushed to the device the customer is standing at.
type CreatePaymentResult struct {
	OrderId string

	// OrderCode is the identifier the gateway knows this order by, and the key its callback will
	// arrive under. It is returned because the caller needs it to reconcile what the gateway later
	// reports against the order it opened; the ordering system quotes OrderId instead.
	OrderCode string

	QrCodeUrl string
	PayUrl    string
}

// CreatePayment records an order and asks its gateway to start collecting.
//
// The order is written before the gateway is called, and is kept whatever the gateway answers.
// The service this module supersedes deleted the order when the gateway refused, which lost the
// only evidence that the attempt had been made — including for the case where the gateway had in
// fact accepted it and the failure was in reading the reply. A refused attempt is left as
// payment_failed instead.
func (this *OrderDomainService) CreatePayment(
	ctx corectx.Context, cmd CreatePaymentCommand,
) (*CreatePaymentResult, *ft.ClientErrors, error) {
	vErrs := ft.NewClientErrors()

	method, err := this.loadActiveMethod(ctx, cmd.PaymentMethodId, vErrs)
	if err != nil || vErrs.Count() > 0 {
		return nil, vErrs, err
	}

	adapter, exists := this.registry.Get(derefString(method.GetAdapterCode()))
	if !exists {
		// The row names an adapter this build does not have enabled. That is a configuration
		// problem the caller cannot fix, but it is still their request that fails, and answering
		// 500 would have them retry against a deployment that will never accept it.
		appendOrderViolation(vErrs, "paymentinvoice.gateway_unavailable",
			"payment method '"+derefString(method.GetCode())+"' is not available on this deployment")
		return nil, vErrs, nil
	}

	assertAmountWithinMethodBounds(cmd.Amount, method, vErrs)

	// Errors accumulate rather than returning here: the adapter's own ValidateOrder runs next, and
	// short-circuiting would report an out-of-bounds amount while hiding a missing terminal id, so
	// the caller would fix one problem and be told about the other on the next attempt.
	profile, err := this.loadProfileForCreate(ctx, cmd.PaymentProfileId, *method, vErrs)
	if err != nil {
		return nil, vErrs, err
	}

	orderReq := itGateway.OrderRequest{
		Amount:        cmd.Amount,
		CurrencyCode:  "",
		Content:       cmd.Content,
		Metadata:      cmd.Metadata,
		MethodConfig:  method.GetConfig(),
		ProfileConfig: profileConfigOf(profile),
	}
	if err := adapter.ValidateOrder(ctx, orderReq, vErrs); err != nil {
		return nil, vErrs, err
	}
	if vErrs.Count() > 0 {
		return nil, vErrs, nil
	}

	adapterMeta, err := adapter.PrepareMetadata(ctx, orderReq)
	if err != nil {
		return nil, vErrs, err
	}

	order, transaction, err := this.persistNewOrder(ctx, cmd, *method, profile, adapterMeta)
	if err != nil {
		return nil, vErrs, err
	}
	if order == nil {
		appendOrderViolation(vErrs, "paymentinvoice.order_code_exhausted",
			"could not allocate an unused order code; please retry")
		return nil, vErrs, nil
	}

	return this.collect(ctx, adapter, *order, *transaction, orderReq, vErrs)
}

// collect calls the gateway and records what it answered, either way.
func (this *OrderDomainService) collect(
	ctx corectx.Context,
	adapter itGateway.PaymentGateway,
	order models.Order,
	transaction models.Transaction,
	orderReq itGateway.OrderRequest,
	vErrs *ft.ClientErrors,
) (*CreatePaymentResult, *ft.ClientErrors, error) {
	orderCode := derefString(order.GetOrderCode())

	created, gatewayErr := adapter.CreatePayment(ctx, itGateway.CreatePaymentRequest{
		OrderRequest: itGateway.OrderRequest{
			Amount:        orderReq.Amount,
			CurrencyCode:  orderReq.CurrencyCode,
			Content:       orderReq.Content,
			Metadata:      order.GetMetadata(),
			MethodConfig:  orderReq.MethodConfig,
			ProfileConfig: orderReq.ProfileConfig,
		},
		OrderCode: orderCode,
	})

	if gatewayErr != nil {
		if err := this.markCreateFailed(ctx, order, transaction, gatewayErr); err != nil {
			return nil, vErrs, err
		}
		// The gateway refusing is not a bug in this service, and the caller can act on it — by
		// paying another way, or by retrying. A 500 would say the opposite.
		appendOrderViolation(vErrs, "paymentinvoice.create_payment_failed",
			"the payment gateway refused the payment: "+gatewayErr.Error())
		return nil, vErrs, nil
	}

	if err := this.markCreateAccepted(ctx, order, transaction, created); err != nil {
		return nil, vErrs, err
	}

	return &CreatePaymentResult{
		OrderId:   derefString(order.GetOrderId()),
		OrderCode: orderCode,
		QrCodeUrl: created.QrCodeUrl,
		PayUrl:    created.PayUrl,
	}, vErrs, nil
}

// markCreateAccepted advances the order to processing and keeps the gateway's reply.
//
// processing rather than payment_success: for every gateway here, a create only means the payer
// has been given something to pay with. The money arriving is a separate event, and treating the
// create as the outcome would release goods against a payment nobody had made.
func (this *OrderDomainService) markCreateAccepted(
	ctx corectx.Context,
	order models.Order,
	transaction models.Transaction,
	created *itGateway.CreatePaymentResult,
) error {
	metadata := mergeMetadata(order.GetMetadata(), map[string]any{
		models.OrderMetaCreateResponse: created.RawResponse,
	})

	return withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus:   models.OrderStatusProcessing,
			models.OrderFieldMetadata: metadata,
		}); err != nil {
			return err
		}

		fields := dmodel.DynamicFields{}
		if created.RefTransactionId != "" {
			fields[models.TransactionFieldRefTransactionId] = created.RefTransactionId
		}
		if created.RawResponse != nil {
			fields[models.TransactionFieldRefPayload] = created.RawResponse
		}
		if len(fields) == 0 {
			return nil
		}
		return writeTransactionFields(tranxCtx, derefString(transaction.GetId()), fields)
	})
}

// markCreateFailed records a refused attempt without deleting anything.
func (this *OrderDomainService) markCreateFailed(
	ctx corectx.Context,
	order models.Order,
	transaction models.Transaction,
	gatewayErr error,
) error {
	metadata := mergeMetadata(order.GetMetadata(), map[string]any{
		models.OrderMetaCreateResponse: map[string]any{"error": gatewayErr.Error()},
	})

	return withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		if err := writeOrderFields(tranxCtx, derefString(order.GetId()), dmodel.DynamicFields{
			models.OrderFieldStatus:   models.OrderStatusPaymentFailed,
			models.OrderFieldMetadata: metadata,
		}); err != nil {
			return err
		}
		return writeTransactionFields(tranxCtx, derefString(transaction.GetId()), dmodel.DynamicFields{
			models.TransactionFieldStatus: models.TransactionStatusFailed,
		})
	})
}

// persistNewOrder writes the order and its payment transaction in one transaction.
//
// Both or neither: a transaction with no order describes a collection against nothing, and an
// order with no transaction loses the record of the attempt that this call is about to make.
//
// A nil order with no error means no unused order code could be allocated, which the caller
// reports as a client error rather than a failure.
func (this *OrderDomainService) persistNewOrder(
	ctx corectx.Context,
	cmd CreatePaymentCommand,
	method models.PaymentMethod,
	profile *models.PaymentProfile,
	adapterMeta map[string]any,
) (*models.Order, *models.Transaction, error) {
	orderCode, err := this.allocateOrderCode(ctx)
	if err != nil || orderCode == "" {
		return nil, nil, err
	}

	source := cmd.Source
	if source == "" {
		source = defaultSourceCode
	}
	orderId := BuildOrderId(source, derefString(method.GetCode()), orderCode)

	content := cmd.Content
	if content == nil || *content == "" {
		// The old service defaulted the description to the order id, which is what the payer
		// then sees on their statement. Keeping it means a transfer can still be traced back.
		content = &orderId
	}

	var order *models.Order
	var transaction *models.Transaction

	err = withOrderTransaction(ctx, func(tranxCtx corectx.Context) error {
		orderFields := dmodel.DynamicFields{
			models.OrderFieldOrderId:         orderId,
			models.OrderFieldOrderCode:       orderCode,
			models.OrderFieldSource:          source,
			models.OrderFieldStatus:          models.OrderStatusPending,
			models.OrderFieldAmount:          cmd.Amount,
			models.OrderFieldRefundAmount:    decimal.Zero,
			models.OrderFieldCurrencyId:      derefString(method.GetCurrencyId()),
			models.OrderFieldPaymentMethodId: derefString(method.GetId()),
			models.OrderFieldContent:         *content,
		}
		if cmd.ReturnUrl != nil && *cmd.ReturnUrl != "" {
			orderFields[models.OrderFieldReturnUrl] = *cmd.ReturnUrl
		}
		// Recorded on the order, not only used for this call: every later step — the refund, the
		// callback verification, the watchdog's question — has to reach for the same credentials
		// the money was taken with.
		if profile != nil {
			orderFields[models.OrderFieldPaymentProfileId] = derefString((*string)(profile.GetId()))
		}
		if metadata := mergeMetadata(cmd.Metadata, adapterMeta); metadata != nil {
			orderFields[models.OrderFieldMetadata] = metadata
		}

		createdOrder, err := createRecord(tranxCtx, models.OrderSchemaName, orderFields)
		if err != nil {
			return err
		}
		order = models.NewOrderFrom(createdOrder)

		createdTransaction, err := createRecord(tranxCtx, models.TransactionSchemaName, dmodel.DynamicFields{
			models.TransactionFieldOrderId:         derefString(order.GetId()),
			models.TransactionFieldOrderBusinessId: orderId,
			models.TransactionFieldStatus:          models.TransactionStatusPending,
			models.TransactionFieldAmount:          cmd.Amount,
			models.TransactionFieldCurrencyId:      derefString(method.GetCurrencyId()),
			models.TransactionFieldPaymentMethodId: derefString(method.GetId()),
			models.TransactionFieldTransactionType: models.TransactionTypePayment,
			models.TransactionFieldContent:         *content,
		})
		if err != nil {
			return err
		}
		transaction = models.NewTransactionFrom(createdTransaction)
		return nil
	})

	return order, transaction, err
}

// allocateOrderCode returns an unused order code, or "" when it could not find one.
//
// The uniqueness check is a best effort, not a guarantee: two callers can pass it with the same
// code before either writes. The unique index on order_code is the actual guarantee, and this
// only keeps that index from being hit in practice.
func (this *OrderDomainService) allocateOrderCode(ctx corectx.Context) (string, error) {
	prefix := EncodeDateToBase36(this.now())
	randomLength := orderCodeLength - len(prefix)
	if randomLength < 0 {
		return "", errors.Errorf("order code prefix '%s' does not fit in %d characters", prefix, orderCodeLength)
	}

	for range orderCodeMaxAttempts {
		suffix, err := randomCode(randomLength)
		if err != nil {
			return "", errors.Wrap(err, "allocateOrderCode")
		}

		candidate := prefix + suffix
		exists, err := orderCodeExists(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", nil
}

// assertAmountWithinMethodBounds checks the amount against what the gateway will accept.
//
// The bounds live on the payment-method row rather than in configuration, because each gateway
// sets its own floor and ceiling by contract, and those change without a release. The upper bound
// is exclusive, which is the old service's rule and is kept so an amount accepted before is
// accepted now.
func assertAmountWithinMethodBounds(
	amount decimal.Decimal, method *models.PaymentMethod, vErrs *ft.ClientErrors,
) {
	if amount.LessThanOrEqual(decimal.Zero) {
		appendFieldViolation(vErrs, models.OrderFieldAmount,
			"paymentinvoice.amount_not_positive", "the amount must be greater than zero")
		return
	}

	if min := method.GetMinAmount(); min != nil && amount.LessThan(*min) {
		appendFieldViolation(vErrs, models.OrderFieldAmount,
			"paymentinvoice.amount_below_minimum",
			"the amount is below the minimum of "+min.String()+" this payment method accepts")
	}
	if max := method.GetMaxAmount(); max != nil && amount.GreaterThanOrEqual(*max) {
		appendFieldViolation(vErrs, models.OrderFieldAmount,
			"paymentinvoice.amount_above_maximum",
			"the amount is at or above the limit of "+max.String()+" this payment method accepts")
	}
}
