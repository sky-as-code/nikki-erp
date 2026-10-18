package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/accounting/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/accounting/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The prefix is the same string the IAM resource codes and schema names share; drifting from it
// leaves the endpoints reachable at a path no client calls.
func TestModuleRoutePrefixIsStable(t *testing.T) {
	assert.Equal(t, "/v1/accounting", modconstants.AccountingRouteV1)
}

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	for schemaName, literal := range map[string]string{
		models.TaxJurisdictionSchemaName:          v1.TaxJurisdictionEngineName,
		models.TaxGroupSchemaName:                 v1.TaxGroupEngineName,
		models.TaxRoundingPolicySchemaName:        v1.TaxRoundingPolicyEngineName,
		models.TaxProductClassificationSchemaName: v1.TaxProductClassificationEngineName,
		models.TaxSchemaName:                      v1.TaxEngineName,
		models.TaxDefinitionVersionSchemaName:     v1.TaxDefinitionVersionEngineName,
		models.TaxRateVersionSchemaName:           v1.TaxRateVersionEngineName,
		models.TaxComponentSchemaName:             v1.TaxComponentEngineName,
		models.TaxMappingSchemaName:               v1.TaxMappingEngineName,
		models.TaxMappingLineSchemaName:           v1.TaxMappingLineEngineName,
		models.TaxRuleSchemaName:                  v1.TaxRuleEngineName,
		models.TaxRuleConditionSchemaName:         v1.TaxRuleConditionEngineName,
		models.TaxRuleResultSchemaName:            v1.TaxRuleResultEngineName,
	} {
		assert.Equal(t, composable.EngineDependencyName(schemaName), literal)
	}
	assert.Len(t, dynamicengines.EngineSchemaNames(), 13)
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// The four calculation actions are POST collection actions of accounting_tax, registered ahead of
// the ":id" routes that would otherwise swallow their literal segment.
func TestTaxCalculationActionsAreCollectionPosts(t *testing.T) {
	noop := func(*echo.Context, map[string]any) error { return nil }
	restEngine := composable.NewRestEngine(models.TaxSchemaName, &v1.TaxRest{}).AddCrudRoutes()
	for _, path := range TaxCalculationActionPaths {
		restEngine.AddRoute(taxAction(path, noop))
	}
	echoApp := echo.New()
	require.NoError(t, restEngine.RegisterRoutes(echoApp.Group(modconstants.AccountingRouteV1), passThrough))

	position := map[string]int{}
	for i, route := range echoApp.Router().Routes() {
		position[route.Method+" "+route.Path] = i
	}
	base := modconstants.AccountingRouteV1 + "/" + models.TaxSchemaName
	for _, path := range []string{"calculate", "simulate", "reverse_full", "reverse_partial"} {
		index, ok := position["POST "+base+"/"+path]
		require.True(t, ok, path)
		assert.Less(t, index, position["POST "+base+"/:id/archived"], path)
	}
}
