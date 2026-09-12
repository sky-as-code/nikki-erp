package usagecheck

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func salesCommand(resources ...ResourceRef) *CheckResourceUsageCommand {
	if len(resources) == 0 {
		resources = []ResourceRef{productRef()}
	}
	return &CheckResourceUsageCommand{
		RequestId:    "REQ-1",
		SourceModule: "inventory",
		TargetModule: "sales",
		Resources:    resources,
	}
}

// handle runs what the bus would run, minus the packet wrapper: handle() unwraps the packet and
// hands the command to answer(), which is everything under test here.
func handle(
	t *testing.T, registry *CheckerRegistry, cmd *CheckResourceUsageCommand,
) (*cqrs.Reply[CheckResourceUsageResult], error) {
	t.Helper()
	handler := &checkUsageHandler{moduleName: "sales", registry: registry}
	return handler.answer(corectx.NewRequestContext(context.Background()), cmd)
}

func registryWithChecker(t *testing.T, resource string, checker ResourceUsageChecker) *CheckerRegistry {
	t.Helper()
	registry := NewCheckerRegistry()
	require.NoError(t, registry.Register("inventory", resource, checker))
	return registry
}

func TestHandler_AnswersPerResource(t *testing.T) {
	registry := registryWithChecker(t, "product_variant",
		CheckerFunc(func(_ corectx.Context, ref ResourceRef) (bool, string, error) {
			return ref.Identifier["id"] == "VAR-1", "sales_order_line", nil
		}))

	otherRef := ResourceRef{
		ResourceName: "product_variant",
		Identifier:   map[string]string{"id": "VAR-2", "org_id": "ORG-1"},
	}

	reply, err := handle(t, registry, salesCommand(productRef(), otherRef))

	require.NoError(t, err)
	require.Len(t, reply.Result.Results, 2)
	assert.Equal(t, "sales", reply.Result.Module)
	assert.Equal(t, "REQ-1", reply.Result.RequestId)
	assert.True(t, reply.Result.Results[0].IsUsed)
	assert.Equal(t, "sales_order_line", reply.Result.Results[0].UsedBy)
	assert.False(t, reply.Result.Results[1].IsUsed)
	assert.Equal(t, map[string]string{"id": "VAR-2", "org_id": "ORG-1"},
		reply.Result.Results[1].Identifier, "the identifier is echoed so answers can be matched")
}

// A checker that fails must surface as an error, never as "not used": the caller blocks on an
// error and would proceed on a false.
func TestHandler_CheckerErrorBecomesAnError(t *testing.T) {
	registry := registryWithChecker(t, "product_variant",
		CheckerFunc(func(corectx.Context, ResourceRef) (bool, string, error) {
			return false, "", errors.New("query failed")
		}))

	_, err := handle(t, registry, salesCommand())

	require.Error(t, err)
	assert.ErrorContains(t, err, "query failed")
}

// Silence would read as "not used" and permit the delete. The owning module believes this module
// answers about the resource, so having no checker is a wiring mistake worth reporting.
func TestHandler_UnknownResourceIsAnError(t *testing.T) {
	registry := registryWithChecker(t, "product_variant", CheckerFunc(
		func(corectx.Context, ResourceRef) (bool, string, error) { return false, "", nil }))

	cmd := salesCommand(ResourceRef{
		ResourceName: "uom",
		Identifier:   map[string]string{"id": "UOM-1"},
	})

	_, err := handle(t, registry, cmd)

	require.Error(t, err)
	assert.ErrorContains(t, err, "no checker")
}

func TestHandler_RejectsCommandAddressedElsewhere(t *testing.T) {
	registry := registryWithChecker(t, "product_variant", CheckerFunc(
		func(corectx.Context, ResourceRef) (bool, string, error) { return false, "", nil }))

	cmd := salesCommand()
	cmd.TargetModule = "purchase"

	_, err := handle(t, registry, cmd)

	require.Error(t, err)
	assert.ErrorContains(t, err, "addressed to purchase")
}

