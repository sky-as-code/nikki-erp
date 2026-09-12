package usagecheck

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

// answerFn produces one dependant's reply, or an error standing in for a handler that failed,
// panicked or never answered.
type answerFn func(cmd *CheckResourceUsageCommand) (CheckResourceUsageResult, error)

// fakeBus routes a request to the answer registered for its target module, and records how many
// calls were in flight at once so the concurrency claim can be asserted rather than assumed.
type fakeBus struct {
	mu         sync.Mutex
	answers    map[string]answerFn
	registered map[string]bool

	inFlight    int32
	maxInFlight int32
}

func newFakeBus() *fakeBus {
	return &fakeBus{answers: map[string]answerFn{}, registered: map[string]bool{}}
}

func (this *fakeBus) answer(module string, fn answerFn) *fakeBus {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.answers[module] = fn
	this.registered[RequestTypeFor(module).String()] = true
	return this
}

func (this *fakeBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }
func (this *fakeBus) RequestNoReply(context.Context, cqrs.Request) error              { return nil }

func (this *fakeBus) IsRequestTypeRegistered(requestType string) bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.registered[requestType]
}

func (this *fakeBus) Request(ctx context.Context, request cqrs.Request, result any) error {
	current := atomic.AddInt32(&this.inFlight, 1)
	for {
		peak := atomic.LoadInt32(&this.maxInFlight)
		if current <= peak || atomic.CompareAndSwapInt32(&this.maxInFlight, peak, current) {
			break
		}
	}
	defer atomic.AddInt32(&this.inFlight, -1)

	cmd, ok := request.(*CheckResourceUsageCommand)
	if !ok {
		return errors.New("unexpected request type")
	}

	this.mu.Lock()
	fn, found := this.answers[cmd.TargetModule]
	this.mu.Unlock()
	if !found {
		return errors.Errorf("no handler for %s", cmd.TargetModule)
	}

	reply, err := fn(cmd)
	if err != nil {
		return err
	}
	*(result.(*CheckResourceUsageResult)) = reply
	return nil
}

// answers builds a reply marking every requested resource used or unused.
func answers(module string, used bool) answerFn {
	return func(cmd *CheckResourceUsageCommand) (CheckResourceUsageResult, error) {
		items := make([]ResourceUsageItem, 0, len(cmd.Resources))
		for _, ref := range cmd.Resources {
			items = append(items, ResourceUsageItem{
				ResourceName: ref.ResourceName,
				Identifier:   ref.Identifier,
				IsUsed:       used,
			})
		}
		return CheckResourceUsageResult{RequestId: cmd.RequestId, Module: module, Results: items}, nil
	}
}

func fails(message string) answerFn {
	return func(*CheckResourceUsageCommand) (CheckResourceUsageResult, error) {
		return CheckResourceUsageResult{}, errors.New(message)
	}
}

type testModule struct {
	name       string
	dependants []string
}

func (this *testModule) Name() string           { return this.name }
func (this *testModule) Deps() []string         { return nil }
func (this *testModule) LabelKey() string       { return this.name }
func (this *testModule) Init() error            { return nil }
func (this *testModule) IsInternal() bool       { return false }
func (this *testModule) ModelPrefix() string    { return this.name }
func (this *testModule) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }
func (this *testModule) Dependants() []string   { return this.dependants }

func registryWith(t *testing.T, owner string, dependants ...string) *modules.ModuleDependantRegistry {
	t.Helper()
	loaded := []modules.InCodeModule{&testModule{name: owner, dependants: dependants}}
	for _, dependant := range dependants {
		loaded = append(loaded, &testModule{name: dependant})
	}
	registry, err := modules.NewModuleDependantRegistry(loaded)
	require.NoError(t, err)
	return registry
}

func productRef() ResourceRef {
	return ResourceRef{
		ResourceName: "product_variant",
		Identifier:   map[string]string{"id": "VAR-1", "org_id": "ORG-1"},
	}
}

func testCtx() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

func TestCheckUsage_AllUnusedPermitsDelete(t *testing.T) {
	bus := newFakeBus().answer("sales", answers("sales", false)).answer("purchase", answers("purchase", false))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.True(t, outcome.MayDelete())
	assert.Empty(t, outcome.BlockingModules())
}

func TestCheckUsage_OneUsedBlocksDelete(t *testing.T) {
	bus := newFakeBus().answer("sales", answers("sales", true)).answer("purchase", answers("purchase", false))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.False(t, outcome.MayDelete())
	assert.Equal(t, []string{"sales"}, outcome.BlockingModules())
}

// §15: an error is not a negative answer. A dependant that cannot answer must block the delete,
// because the alternative is deleting a resource something may still reference.
func TestCheckUsage_ErrorBlocksDelete(t *testing.T) {
	bus := newFakeBus().answer("sales", fails("database is down")).answer("purchase", answers("purchase", false))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err, "a refusal is an outcome, not a dispatch failure")
	assert.False(t, outcome.MayDelete())
	assert.Equal(t, []string{"sales"}, outcome.BlockingModules())
	assert.ErrorContains(t, outcome.FirstError(), "database is down")
}

