// Package app holds the payment module's background work: notifying the ordering system that a
// payment settled, and the sweeps that recover orders no callback ever arrived for.
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The wire contract with the ordering system.
//
// These spellings are the old NestJS service's and are load-bearing: the vending machines read
// them, and they were not updated when this module replaced that service. Renaming a key here, or
// sending the snake_case status this module stores internally, silently stops every machine from
// learning that a payment succeeded — the machine holds the goods and the customer has paid.
const (
	syncTypePaymentResult = "paymentResult"

	syncStatusSuccess = "success"
	syncStatusFailure = "failure"
)

// ResultSyncPayload is the body posted to an order's return_url.
type ResultSyncPayload struct {
	Type          string `json:"type"`
	Status        string `json:"status"`
	OrderId       string `json:"orderId"`
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"paymentMethod"`
	ResponseTime  int64  `json:"responseTime"`
}

// ResultSyncRequest is what the caller knows about the notification to send.
type ResultSyncRequest struct {
	ReturnUrl string
	OrderId   string

	// Status is this module's own snake_case status. It is translated on the way out; callers
	// pass what they hold rather than pre-translating, so the mapping lives in one place.
	Status string

	Amount        int64
	PaymentMethod string
}

// ResultSyncOutcome records what came of one notification, for the order's sync log.
type ResultSyncOutcome struct {
	// Status is "success" or "failure", stored verbatim in the order's last_sync_status.
	Status string

	// Attempts is how many times the tenant was called, which is what bounds the retry job.
	Attempts int

	// Detail explains a failure. It is deliberately short: it goes into a JSON column that is
	// read by a human deciding whether to retry, not parsed.
	Detail string
}

func (this ResultSyncOutcome) Succeeded() bool {
	return this.Status == syncStatusSuccess
}

// ResultSyncClient notifies the ordering system that a payment reached a verdict.
type ResultSyncClient struct {
	httpClient *http.Client
	maxRetries int
}

// NewResultSyncClient builds the client with the deployment's timeout and retry bounds.
//
// The timeout is per attempt rather than across all of them, because it is a bound on how long one
// unresponsive tenant may hold a worker, and that is what matters when a sweep is working through
// a backlog.
func NewResultSyncClient(timeout time.Duration, maxRetries int) *ResultSyncClient {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if maxRetries < 1 {
		maxRetries = 1
	}
	return &ResultSyncClient{
		httpClient: &http.Client{Timeout: timeout},
		maxRetries: maxRetries,
	}
}

// Sync posts the payment result, retrying a failed attempt with linear backoff.
//
// A retry is safe because the notification is a statement of fact rather than a command: the same
// result arriving twice tells the ordering system the same thing. That is also why a non-2xx is
// retried at all — the tenant being briefly down must not lose the only notice it gets.
func (this *ResultSyncClient) Sync(ctx context.Context, req ResultSyncRequest) ResultSyncOutcome {
	if req.ReturnUrl == "" {
		// An order with no callback URL has nobody to notify. That is a complete outcome, not a
		// failure: recording it as one would have the retry job chase it forever.
		return ResultSyncOutcome{Status: syncStatusSuccess}
	}

	body, err := json.Marshal(ResultSyncPayload{
		Type:          syncTypePaymentResult,
		Status:        StatusToLegacyEnum(req.Status),
		OrderId:       req.OrderId,
		Amount:        req.Amount,
		PaymentMethod: req.PaymentMethod,
		ResponseTime:  time.Now().UnixMilli(),
	})
	if err != nil {
		return ResultSyncOutcome{Status: syncStatusFailure, Attempts: 0, Detail: "payload could not be encoded"}
	}

	var lastDetail string
	for attempt := 1; attempt <= this.maxRetries; attempt++ {
		detail, ok := this.postOnce(ctx, req.ReturnUrl, body)
		if ok {
			return ResultSyncOutcome{Status: syncStatusSuccess, Attempts: attempt}
		}
		lastDetail = detail

		// A canceled context means the process is shutting down. Continuing would spend the
		// shutdown window on a call that cannot complete, so the attempts made so far are
		// reported and the retry job picks it up next run.
		if ctx.Err() != nil {
			break
		}
		if attempt < this.maxRetries {
			select {
			case <-time.After(time.Duration(attempt) * time.Second):
			case <-ctx.Done():
			}
		}
	}

	return ResultSyncOutcome{
		Status:   syncStatusFailure,
		Attempts: this.maxRetries,
		Detail:   lastDetail,
	}
}

// postOnce makes one attempt and reports whether the tenant accepted it.
func (this *ResultSyncClient) postOnce(ctx context.Context, url string, body []byte) (string, bool) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		// A URL that cannot be turned into a request will not become one on a retry, but it is
		// still reported as a failure so the order carries the reason.
		return "the return_url is not a usable address", false
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := this.httpClient.Do(request)
	if err != nil {
		return "the ordering system could not be reached", false
	}
	defer response.Body.Close()

	// The body is drained and discarded so the connection can be reused; the ordering system's
	// reply carries nothing this module acts on.
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return "", true
	}
	return "the ordering system answered " + response.Status, false
}

// StatusToLegacyEnum renders an order status the way the ordering system spells it.
//
// This module stores snake_case, the service it replaced sent camelCase, and the machines reading
// these notifications were never updated. The translation is therefore part of the external
// contract rather than a formatting preference, and it is exported so the one test that pins it
// can reach it.
//
// An unrecognised status is passed through unchanged: inventing a spelling for a status added
// later would be worse than sending the one the database holds, which is at least traceable.
func StatusToLegacyEnum(status string) string {
	switch status {
	case models.OrderStatusPaymentSuccess:
		return "paymentSuccess"
	case models.OrderStatusPaymentFailed:
		return "paymentFailed"
	case models.OrderStatusRefundSuccess:
		return "refundSuccess"
	case models.OrderStatusRefundFailed:
		return "refundFailed"
	case models.OrderStatusProcessing:
		return "processing"
	case models.OrderStatusPending:
		return "pending"
	case models.OrderStatusCanceled:
		return "canceled"
	case models.OrderStatusExpired:
		return "expired"
	default:
		return status
	}
}
