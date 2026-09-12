package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The UoM rules now ask the consuming modules over the command bus instead of calling a probe
// registered into this package. These fixtures stand in for that conversation: a bus that
// answers the way one dependant would, so the rule can be tested without building four modules.

// usageReply is one dependant's scripted answer. A failure stands in for a module that could not
// answer at all, which the rules must treat as "in use" rather than as silence.
type usageReply struct {
	module  string
	used    bool
	failure string
}

type uomFakeBus struct {
	replies map[string]usageReply
	calls   int
}

func (this *uomFakeBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }
func (this *uomFakeBus) RequestNoReply(context.Context, cqrs.Request) error              { return nil }
func (this *uomFakeBus) IsRequestTypeRegistered(string) bool                             { return true }

func (this *uomFakeBus) Request(_ context.Context, request cqrs.Request, result any) error {
	this.calls++

	cmd, ok := request.(*usagecheck.CheckResourceUsageCommand)
	if !ok {
		return errors.New("unexpected request type")
	}
	reply, scripted := this.replies[cmd.TargetModule]
	if !scripted {
		return errors.Errorf("no handler for %s", cmd.TargetModule)
	}
	if reply.failure != "" {
		return errors.New(reply.failure)
	}

	items := make([]usagecheck.ResourceUsageItem, 0, len(cmd.Resources))
	for _, ref := range cmd.Resources {
		items = append(items, usagecheck.ResourceUsageItem{
			ResourceName: ref.ResourceName,
			Identifier:   ref.Identifier,
			IsUsed:       reply.used,
		})
	}
	*(result.(*usagecheck.CheckResourceUsageResult)) = usagecheck.CheckResourceUsageResult{
		RequestId: cmd.RequestId,
		Module:    cmd.TargetModule,
		Results:   items,
	}
	return nil
}

// essentialOwner declares the dependants the dispatcher will fan out to. Only the modules a test
// scripts an answer for are loaded, so an unscripted one cannot silently be skipped.
type essentialOwner struct {
	name       string
	dependants []string
}

func (this *essentialOwner) Name() string           { return this.name }
func (this *essentialOwner) Deps() []string         { return nil }
func (this *essentialOwner) LabelKey() string       { return this.name }
func (this *essentialOwner) Init() error            { return nil }
func (this *essentialOwner) IsInternal() bool       { return false }
func (this *essentialOwner) ModelPrefix() string    { return this.name }
func (this *essentialOwner) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }
func (this *essentialOwner) Dependants() []string   { return this.dependants }

// uomDispatcher builds a dispatcher whose dependants are exactly the modules scripted here.
func uomDispatcher(t *testing.T, replies ...usageReply) *usagecheck.Dispatcher {
	t.Helper()

	bus := &uomFakeBus{replies: map[string]usageReply{}}
	dependants := make([]string, 0, len(replies))
	for _, reply := range replies {
		bus.replies[reply.module] = reply
		dependants = append(dependants, reply.module)
	}

	loaded := []modules.InCodeModule{
		&essentialOwner{name: modconstants.EssentialModuleName, dependants: dependants},
	}
	for _, dependant := range dependants {
		loaded = append(loaded, &essentialOwner{name: dependant})
	}
	registry, err := modules.NewModuleDependantRegistry(loaded)
	require.NoError(t, err)

	return usagecheck.NewDispatcher(bus, registry, nil)
}

// dispatcherOn builds a dispatcher over a bus the test holds, for asserting whether it was used
// at all.
func dispatcherOn(t *testing.T, bus *uomFakeBus) *usagecheck.Dispatcher {
	t.Helper()

	if bus.replies == nil {
		bus.replies = map[string]usageReply{}
	}
	registry, err := modules.NewModuleDependantRegistry([]modules.InCodeModule{
		&essentialOwner{name: modconstants.EssentialModuleName, dependants: []string{"sales"}},
		&essentialOwner{name: "sales"},
	})
	require.NoError(t, err)

	return usagecheck.NewDispatcher(bus, registry, nil)
}

func uomTestCtx() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}

func newUomFixtureWithId(t *testing.T, id string) *models.Uom {
	t.Helper()

	uom := newUomFixture(t, models.UomTypeBiggerEqual, "1000", "0.01")
	uom.SetId(&id)
	return uom
}
