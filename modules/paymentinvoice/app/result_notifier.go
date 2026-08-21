package app

import (
	"context"
	"time"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// NotifyTarget is the order a notification is about, reduced to what sending one needs.
//
// The three fields come from whoever settled the order and are passed in rather than re-read,
// because the callback path already holds them from the transaction that applied the verdict.
type NotifyTarget struct {
	// Pk is the order's primary key, which the outcome is recorded against.
	Pk string

	// OrderId is the identifier the ordering system knows the order by.
	OrderId string

	// ReturnUrl is where that system listens. An empty one is a complete outcome, not a failure —
	// see ResultSyncClient.Sync.
	ReturnUrl string
}

// orderSyncStore is the part of the order service a notification needs.
//
// It is an interface only so that the notifier can be tested without a database: sending a
// notification is the step that decides whether a paying customer's machine opens, and that is
// worth being able to exercise directly.
type orderSyncStore interface {
	SyncFactsFor(ctx corectx.Context, orderId string) (*services.SyncFacts, error)
	RecordSyncOutcome(ctx corectx.Context, orderPk string, outcome services.SyncOutcome) error
}

// ResultNotifier tells the ordering system what became of an order and records how that went.
//
// Both paths that settle an order use it — the gateway callbacks and the watchdog sweep — so that
// a payment is reported the same way however it was discovered, and so the recording of the
// outcome cannot be forgotten by one of them.
type ResultNotifier struct {
	orders orderSyncStore
	client *ResultSyncClient
	logger logging.LoggerService

	// now is injected so a test can pin the timestamp written into the sync log.
	now func() time.Time
}

// NewResultNotifier takes the order service concretely rather than as the interface above, because
// that is what the dependency container holds and what every caller already has.
func NewResultNotifier(
	orders *services.OrderDomainService,
	client *ResultSyncClient,
	logger logging.LoggerService,
) *ResultNotifier {
	return &ResultNotifier{
		orders: orders,
		client: client,
		logger: logger,
		now:    time.Now,
	}
}

// Notify posts the payment result and writes the outcome onto the order.
//
// The outcome is recorded whether or not it succeeded: a failure is what the retry sweep looks
// for, so losing it would mean the notification is never re-attempted.
//
// It returns nothing. Nobody can act on a failure here — the gateway must still be answered, and
// the sweep has already moved on to the next order — so a failure is logged and left for the
// retry to pick up off the order's own state.
func (this *ResultNotifier) Notify(ctx corectx.Context, target NotifyTarget, status string) {
	facts, err := this.orders.SyncFactsFor(ctx, target.OrderId)
	if err != nil {
		this.logger.Warnf("paymentinvoice: order '%s' could not be read for notification: %s",
			target.OrderId, err.Error())
		return
	}

	outcome := this.client.Sync(ctx, ResultSyncRequest{
		ReturnUrl:     target.ReturnUrl,
		OrgId:         facts.OrgId,
		OrderId:       target.OrderId,
		Status:        status,
		Amount:        facts.Amount,
		PaymentMethod: facts.PaymentMethod,
	})

	if err := this.orders.RecordSyncOutcome(ctx, target.Pk, SyncOutcomeOf(outcome, this.now())); err != nil {
		this.logger.Errorf("paymentinvoice: order '%s' sync outcome could not be recorded: %s",
			target.OrderId, err.Error())
	}
	if !outcome.Succeeded() {
		this.logger.Warnf("paymentinvoice: order '%s' could not be reported to its caller: %s",
			target.OrderId, outcome.Detail)
	}
}

// NotifyDetached sends the notification on its own goroutine, off the caller's request.
//
// This is what the gateway callbacks use. Notifying inline would hold the gateway's connection for
// as long as the ordering system takes to answer — up to the client's timeout times its retries —
// and every one of these gateways treats a slow callback as a failed one and sends it again. The
// customer's machine would then be told once per retry, and the gateway would eventually give up
// on a callback that had in fact been applied the first time.
//
// The context is detached from the request so the send is not canceled the moment the handler
// answers the gateway, while keeping the values the request carries — the tenant above all, which
// every query below this point is scoped by.
//
// A send that is lost to a crash or a shutdown is not lost for good: the order was settled in its
// own transaction before this was reached, and an order settled but never reported is exactly what
// the retry sweep looks for.
func (this *ResultNotifier) NotifyDetached(ctx corectx.Context, target NotifyTarget, status string) {
	bgCtx := corectx.NewRequestContext(context.WithoutCancel(ctx.InnerContext()))

	go func() {
		defer func() {
			// A panic on this goroutine would take the process down with it, and the request that
			// spawned it has already been answered. It is contained and logged instead.
			if recovered := recover(); recovered != nil {
				this.logger.Errorf("paymentinvoice: notifying order '%s' panicked: %v",
					target.OrderId, recovered)
			}
		}()
		this.Notify(bgCtx, target, status)
	}()
}

// SyncOutcomeOf renders a send's result as the record written onto the order.
func SyncOutcomeOf(outcome ResultSyncOutcome, at time.Time) services.SyncOutcome {
	return services.SyncOutcome{
		Status:   outcome.Status,
		Attempts: outcome.Attempts,
		Detail:   outcome.Detail,
		At:       at,
	}
}
