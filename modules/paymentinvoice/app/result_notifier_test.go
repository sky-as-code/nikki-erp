package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
	itEvent "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/event"
)

// Notifying is the step that turns a settled payment into goods the customer can take, and the
// recorded outcome is the only thing that gets a lost notification retried. These tests pin both.

// fakeOrderStore stands in for the order service, which otherwise needs a database.
type fakeOrderStore struct {
	orgId            string
	amount           int64
	method           string
	refTransactionId string
	metadata         map[string]any
	err              error

	// gate, when set, holds the notification at its first step until the test releases it. It is
	// what lets a test be sure the send happens after the request that spawned it has ended.
	gate chan struct{}

	mutex    sync.Mutex
	recorded []services.SyncOutcome
	pks      []string
}

func (this *fakeOrderStore) SyncFactsFor(_ corectx.Context, _ string) (*services.SyncFacts, error) {
	if this.gate != nil {
		<-this.gate
	}
	if this.err != nil {
		return nil, this.err
	}
	return &services.SyncFacts{
		OrgId:         this.orgId,
		Amount:        this.amount,
		PaymentMethod: this.method,
	}, nil
}

func (this *fakeOrderStore) SettlementFactsFor(
	_ corectx.Context, _ string,
) (*services.SettlementFacts, error) {
	if this.err != nil {
		return nil, this.err
	}
	return &services.SettlementFacts{
		OrgId:            this.orgId,
		Amount:           decimal.NewFromInt(this.amount),
		RefTransactionId: this.refTransactionId,
		Metadata:         this.metadata,
	}, nil
}

func (this *fakeOrderStore) RecordSyncOutcome(
	_ corectx.Context, orderPk string, outcome services.SyncOutcome,
) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.recorded = append(this.recorded, outcome)
	this.pks = append(this.pks, orderPk)
	return nil
}

func (this *fakeOrderStore) outcomes() []services.SyncOutcome {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return append([]services.SyncOutcome(nil), this.recorded...)
}

// fakeSettledPublisher records what was announced, so a test can assert on the in-process half of
// a settlement as well as the HTTP half.
type fakeSettledPublisher struct {
	mutex     sync.Mutex
	published []itEvent.PaymentSettledEvent
}

func (this *fakeSettledPublisher) PublishAsync(
	_ corectx.Context, event itEvent.PaymentSettledEvent,
) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.published = append(this.published, event)
}

func (this *fakeSettledPublisher) events() []itEvent.PaymentSettledEvent {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return append([]itEvent.PaymentSettledEvent(nil), this.published...)
}

func newTestNotifier(store orderSyncStore, retries int) *ResultNotifier {
	notifier := NewResultNotifier(
		nil, NewResultSyncClient(2*time.Second, retries), nil, logging.NewLogger())
	notifier.orders = store
	return notifier
}

func newAnnouncingNotifier(
	store orderSyncStore, publisher itEvent.PaymentSettledEventPublisher,
) *ResultNotifier {
	notifier := NewResultNotifier(
		nil, NewResultSyncClient(2*time.Second, 0), publisher, logging.NewLogger())
	notifier.orders = store
	return notifier
}

func testContext() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

// The happy path: the machine is called and the success is written down, so the retry sweep leaves
// the order alone.
//
// The organization is carried too. It is not on NotifyTarget — it comes off the order when its
// facts are read — so a notification that reached the ordering system without one would look like
// the caller's mistake rather than a missing read here.
func TestNotifyPostsTheResultAndRecordsSuccess(t *testing.T) {
	var gotContentType string
	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := &fakeOrderStore{orgId: "01JBQ0000000000000000ORG", amount: 150000, method: "momo"}
	newTestNotifier(store, 1).Notify(testContext(), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "VDMCMOM0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	assert.Equal(t, "application/json", gotContentType)
	assert.Equal(t, "01JBQ0000000000000000ORG", received["org_id"])
	assert.Equal(t, "VDMCMOM0Q8HABCDEFGH", received["orderId"])

	recorded := store.outcomes()
	require.Len(t, recorded, 1)
	assert.Equal(t, services.SyncStatusSuccess, recorded[0].Status)
	assert.Equal(t, []string{"order-pk"}, store.pks)
}

// A tenant that is down must leave a failure behind. Without it the retry sweep has nothing to
// find, and the only notice of a completed payment is lost for good.
func TestNotifyRecordsAFailureWhenTheTenantIsDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	store := &fakeOrderStore{amount: 1000, method: "vietqr"}
	newTestNotifier(store, 1).Notify(testContext(), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "VDMCVQR0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	recorded := store.outcomes()
	require.Len(t, recorded, 1)
	assert.Equal(t, services.SyncStatusFailure, recorded[0].Status)
	assert.NotEmpty(t, recorded[0].Detail, "a failure carries its reason for whoever reconciles it")
}

// An order whose facts cannot be read is not reported as a failed notification: nothing was sent,
// and recording a failure would have the sweep chase an order it cannot read either.
func TestNotifyRecordsNothingWhenTheOrderCannotBeRead(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := &fakeOrderStore{err: assert.AnError}
	newTestNotifier(store, 1).Notify(testContext(), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "VDMCMOM0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	assert.False(t, called)
	assert.Empty(t, store.outcomes())
}

// The callbacks detach the send so the gateway is answered immediately. It still has to arrive:
// detached must mean "not on the caller's thread", never "best effort".
func TestNotifyDetachedStillDelivers(t *testing.T) {
	arrived := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		arrived <- struct{}{}
	}))
	defer server.Close()

	store := &fakeOrderStore{amount: 500, method: "mpos"}
	newTestNotifier(store, 1).NotifyDetached(testContext(), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "VDMCMPO0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	select {
	case <-arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("the detached notification never reached the ordering system")
	}
}

