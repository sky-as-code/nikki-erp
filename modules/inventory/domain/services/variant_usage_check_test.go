package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// Deleting a variant asks Sales, Purchase and the vending machine first. The foreign keys are
// ON DELETE RESTRICT, so the database refuses the statement anyway — but only as a driver error
// naming a constraint. These tests pin the business answer: a refusal that says which module to
// go and clear.

type variantUsageReply struct {
	module  string
	used    bool
	failure string
}

type variantFakeBus struct {
	replies map[string]variantUsageReply
	calls   int
}

func (this *variantFakeBus) SubscribeRequests(context.Context, ...cqrs.RequestHandler) error {
	return nil
}
func (this *variantFakeBus) RequestNoReply(context.Context, cqrs.Request) error { return nil }
func (this *variantFakeBus) IsRequestTypeRegistered(string) bool                { return true }

func (this *variantFakeBus) Request(_ context.Context, request cqrs.Request, result any) error {
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

type variantOwnerModule struct {
	name       string
	dependants []string
}

func (this *variantOwnerModule) Name() string           { return this.name }
func (this *variantOwnerModule) Deps() []string         { return nil }
func (this *variantOwnerModule) LabelKey() string       { return this.name }
func (this *variantOwnerModule) Init() error            { return nil }
func (this *variantOwnerModule) IsInternal() bool       { return false }
func (this *variantOwnerModule) ModelPrefix() string    { return this.name }
func (this *variantOwnerModule) Version() semver.SemVer { return *semver.MustParseSemVer("v1.0.0") }
func (this *variantOwnerModule) Dependants() []string   { return this.dependants }

func variantDispatcher(t *testing.T, replies ...variantUsageReply) *usagecheck.Dispatcher {
	t.Helper()

	bus := &variantFakeBus{replies: map[string]variantUsageReply{}}
	dependants := make([]string, 0, len(replies))
	for _, reply := range replies {
		bus.replies[reply.module] = reply
		dependants = append(dependants, reply.module)
	}

	loaded := []modules.InCodeModule{
		&variantOwnerModule{name: modconstants.InventoryModuleName, dependants: dependants},
	}
	for _, dependant := range dependants {
		loaded = append(loaded, &variantOwnerModule{name: dependant})
	}
	registry, err := modules.NewModuleDependantRegistry(loaded)
	require.NoError(t, err)

	return usagecheck.NewDispatcher(bus, registry, nil)
}

func variantParams() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.ProductVariantFieldId: "VAR-1",
		"org_id":                     "ORG-1",
	}
}

func TestAssertVariantDeletableBlocksAVariantInUse(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertVariantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t, variantUsageReply{module: "sales", used: true}),
		variantParams(), vErrs)

	require.NoError(t, err)
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, usagecheck.ErrorKeyResourceInUse, (*vErrs)[0].Key)
	assert.Contains(t, (*vErrs)[0].Message, "sales", "the refusal names where to clear the reference")
}

func TestAssertVariantDeletablePermitsAnUnusedVariant(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertVariantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t,
			variantUsageReply{module: "sales", used: false},
			variantUsageReply{module: "purchase", used: false},
			variantUsageReply{module: "vendingmachine", used: false}),
		variantParams(), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

// The kiosk case: a variant nobody has sold but a machine is currently offering still blocks.
func TestAssertVariantDeletableBlocksOnTheVendingMachine(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertVariantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t,
			variantUsageReply{module: "sales", used: false},
			variantUsageReply{module: "purchase", used: false},
			variantUsageReply{module: "vendingmachine", used: true}),
		variantParams(), vErrs)

	require.NoError(t, err)
	require.Equal(t, 1, vErrs.Count())
	assert.Contains(t, (*vErrs)[0].Message, "vendingmachine")
}

// A module that cannot answer blocks the delete: an unanswered question is not a "no".
func TestAssertVariantDeletableRefusesWhenAModuleCannotAnswer(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertVariantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t,
			variantUsageReply{module: "sales", failure: "database is down"},
			variantUsageReply{module: "purchase", used: false}),
		variantParams(), vErrs)

	require.NoError(t, err)
	assert.Equal(t, 1, vErrs.Count())
	assert.Equal(t, usagecheck.ErrorKeyResourceInUse, (*vErrs)[0].Key)
}

// Nothing to identify means nothing to ask about.
func TestAssertVariantDeletableSkipsWithoutAnId(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertVariantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t, variantUsageReply{module: "sales", used: true}),
		dmodel.DynamicFields{}, vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}