// A dependant with no subscribed handler reaches the bus as "no handler", which must block for
// the same reason an error does.
func TestCheckUsage_MissingHandlerBlocksDelete(t *testing.T) {
	bus := newFakeBus().answer("purchase", answers("purchase", false))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.False(t, outcome.MayDelete())
	assert.Equal(t, []string{"sales"}, outcome.BlockingModules())
}

// A short reply leaves resources unanswered, and an unanswered resource must not read as unused.
func TestCheckUsage_ShortReplyBlocksDelete(t *testing.T) {
	bus := newFakeBus().answer("sales", func(cmd *CheckResourceUsageCommand) (CheckResourceUsageResult, error) {
		return CheckResourceUsageResult{RequestId: cmd.RequestId, Module: "sales"}, nil
	})
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.False(t, outcome.MayDelete())
	assert.ErrorContains(t, outcome.FirstError(), "answered 0 of 1")
}

// Every blocking module is reported, not just the first to answer, so one round trip tells the
// user everything they must clear.
func TestCheckUsage_ReportsEveryBlockingModule(t *testing.T) {
	bus := newFakeBus().
		answer("sales", answers("sales", true)).
		answer("purchase", fails("timeout")).
		answer("vendingmachine", answers("vendingmachine", false))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase", "vendingmachine"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.Equal(t, []string{"purchase", "sales"}, outcome.BlockingModules())
}

// BR-DEL-001: no dependants means nothing to ask, and no bus traffic at all.
func TestCheckUsage_NoDependantsPermitsDeleteWithoutDispatch(t *testing.T) {
	bus := newFakeBus()
	dispatcher := NewDispatcher(bus, registryWith(t, "essential"), nil)

	outcome, err := dispatcher.CheckUsage(testCtx(), "essential", []ResourceRef{productRef()})

	require.NoError(t, err)
	assert.True(t, outcome.MayDelete())
	assert.Equal(t, int32(0), bus.maxInFlight, "with nobody to ask, nothing is sent")
}

// §13: the dependants are asked in parallel. Sequential dispatch would cost the sum of the
// latencies on every delete, and the answers are independent.
func TestCheckUsage_DispatchesConcurrently(t *testing.T) {
	const delay = 80 * time.Millisecond
	slow := func(module string) answerFn {
		return func(cmd *CheckResourceUsageCommand) (CheckResourceUsageResult, error) {
			time.Sleep(delay)
			return answers(module, false)(cmd)
		}
	}
	bus := newFakeBus().
		answer("sales", slow("sales")).
		answer("purchase", slow("purchase")).
		answer("vendingmachine", slow("vendingmachine"))
	dispatcher := NewDispatcher(bus, registryWith(t, "inventory", "sales", "purchase", "vendingmachine"), nil)

	started := time.Now()
	outcome, err := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})
	elapsed := time.Since(started)

	require.NoError(t, err)
	assert.True(t, outcome.MayDelete())
	assert.Equal(t, int32(3), bus.maxInFlight, "all three were in flight at once")
	assert.Less(t, elapsed, 3*delay, "sequential dispatch would take at least the sum")
}

// An absent dependant - vendingmachine in the nikkierp binary - is skipped rather than dispatched
// to, so it cannot become a permanent refusal. See docs/problems/inventory/001.
func TestCheckUsage_SkipsDependantAbsentFromThisBinary(t *testing.T) {
	registry, err := modules.NewModuleDependantRegistry([]modules.InCodeModule{
		&testModule{name: "inventory", dependants: []string{"sales", "vendingmachine"}},
		&testModule{name: "sales"},
	})
	require.NoError(t, err)

	bus := newFakeBus().answer("sales", answers("sales", false))
	dispatcher := NewDispatcher(bus, registry, nil)

	outcome, dispatchErr := dispatcher.CheckUsage(testCtx(), "inventory", []ResourceRef{productRef()})

	require.NoError(t, dispatchErr)
	assert.True(t, outcome.MayDelete(), "the absent module is not asked, so it cannot block")
	assert.Len(t, outcome.Answers, 1)
}

func TestCheckUsage_RejectsEmptyResourceList(t *testing.T) {
	dispatcher := NewDispatcher(newFakeBus(), registryWith(t, "inventory", "sales"), nil)

	_, err := dispatcher.CheckUsage(testCtx(), "inventory", nil)

	assert.Error(t, err)
}

func TestAssertDependantsSubscribed(t *testing.T) {
	registry := registryWith(t, "inventory", "sales", "purchase")

	subscribed := newFakeBus().answer("sales", answers("sales", false)).
		answer("purchase", answers("purchase", false))
	assert.NoError(t, AssertDependantsSubscribed(subscribed, registry, "inventory"))

	partial := newFakeBus().answer("sales", answers("sales", false))
	err := AssertDependantsSubscribed(partial, registry, "inventory")
	require.Error(t, err)
	assert.ErrorContains(t, err, "purchase")
}
