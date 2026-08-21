package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The notification this client sends is how a vending machine learns it may release the goods. A
// defect here is a customer who paid and got nothing, so these tests pin both the wire shape and
// the retry behaviour.

func newTestClient(server *httptest.Server, retries int) *ResultSyncClient {
	client := NewResultSyncClient(2*time.Second, retries)
	_ = server
	return client
}

// The JSON keys are the old NestJS service's, and the machines reading them were never updated.
// A rename is a silent outage: the machine sees a body it cannot parse and holds the goods.
func TestThePayloadKeepsTheLegacyWireShape(t *testing.T) {
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outcome := newTestClient(server, 1).Sync(context.Background(), ResultSyncRequest{
		ReturnUrl:     server.URL,
		OrgId:         "01JBQ0000000000000000ORG",
		OrderId:       "VDMCMOM0Q8HABCDEFGH",
		Status:        models.OrderStatusPaymentSuccess,
		Amount:        150000,
		PaymentMethod: "momo",
	})

	require.True(t, outcome.Succeeded())
	assert.Equal(t, "paymentResult", received["type"])
	assert.Equal(t, "paymentSuccess", received["status"])
	assert.Equal(t, "VDMCMOM0Q8HABCDEFGH", received["orderId"])
	assert.Equal(t, float64(150000), received["amount"])
	assert.Equal(t, "momo", received["paymentMethod"])
	assert.NotZero(t, received["responseTime"])

	// The ordering system matches this against its own copy of the order and answers "no such
	// order" without it. The key is snake_case on purpose — see ResultSyncPayload.
	assert.Equal(t, "01JBQ0000000000000000ORG", received["org_id"])
}

// The status crosses the wire in the spelling the ordering system knows. Sending this module's
// own snake_case would have every machine fall through to its unknown-status branch.
func TestTheStatusIsTranslatedToTheLegacySpelling(t *testing.T) {
	for internal, legacy := range map[string]string{
		models.OrderStatusPaymentSuccess: "paymentSuccess",
		models.OrderStatusPaymentFailed:  "paymentFailed",
		models.OrderStatusRefundSuccess:  "refundSuccess",
		models.OrderStatusRefundFailed:   "refundFailed",
		models.OrderStatusExpired:        "expired",
		models.OrderStatusCanceled:       "canceled",
	} {
		assert.Equal(t, legacy, StatusToLegacyEnum(internal), internal)
	}
}

// A status this module gains later must not be silently renamed to something the ordering system
// would misread. Passing it through unchanged is at least traceable to a real database value.
func TestAnUnknownStatusIsPassedThroughUnchanged(t *testing.T) {
	assert.Equal(t, "something_new", StatusToLegacyEnum("something_new"))
}

// A tenant that is briefly down must not lose the only notice it gets that a payment settled.
func TestATransientFailureIsRetriedUntilItSucceeds(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outcome := newTestClient(server, 3).Sync(context.Background(), ResultSyncRequest{
		ReturnUrl: server.URL,
		OrderId:   "ORDER1",
		Status:    models.OrderStatusPaymentSuccess,
	})

	assert.True(t, outcome.Succeeded())
	assert.Equal(t, 3, outcome.Attempts)
	assert.EqualValues(t, 3, atomic.LoadInt32(&calls))
}

// Retries are bounded. A permanently unreachable tenant must stop being called, or one bad
// return_url occupies the sweep forever.
func TestRetriesStopAtTheConfiguredBound(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	outcome := newTestClient(server, 2).Sync(context.Background(), ResultSyncRequest{
		ReturnUrl: server.URL,
		OrderId:   "ORDER1",
		Status:    models.OrderStatusPaymentSuccess,
	})

	assert.False(t, outcome.Succeeded())
	assert.Equal(t, 2, outcome.Attempts)
	assert.EqualValues(t, 2, atomic.LoadInt32(&calls))
	assert.NotEmpty(t, outcome.Detail, "a failure must say why, for whoever reads the sync log")
}

// An order with no callback URL has nobody to tell. Recording that as a failure would have the
// retry sweep chase it every five minutes forever.
func TestAnOrderWithNoCallbackUrlIsACompleteOutcome(t *testing.T) {
	outcome := NewResultSyncClient(time.Second, 3).Sync(context.Background(), ResultSyncRequest{
		ReturnUrl: "",
		OrderId:   "ORDER1",
		Status:    models.OrderStatusPaymentSuccess,
	})

	assert.True(t, outcome.Succeeded())
	assert.Zero(t, outcome.Attempts, "nothing was attempted, so nothing should be counted")
}

// A shutdown must not be spent retrying. The attempts made are reported so the retry sweep can
// pick the order up on the next run.
func TestACanceledContextStopsRetrying(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	outcome := newTestClient(server, 5).Sync(ctx, ResultSyncRequest{
		ReturnUrl: server.URL,
		OrderId:   "ORDER1",
		Status:    models.OrderStatusPaymentSuccess,
	})

	assert.False(t, outcome.Succeeded())
	assert.LessOrEqual(t, atomic.LoadInt32(&calls), int32(1),
		"a canceled context must not keep calling the tenant")
}

// A tenant that is slow must not hold the sweep: the timeout bounds one attempt, and the sweep
// works through a page of orders.
func TestASlowTenantIsAbandonedAtTheTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	started := time.Now()
	outcome := NewResultSyncClient(50*time.Millisecond, 1).Sync(context.Background(), ResultSyncRequest{
		ReturnUrl: server.URL,
		OrderId:   "ORDER1",
		Status:    models.OrderStatusPaymentSuccess,
	})

	assert.False(t, outcome.Succeeded())
	assert.Less(t, time.Since(started), 250*time.Millisecond)
}

// A misconfigured client must not become one that never retries or has no timeout at all.
func TestDegenerateConfigurationFallsBackToSafeValues(t *testing.T) {
	client := NewResultSyncClient(0, 0)

	assert.Positive(t, client.httpClient.Timeout, "a zero timeout would wait forever")
	assert.GreaterOrEqual(t, client.maxRetries, 1, "zero retries would send nothing at all")
}
