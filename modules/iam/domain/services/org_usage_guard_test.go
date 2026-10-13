package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
)

const testDependant = "parkinglot"

// orgUsageBus answers every usage check the way the one scripted dependant would.
type orgUsageBus struct {
	used    bool
	failure string
}

func (this *orgUsageBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error { return nil }
func (this *orgUsageBus) RequestNoReply(context.Context, cqrs.Request) error              { return nil }
func (this *orgUsageBus) IsRequestTypeRegistered(string) bool                             { return true }

func (this *orgUsageBus) Request(_ context.Context, request cqrs.Request, result any) error {
	cmd, ok := request.(*usagecheck.CheckResourceUsageCommand)
	if !ok {
		return errors.New("unexpected request type")
	}
	if this.failure != "" {
		return errors.New(this.failure)
	}
	items := make([]usagecheck.ResourceUsageItem, 0, len(cmd.Resources))
	for _, ref := range cmd.Resources {
		items = append(items, usagecheck.ResourceUsageItem{
			ResourceName: ref.ResourceName, Identifier: ref.Identifier, IsUsed: this.used,
		})
	}
	*(result.(*usagecheck.CheckResourceUsageResult)) = usagecheck.CheckResourceUsageResult{
		RequestId: cmd.RequestId, Module: cmd.TargetModule, Results: items,
	}
	return nil
}

type orgTestModule struct {
	name       string
	dependants []string
}

func (this *orgTestModule) Name() string           { return this.name }
func (this *orgTestModule) Deps() []string         { return nil }
func (this *orgTestModule) LabelKey() string       { return this.name }
func (this *orgTestModule) Init() error            { return nil }
func (this *orgTestModule) IsInternal() bool       { return false }
func (this *orgTestModule) ModelPrefix() string    { return this.name }
func (this *orgTestModule) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }
func (this *orgTestModule) Dependants() []string   { return this.dependants }

func orgDispatcher(t *testing.T, bus *orgUsageBus) *usagecheck.Dispatcher {
	t.Helper()
	registry, err := modules.NewModuleDependantRegistry([]modules.InCodeModule{
		&orgTestModule{name: c.IamModuleName, dependants: []string{testDependant}},
		&orgTestModule{name: testDependant},
	})
	require.NoError(t, err)
	return usagecheck.NewDispatcher(bus, registry, nil)
}

// deleteRecorder is the engine's own resource service; only Delete is reached by these tests.
type deleteRecorder struct {
	drif.DynamicResourceService
	deleted int
}

func (this *deleteRecorder) Delete(corectx.Context, dmodel.DynamicFields) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.deleted++
	return &dyn.OpResult[dyn.MutateResultData]{HasData: true}, nil
}

func deleteOrgThroughEngine(t *testing.T, bus *orgUsageBus) (*dyn.OpResult[dyn.MutateResultData], *deleteRecorder, error) {
	t.Helper()
	base := &deleteRecorder{}
	guard := NewOrgEngineDeleteGuard(base, orgDispatcher(t, bus))
	ctx := corectx.NewRequestContext(context.Background())
	result, err := guard.Delete(ctx, dmodel.DynamicFields{basemodel.FieldId: "01JORG0000000000000000000A"})
	return result, base, err
}

func TestOrgEngineDeleteGuard_UnusedOrgIsDeleted(t *testing.T) {
	result, base, err := deleteOrgThroughEngine(t, &orgUsageBus{used: false})

	require.NoError(t, err)
	assert.Equal(t, 0, result.ClientErrors.Count())
	assert.Equal(t, 1, base.deleted)
}

func TestOrgEngineDeleteGuard_UsedOrgIsRefused(t *testing.T) {
	result, base, err := deleteOrgThroughEngine(t, &orgUsageBus{used: true})

	require.NoError(t, err)
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, usagecheck.ErrorKeyResourceInUse, result.ClientErrors[0].Key)
	assert.Equal(t, 0, base.deleted)
}

func TestOrgEngineDeleteGuard_UnansweredCheckRefuses(t *testing.T) {
	result, base, err := deleteOrgThroughEngine(t, &orgUsageBus{failure: "parkinglot is down"})

	require.NoError(t, err)
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, 0, base.deleted)
}
