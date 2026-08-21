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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/services"
)

// Notifying is the step that turns a settled payment into goods the customer can take, and the
// recorded outcome is the only thing that gets a lost notification retried. These tests pin both.

// fakeOrderStore stands in for the order service, which otherwise needs a database.
type fakeOrderStore struct {
	orgId  string
	amount int64
	method string
	err    error

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

func newTestNotifier(store orderSyncStore, retries int) *ResultNotifier {
	notifier := NewResultNotifier(nil, NewResultSyncClient(2*time.Second, retries), logging.NewLogger())
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