func TestHandler_RejectsMalformedCommand(t *testing.T) {
	registry := NewCheckerRegistry()

	cases := map[string]func(*CheckResourceUsageCommand){
		"no source module": func(cmd *CheckResourceUsageCommand) { cmd.SourceModule = "" },
		"no resources":     func(cmd *CheckResourceUsageCommand) { cmd.Resources = nil },
		"no resource name": func(cmd *CheckResourceUsageCommand) {
			cmd.Resources = []ResourceRef{{Identifier: map[string]string{"id": "X"}}}
		},
		"no identifier": func(cmd *CheckResourceUsageCommand) {
			cmd.Resources = []ResourceRef{{ResourceName: "product_variant"}}
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := salesCommand()
			mutate(cmd)

			_, err := handle(t, registry, cmd)

			assert.Error(t, err)
		})
	}
}

func TestCheckerRegistry_RejectsDuplicateRegistration(t *testing.T) {
	checker := CheckerFunc(func(corectx.Context, ResourceRef) (bool, string, error) {
		return false, "", nil
	})
	registry := NewCheckerRegistry()
	require.NoError(t, registry.Register("inventory", "product_variant", checker))

	err := registry.Register("inventory", "product_variant", checker)

	require.Error(t, err, "two checkers for one resource means one of them silently does not run")
	assert.ErrorContains(t, err, "already registered")
}

func TestCheckerRegistry_SeparatesBySourceModuleAndResource(t *testing.T) {
	registry := NewCheckerRegistry()
	uom := CheckerFunc(func(corectx.Context, ResourceRef) (bool, string, error) {
		return true, "uom", nil
	})
	variant := CheckerFunc(func(corectx.Context, ResourceRef) (bool, string, error) {
		return false, "", nil
	})
	require.NoError(t, registry.Register("essential", "uom", uom))
	require.NoError(t, registry.Register("inventory", "product_variant", variant))

	_, foundUom := registry.Checker("essential", "uom")
	_, foundVariant := registry.Checker("inventory", "product_variant")
	_, foundCrossed := registry.Checker("essential", "product_variant")

	assert.True(t, foundUom)
	assert.True(t, foundVariant)
	assert.False(t, foundCrossed, "the pair is the key, not either half")
}

func TestCheckerRegistry_RejectsIncompleteRegistration(t *testing.T) {
	registry := NewCheckerRegistry()
	checker := CheckerFunc(func(corectx.Context, ResourceRef) (bool, string, error) {
		return false, "", nil
	})

	assert.Error(t, registry.Register("", "product_variant", checker))
	assert.Error(t, registry.Register("inventory", "", checker))
	assert.Error(t, registry.Register("inventory", "product_variant", nil))
}

// The bus derives a handler's topic from the request its NewRequest returns. cqrs.NewHandler
// returns a ZERO value, whose TargetModule is empty — so every module would subscribe to the
// same "_usage.checkResourceUsage" topic, the second to start would collide with the first, and
// no lookup for a module-specific type would ever find a handler.
//
// This cost a failed boot to discover, because a fake bus does not exercise NewRequest.
func TestSubscribedTypeIsModuleSpecific(t *testing.T) {
	handler := newModuleHandler("sales", nil)

	request, ok := handler.NewRequest().(*CheckResourceUsageCommand)
	require.True(t, ok)

	assert.Equal(t, "sales", request.TargetModule)
	assert.Equal(t, RequestTypeFor("sales").String(), request.CqrsRequestType().String(),
		"the topic subscribed to must be the one the owning module addresses")
	assert.NotEqual(t, CheckResourceUsageCommand{}.CqrsRequestType().String(),
		request.CqrsRequestType().String(), "a zero value would collide across modules")
}

// Two modules must land on two different topics.
func TestSubscribedTypesDifferBetweenModules(t *testing.T) {
	sales := newModuleHandler("sales", nil).NewRequest().(*CheckResourceUsageCommand)
	purchase := newModuleHandler("purchase", nil).NewRequest().(*CheckResourceUsageCommand)

	assert.NotEqual(t, sales.CqrsRequestType().String(), purchase.CqrsRequestType().String())
}
