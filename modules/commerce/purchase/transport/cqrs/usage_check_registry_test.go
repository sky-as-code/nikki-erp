package cqrs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commconstants "github.com/sky-as-code/nikki-erp/modules/commerce/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// Inventory asks Purchase in process about each resource it guards, so every one must resolve here.
func TestUsageCheckRegistryAnswersEveryInventoryResource(t *testing.T) {
	registry, err := InventoryUsageRegistry()
	require.NoError(t, err)

	for _, resource := range []string{
		usagecheck.ResourceProductVariant,
		usagecheck.ResourceProductTemplate,
		usagecheck.ResourceStockQuant,
	} {
		_, found := registry.Checker(commconstants.CommerceModuleName, resource)
		assert.Truef(t, found, "no checker for %s.%s", commconstants.CommerceModuleName, resource)
	}
}

// No Purchase table names a stock quant, so the checker answers "unused" without searching anything.
func TestStockQuantCheckerAnswersUnused(t *testing.T) {
	registry, err := InventoryUsageRegistry()
	require.NoError(t, err)

	checker, found := registry.Checker(commconstants.CommerceModuleName, usagecheck.ResourceStockQuant)
	require.True(t, found)

	used, usedBy, err := checker.IsUsed(nil, usagecheck.ResourceRef{
		ResourceName: usagecheck.ResourceStockQuant,
		Identifier:   usagecheck.NewIdentifier("01QUANT00000000000000001", "01ORG00000000000000000001"),
	})
	require.NoError(t, err)
	assert.False(t, used)
	assert.Empty(t, usedBy)
}
