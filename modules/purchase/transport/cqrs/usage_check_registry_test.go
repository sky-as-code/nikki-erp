package cqrs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	invconstants "github.com/sky-as-code/nikki-erp/modules/inventory/constants"
)

// Inventory asks every dependant about each resource it guards; a pair with no checker is answered
// with an error and refuses the delete. So every resource Inventory dispatches must resolve here.
func TestUsageCheckRegistryAnswersEveryInventoryResource(t *testing.T) {
	registry, err := newUsageCheckRegistry()
	require.NoError(t, err)

	for _, resource := range []string{
		usagecheck.ResourceProductVariant,
		usagecheck.ResourceProductTemplate,
		usagecheck.ResourceStockQuant,
	} {
		_, found := registry.Checker(invconstants.InventoryModuleName, resource)
		assert.Truef(t, found, "no checker for %s.%s", invconstants.InventoryModuleName, resource)
	}
}

// No Purchase table names a stock quant, so the checker answers "unused" without searching anything.
func TestStockQuantCheckerAnswersUnused(t *testing.T) {
	registry, err := newUsageCheckRegistry()
	require.NoError(t, err)

	checker, found := registry.Checker(invconstants.InventoryModuleName, usagecheck.ResourceStockQuant)
	require.True(t, found)

	used, usedBy, err := checker.IsUsed(nil, usagecheck.ResourceRef{
		ResourceName: usagecheck.ResourceStockQuant,
		Identifier:   usagecheck.NewIdentifier("01QUANT00000000000000001", "01ORG00000000000000000001"),
	})
	require.NoError(t, err)
	assert.False(t, used)
	assert.Empty(t, usedBy)
}