// The request that spawned a detached send is over by the time it runs. Reading the request's own
// context would cancel it the moment the gateway is answered — every notification lost.
func TestNotifyDetachedSurvivesTheRequestEnding(t *testing.T) {
	arrived := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		arrived <- struct{}{}
	}))
	defer server.Close()

	// The send is held until the request has definitely ended, so this pins the behaviour rather
	// than racing the goroutine against the cancel.
	requestCtx, endRequest := context.WithCancel(context.Background())
	store := &fakeOrderStore{amount: 500, method: "mpos", gate: make(chan struct{})}
	newTestNotifier(store, 1).NotifyDetached(corectx.NewRequestContext(requestCtx), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "VDMCMPO0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)
	endRequest()
	close(store.gate)

	select {
	case <-arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("the notification was canceled along with the request that spawned it")
	}
}

// Announcing a settled order.
//
// The in-process announcement and the HTTP sync are two paths to the same news, and these pin the
// half that other modules in this build listen to.

// A settled order reaches subscribers with the facts they correlate on: the exact amount, the
// gateway's reference, and whatever the opening caller attached.
func TestNotifyAnnouncesASettledOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	publisher := &fakeSettledPublisher{}
	store := &fakeOrderStore{
		orgId:            "org-1",
		amount:           1500,
		method:           "momo",
		refTransactionId: "gw-txn-9",
		metadata:         map[string]any{"sales_payment_id": "pay-1"},
	}

	newAnnouncingNotifier(store, publisher).Notify(testContext(), NotifyTarget{
		Pk:        "order-pk",
		OrderId:   "SALEMOM0Q8HABCDEFGH",
		ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	events := publisher.events()
	require.Len(t, events, 1, "a settled order must be announced exactly once")

	announced := events[0]
	assert.Equal(t, itEvent.PaymentSettledPaid, announced.Type)
	assert.Equal(t, "SALEMOM0Q8HABCDEFGH", announced.OrderId)
	assert.Equal(t, "order-pk", announced.OrderPk)
	assert.Equal(t, "org-1", announced.OrgId)
	assert.Equal(t, "gw-txn-9", announced.RefTransactionId)

	// A string, not a number: money that has been through a float is no longer the money taken.
	assert.Equal(t, "1500", announced.Amount)

	// The caller's own correlation, echoed back untouched.
	assert.Equal(t, "pay-1", announced.Metadata["sales_payment_id"])
}

// Every terminal verdict is announced, each as its own type: a subscriber has to tell a customer
// who paid from one who never did.
func TestNotifyAnnouncesEveryTerminalVerdict(t *testing.T) {
	cases := []struct {
		status string
		want   itEvent.PaymentSettledType
	}{
		{models.OrderStatusPaymentSuccess, itEvent.PaymentSettledPaid},
		{models.OrderStatusPaymentFailed, itEvent.PaymentSettledFailed},
		{models.OrderStatusExpired, itEvent.PaymentSettledExpired},
		{models.OrderStatusCanceled, itEvent.PaymentSettledCanceled},
	}

	for _, testCase := range cases {
		t.Run(testCase.status, func(t *testing.T) {
			publisher := &fakeSettledPublisher{}
			store := &fakeOrderStore{orgId: "org-1", amount: 100}

			// No ReturnUrl: nothing asked for an HTTP callback, and the announcement must happen
			// anyway — the two are independent.
			newAnnouncingNotifier(store, publisher).Notify(testContext(), NotifyTarget{
				Pk: "order-pk", OrderId: "SALEMOM0Q8HABCDEFGH",
			}, testCase.status)

			events := publisher.events()
			require.Len(t, events, 1)
			assert.Equal(t, testCase.want, events[0].Type)
		})
	}
}

// A status that is not a verdict wakes nobody. Announcing "processing" would rouse every subscriber
// for an order the customer is still in the middle of paying.
func TestNotifyDoesNotAnnounceANonVerdict(t *testing.T) {
	publisher := &fakeSettledPublisher{}
	store := &fakeOrderStore{orgId: "org-1", amount: 100}

	newAnnouncingNotifier(store, publisher).Notify(testContext(), NotifyTarget{
		Pk: "order-pk", OrderId: "SALEMOM0Q8HABCDEFGH",
	}, models.OrderStatusProcessing)

	assert.Empty(t, publisher.events(), "only a terminal verdict is announced")
}

// A build with no publisher wired still settles orders. The announcement is an optimization over
// the order row, never a requirement for the money to be recorded.
func TestNotifyWithoutAPublisherStillRecordsTheOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := &fakeOrderStore{orgId: "org-1", amount: 100}
	newTestNotifier(store, 1).Notify(testContext(), NotifyTarget{
		Pk: "order-pk", OrderId: "SALEMOM0Q8HABCDEFGH", ReturnUrl: server.URL,
	}, models.OrderStatusPaymentSuccess)

	assert.Len(t, store.outcomes(), 1, "the sync outcome is recorded with or without a publisher")
}
