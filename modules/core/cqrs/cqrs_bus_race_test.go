package cqrs

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

// These tests guard core-001 (docs/problems/core/001-cqrs-reply-context-race.md). They only prove
// anything under the race detector: `go test -race ./modules/core/cqrs/`.

type echoRequest struct {
	Delay time.Duration `json:"delay"`
	Value int           `json:"value"`
}

func (this echoRequest) CqrsRequestType() RequestType {
	return RequestType{Module: "test", Submodule: "race", Action: "echo"}
}

type echoResult struct {
	Value int `json:"value"`
}

type raceCtxKey struct{}

func newTestBus(t *testing.T, maxTimeout time.Duration) *WatermillCqrsBus {
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermill.NopLogger{})
	bus := &WatermillCqrsBus{
		logger:     logging.NewLogger(),
		publisher:  pubSub,
		subscriber: pubSub,
		marshaler:  cqrs.JSONMarshaler{GenerateName: cqrs.NamedStruct(cqrs.StructName)},
		maxTimeout: maxTimeout,
	}
	handler := NewHandler(func(ctx context.Context, packet *RequestPacket[echoRequest]) (*Reply[echoResult], error) {
		req := packet.Request()
		time.Sleep(req.Delay)
		return &Reply[echoResult]{Result: echoResult{Value: req.Value}}, nil
	})
	subCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancel()
		_ = pubSub.Close()
	})
	require.NoError(t, bus.SubscribeRequests(subCtx, handler))
	return bus
}

// requestThenMutate is what ResolveLocale and ResolveEnabledLanguages do: send a request on the
// caller's context, then memoize the answer on that same context with WithValue. Before the fix,
// the reply goroutine's deferred cancel was still reading the context's inner field while
// WithValue rewrote it, which tore the interface value and crashed the process.
func requestThenMutate(bus *WatermillCqrsBus, parent context.Context, req echoRequest) error {
	ctx := corectx.NewRequestContext(parent)
	var result echoResult
	err := bus.Request(ctx, req, &result)
	for i := 0; i < 50; i++ {
		ctx.WithValue(raceCtxKey{}, i)
		ctx.SetDomainConstraints(nil)
	}
	if err == nil && result.Value != req.Value {
		return assert.AnError
	}
	return err
}

func TestRequestDoesNotShareTheCallersMutableContext(t *testing.T) {
	bus := newTestBus(t, 5*time.Second)
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			errs <- requestThenMutate(bus, parent, echoRequest{Value: value})
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		assert.NoError(t, err)
	}
}

// A request that times out returns before its handler replies. The caller then mutates its
// context, and the late reply must neither block the bus nor touch the caller.
func TestTimedOutRequestDoesNotShareTheCallersMutableContext(t *testing.T) {
	bus := newTestBus(t, 200*time.Millisecond)
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	// The bus serves one request type serially, so a few slow requests suffice to time all out.
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			err := requestThenMutate(bus, parent, echoRequest{Value: value, Delay: 300 * time.Millisecond})
			assert.ErrorContains(t, err, "timeout for request")
		}(i)
	}
	wg.Wait()
	// Late replies land on reply topics nobody listens to any more; the bus must still serve.
	time.Sleep(time.Second)
	assert.NoError(t, requestThenMutate(bus, parent, echoRequest{Value: 7}))
}

// A caller whose context is cancelled gets the cancellation back rather than hanging.
func TestCancelledCallerContextEndsTheRequest(t *testing.T) {
	bus := newTestBus(t, 5*time.Second)
	parent, cancel := context.WithCancel(context.Background())
	time.AfterFunc(20*time.Millisecond, cancel)

	err := requestThenMutate(bus, parent, echoRequest{Value: 1, Delay: 200 * time.Millisecond})

	assert.ErrorIs(t, err, context.Canceled)
}
