package services

import (
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The statuses reported to the ordering system after a sweep reaches a verdict.
//
// They are this module's own status values rather than new names, so that everything the ordering
// system is told goes through the one translation in the app layer and there is no second
// vocabulary to keep in step.
const (
	OrderStatusPaidForSync    = models.OrderStatusPaymentSuccess
	OrderStatusFailedForSync  = models.OrderStatusPaymentFailed
	OrderStatusExpiredForSync = models.OrderStatusExpired
)

// The values written to an order's last_sync_status. They match the schema's enum.
const (
	SyncStatusSuccess = "success"
	SyncStatusFailure = "failure"
)

// syncLogEntriesKey is the key inside the sync_logs jsonmap that holds the entries.
//
// The column is a map rather than an array because the dynamic-model system has no array-typed
// field, so the list lives one level down under this key.
const syncLogEntriesKey = "entries"

// syncLogLimit bounds how many attempts are kept on one order.
//
// The log is evidence for a human reconciling a payment, not an audit trail, and an order whose
// tenant has been down for a week would otherwise accumulate an unbounded JSON blob in a column
// that is read whenever the order is.
const syncLogLimit = 20

// SyncOutcome is what came of notifying the ordering system once.
type SyncOutcome struct {
	Status   string
	Attempts int
	Detail   string
	At       time.Time
}

// PendingSyncOrder is an order the ordering system was never successfully told about.
type PendingSyncOrder struct {
	Pk        string
	OrderId   string
	Status    string
	ReturnUrl string
}

// FindOrdersNeedingSync returns settled orders whose notification has not got through.
//
// Only orders that reached a verdict are in scope — there is nothing to report about one still in
// flight — and only those whose last attempt failed. An order that has exhausted its attempts is
// excluded by the same count the client bounds itself by, so a permanently unreachable tenant
// stops being chased rather than being retried forever.
func FindOrdersNeedingSync(ctx corectx.Context) ([]PendingSyncOrder, error) {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.OrderFieldStatus, dmodel.In,
			models.OrderStatusPaymentSuccess,
			models.OrderStatusPaymentFailed,
			models.OrderStatusExpired,
		),
		*dmodel.NewSearchNode().NewCondition(
			models.OrderFieldLastSyncStatus, dmodel.Equals, SyncStatusFailure),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  sweepPageSize,
	})
	if err != nil {
		return nil, errors.Wrap(err, "FindOrdersNeedingSync")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}

	pending := make([]PendingSyncOrder, 0, len(found.Data.Items))
	for _, item := range found.Data.Items {
		order := models.NewOrderFrom(item)

		// An order with no callback URL cannot be notified and must not be re-queued every five
		// minutes. It should not be in this set at all — the client records such an order as a
		// success — so this guards against one that predates that behaviour.
		returnUrl := derefString(order.GetReturnUrl())
		if returnUrl == "" {
			continue
		}
		if syncAttemptsOf(order.GetSyncLogs()) >= syncLogLimit {
			continue
		}

		pending = append(pending, PendingSyncOrder{
			Pk:        derefString(order.GetId()),
			OrderId:   derefString(order.GetOrderId()),
			Status:    derefString(order.GetStatus()),
			ReturnUrl: returnUrl,
		})
	}
	return pending, nil
}

// SyncFactsFor reads the two order facts the notification carries beyond its status.
//
// The amount is rendered as a whole number because that is what the ordering system has always
// been sent; this module stores a decimal so that it can denominate in a currency with a different
// minor unit, and the truncation happens here at the boundary rather than in storage.
func (this *OrderDomainService) SyncFactsFor(
	ctx corectx.Context, orderId string,
) (amount int64, paymentMethod string, err error) {
	order, err := findOrderByBusinessId(ctx, orderId)
	if err != nil {
		return 0, "", err
	}
	if order == nil {
		return 0, "", errors.Errorf("SyncFactsFor: no order with id '%s'", orderId)
	}

	amount = derefDecimal(order.GetAmount()).IntPart()

	methodId := derefString((*string)(order.GetPaymentMethodId()))
	if methodId == "" {
		return amount, "", nil
	}

	engine, err := engineFor(models.PaymentMethodSchemaName)
	if err != nil {
		return 0, "", err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.PaymentMethodFieldId: methodId,
	})
	if err != nil {
		return 0, "", errors.Wrap(err, "SyncFactsFor")
	}
	if found == nil || !found.HasData {
		return amount, "", nil
	}

	// The method's own code is sent, not the adapter's: the ordering system was given the code of
	// the method the customer chose, and two methods may be served by one adapter.
	return amount, derefString(models.NewPaymentMethodFrom(found.Data).GetCode()), nil
}

// RecordSyncOutcome writes the result of a notification onto the order.
//
// Both the summary status and the appended log entry are written, because the two answer different
// questions: last_sync_status is what the retry sweep filters on, and the log is what a human
// reads when asking why a tenant says it was never told.
func (this *OrderDomainService) RecordSyncOutcome(
	ctx corectx.Context, orderPk string, outcome SyncOutcome,
) error {
	engine, err := engineFor(models.OrderSchemaName)
	if err != nil {
		return err
	}

	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{
		models.OrderFieldId: orderPk,
	})
	if err != nil {
		return errors.Wrap(err, "RecordSyncOutcome")
	}
	if found == nil || !found.HasData {
		return nil
	}

	order := models.NewOrderFrom(found.Data)
	return writeOrderFields(ctx, orderPk, dmodel.DynamicFields{
		models.OrderFieldLastSyncStatus: outcome.Status,
		models.OrderFieldSyncLogs:       appendSyncLog(order.GetSyncLogs(), outcome),
	})
}

// appendSyncLog adds one entry to an order's sync log, keeping the most recent entries.
//
// The keys are the old service's spellings, so a log written before this module took over reads
// the same way as one written after it.
func appendSyncLog(existing map[string]any, outcome SyncOutcome) map[string]any {
	entries := syncLogEntriesOf(existing)

	entries = append(entries, map[string]any{
		"type":      "paymentResult",
		"status":    outcome.Status,
		"attempts":  outcome.Attempts,
		"detail":    outcome.Detail,
		"timestamp": outcome.At.UTC().Format(time.RFC3339),
	})
	if len(entries) > syncLogLimit {
		entries = entries[len(entries)-syncLogLimit:]
	}

	return map[string]any{syncLogEntriesKey: entries}
}

// syncLogEntriesOf reads the entry list out of a sync_logs column.
//
// The column arrives as whatever the JSON decoder produced, so the list may be []any even though
// it was written as []map[string]any. Anything unreadable is treated as an empty log rather than
// failing: a malformed log must not stop a payment from being reported.
func syncLogEntriesOf(logs map[string]any) []any {
	if logs == nil {
		return nil
	}
	entries, _ := logs[syncLogEntriesKey].([]any)
	return entries
}

// syncAttemptsOf counts how many times an order has been reported on.
func syncAttemptsOf(logs map[string]any) int {
	return len(syncLogEntriesOf(logs))
}
