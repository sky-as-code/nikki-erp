package orglifecycle

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

type answerFn func(ctx context.Context, cmd *OrgLifecycleCommand) error

// fakeBus answers a module's command with the function registered for it; a module with no
// function is unsubscribed.
type fakeBus struct {
	mutex   sync.Mutex
	answers map[string]answerFn
	asked   []string
}

func newFakeBus() *fakeBus {
	return &fakeBus{answers: map[string]answerFn{}}
}

func (this *fakeBus) answer(module string, fn answerFn) *fakeBus {
	for _, action := range []string{ActionInitOrg, ActionCleanUpOrg} {
		this.answers[RequestTypeFor(module, action).String()] = fn
	}
	return this
}

func (this *fakeBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }
func (this *fakeBus) RequestNoReply(context.Context, cqrs.Request) error              { return nil }

func (this *fakeBus) IsRequestTypeRegistered(requestType string) bool {
	_, found := this.answers[requestType]
	return found
}

func (this *fakeBus) Request(ctx context.Context, request cqrs.Request, _ any) error {
	cmd := request.(*OrgLifecycleCommand)
	this.mutex.Lock()
	this.asked = append(this.asked, cmd.TargetModule)
	this.mutex.Unlock()
	return this.answers[cmd.CqrsRequestType().String()](ctx, cmd)
}

func ok(context.Context, *OrgLifecycleCommand) error { return nil }

func dispatch(t *testing.T, bus *fakeBus, waves [][]string) *Outcome {
	t.Helper()
	dispatcher := NewDispatcher(bus, nil).WithTimeout(200 * time.Millisecond)
	outcome, err := dispatcher.Dispatch(
		corectx.NewRequestContext(context.Background()), ActionInitOrg, waves, "ORG-1", "USR-1")
	require.NoError(t, err)
	return outcome
}

func TestDispatch_AllWavesSucceed(t *testing.T) {
	bus := newFakeBus().answer("essential", ok).answer("inventory", ok).answer("sales", ok)

	outcome := dispatch(t, bus, [][]string{{"essential"}, {"inventory", "sales"}})

	assert.True(t, outcome.Succeeded())
	assert.Equal(t, "essential", bus.asked[0], "a lower wave is asked first")
	assert.Len(t, bus.asked, 3)
}

func TestDispatch_FailedWaveSkipsTheRest(t *testing.T) {
	bus := newFakeBus().
		answer("essential", ok).
		answer("inventory", func(context.Context, *OrgLifecycleCommand) error { return errors.New("boom") }).
		answer("sales", ok)

	outcome := dispatch(t, bus, [][]string{{"essential", "inventory"}, {"sales"}})

	assert.False(t, outcome.Succeeded())
	assert.Equal(t, []string{"inventory"}, outcome.FailedModules())
	assert.Equal(t, []string{"sales"}, outcome.SkippedModules())
	assert.NotContains(t, bus.asked, "sales")
}

func TestDispatch_UnsubscribedTimeoutAndPanicAreFailures(t *testing.T) {
	bus := newFakeBus().
		answer("slow", func(ctx context.Context, _ *OrgLifecycleCommand) error {
			<-ctx.Done()
			return ctx.Err()
		}).
		answer("crashing", func(context.Context, *OrgLifecycleCommand) error { panic("nil map") })

	outcome := dispatch(t, bus, [][]string{{"slow", "crashing", "unsubscribed"}})

	assert.Equal(t, []string{"crashing", "slow", "unsubscribed"}, outcome.FailedModules())
}

func TestDispatch_CommandCarriesOrgAndRequester(t *testing.T) {
	var got OrgLifecycleCommand
	bus := newFakeBus().answer("sales", func(_ context.Context, cmd *OrgLifecycleCommand) error {
		got = *cmd
		return nil
	})

	dispatch(t, bus, [][]string{{"sales"}})

	assert.Equal(t, "ORG-1", got.OrgId)
	assert.Equal(t, "USR-1", got.RequestedBy)
	assert.Equal(t, ActionInitOrg, got.Action)
}

func TestReverseWaves(t *testing.T) {
	assert.Equal(t, [][]string{{"c"}, {"b"}, {"a"}}, ReverseWaves([][]string{{"a"}, {"b"}, {"c"}}))
}
