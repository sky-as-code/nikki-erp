package jobscheduler

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"

	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/app"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/external"
	repo "github.com/sky-as-code/nikki-erp/modules/jobscheduler/infra/repository"
)

// The dig container is process-wide, so registration happens once for the whole package. A
// second Init would fail with "already provided" rather than telling us anything about the
// wiring.
var registerOnce sync.Once

func registerModuleDeps(t *testing.T) {
	t.Helper()
	var err error
	registerOnce.Do(func() {
		// cqrs.CqrsBus and *client.HttpClient come from core in a running application. Standing
		// in for them here is what lets the rest of the graph be exercised for real, rather than
		// stopping at the first dependency this module does not own.
		if err = deps.Register(
			func() cqrs.CqrsBus { return &stubCqrsBus{} },
			func() *client.HttpClient { return &client.HttpClient{Client: http.Client{}} },
		); err != nil {
			return
		}
		for _, init := range []func() error{
			external.InitExternalServices,
			repo.InitRepositories,
			services.InitDomainServices,
			app.InitApplicationServices,
		} {
			if err = init(); err != nil {
				return
			}
		}
	})
	require.NoError(t, err)
}

// TestEveryRegistrationSucceeds proves each layer's Init registers cleanly.
//
// It catches the mistake dig cannot: registering two constructors for the same type, which fails
// at registration rather than at resolution and would otherwise only appear at boot.
//
// It stops short of resolving the repositories and the transport. Those need a database client,
// a config service and the HTTP server's *echo.Group, all provided by core in a running
// application; standing them up here would make this a test of the platform rather than of this
// module. What the module itself wires - the executors, the dispatcher and the waker - is
// resolved for real in the two tests below.
func TestEveryRegistrationSucceeds(t *testing.T) {
	registerModuleDeps(t)
}

// The domain services depend on EngineWaker, while OnAppStarted needs the concrete
// *DeferredWaker so it can attach the engine. Both must resolve to the same object: if the
// interface registration built a second waker, a job created over REST would wake a waker no
// engine was ever attached to, and the horizon would not move until the next reconciliation -
// a scheduler that looks like it works but is always up to a minute late.
func TestWakerInterfaceAndConcreteTypeAreTheSameInstance(t *testing.T) {
	registerModuleDeps(t)

	var concrete *services.DeferredWaker
	var asInterface services.EngineWaker

	require.NoError(t, deps.Invoke(func(c *services.DeferredWaker, i services.EngineWaker) {
		concrete, asInterface = c, i
	}))

	require.Same(t, concrete, asInterface,
		"the interface registration must return the singleton, not a second waker")
}

// The dispatcher is registered under both its own type and the domain's validator interface, for
// the same reason and with the same failure mode as the waker.
func TestDispatcherSatisfiesTheDomainValidatorPort(t *testing.T) {
	registerModuleDeps(t)

	var concrete *external.ActionDispatcher
	var asInterface services.ActionConfigValidator

	require.NoError(t, deps.Invoke(func(
		c *external.ActionDispatcher, i services.ActionConfigValidator,
	) {
		concrete, asInterface = c, i
	}))

	require.Same(t, concrete, asInterface)
}

// stubCqrsBus stands in for the bus. Nothing here is called: the graph only needs something of
// the right type to hand the command executor.
type stubCqrsBus struct{}

func (*stubCqrsBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }

func (*stubCqrsBus) RequestNoReply(context.Context, cqrs.Request) error { return nil }

func (*stubCqrsBus) Request(context.Context, cqrs.Request, any) error { return nil }

func (*stubCqrsBus) IsRequestTypeRegistered(string) bool { return false }
