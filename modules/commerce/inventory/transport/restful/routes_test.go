package restful

import (
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The container names in the REST structs' tags are literals; they must be what the onions are
// registered under, or the handlers resolve nothing at boot.
func TestEngineNamesMatchTheRegisteredOnions(t *testing.T) {
	for schemaName, engineName := range map[string]string{
		models.ProductCategorySchemaName:                     v1.ProductCategoryEngineName,
		models.BrandSchemaName:                               v1.BrandEngineName,
		models.ProductAttributeSchemaName:                    v1.ProductAttributeEngineName,
		models.ProductAttributeValueSchemaName:               v1.ProductAttributeValueEngineName,
		models.ProductTemplateSchemaName:                     v1.ProductTemplateEngineName,
		models.ProductTemplateAttributeSchemaName:            v1.ProductTemplateAttributeEngineName,
		models.ProductTemplateAttributeValueSchemaName:       v1.ProductTemplateAttributeValueEngineName,
		models.ProductVariantSchemaName:                      v1.ProductVariantEngineName,
		models.ProductVariantAttributeValueSchemaName:        v1.ProductVariantAttributeValueEngineName,
		models.ProductTemplateCombinationExclusionSchemaName: v1.ProductTemplateCombinationExclusionEngineName,
		models.WarehouseSchemaName:                           v1.WarehouseEngineName,
		models.StorageCategorySchemaName:                     v1.StorageCategoryEngineName,
		models.InventoryLocationSchemaName:                   v1.InventoryLocationEngineName,
		models.WarehouseSupplyRelationSchemaName:             v1.WarehouseSupplyRelationEngineName,
		models.PutawayRuleSchemaName:                         v1.PutawayRuleEngineName,
		models.StockOperationTypeSchemaName:                  v1.StockOperationTypeEngineName,
		models.StockQuantSchemaName:                          v1.StockQuantEngineName,
		models.StockTransferSchemaName:                       v1.StockTransferEngineName,
		models.StockMoveSchemaName:                           v1.StockMoveEngineName,
		models.StockMoveLineSchemaName:                       v1.StockMoveLineEngineName,
		models.StockMoveDependencySchemaName:                 v1.StockMoveDependencyEngineName,
		models.StockScrapSchemaName:                          v1.StockScrapEngineName,
		models.StockProductConfigSchemaName:                  v1.StockProductConfigEngineName,
		models.StockReservationSchemaName:                    v1.StockReservationEngineName,
	} {
		assert.Equal(t, composable.EngineDependencyName(schemaName), engineName, schemaName)
	}
}

// registeredRoutes runs one resource's route registration on a throwaway echo instance and
// returns "METHOD path" in registration order.
func registeredRoutes(t *testing.T, register func(route *echo.Group) error) []string {
	t.Helper()

	echoApp := echo.New()
	require.NoError(t, register(echoApp.Group("/v1/inventory")))

	routes := echoApp.Router().Routes()
	result := make([]string, 0, len(routes))
	for _, route := range routes {
		result = append(result, route.Method+" "+route.Path)
	}
	return result
}

func passThrough(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

// Every custom route the module declares, with the collection-level ones registered ahead of the
// ":id" patterns that would otherwise swallow them, and get_effective staying a GET.
func TestCustomRoutesRegisterAheadOfIdPatterns(t *testing.T) {
	routes := registeredRoutes(t, func(route *echo.Group) error {
		quant := &v1.StockQuantRest{}
		transfer := &v1.StockTransferRest{}
		template := &v1.ProductTemplateRest{}
		variant := &v1.ProductVariantRest{}
		return errorsJoin(
			restEngineOf(models.StockQuantSchemaName, quant, quantRoutes(quant)).RegisterRoutes(route, passThrough),
			restEngineOf(models.StockTransferSchemaName, transfer, transferRoutes(transfer)).RegisterRoutes(route, passThrough),
			restEngineOf(models.ProductTemplateSchemaName, template, templateRoutes(template)).RegisterRoutes(route, passThrough),
			restEngineOf(models.ProductVariantSchemaName, variant, variantRoutes(variant)).RegisterRoutes(route, passThrough),
		)
	})

	quantBase := "/v1/inventory/" + models.StockQuantSchemaName
	assert.Less(t, indexOf(routes, "POST "+quantBase+"/variant_stock_summary"), indexOf(routes, "GET "+quantBase+"/:id"))
	assert.Less(t, indexOf(routes, "POST "+quantBase+"/:id/enter_count"), indexOf(routes, "PATCH "+quantBase+"/:id"))

	transferBase := "/v1/inventory/" + models.StockTransferSchemaName
	assert.Less(t, indexOf(routes, "POST "+transferBase+"/reserve_for_source"), indexOf(routes, "GET "+transferBase+"/:id"))
	assert.Contains(t, routes, "POST "+transferBase+"/:id/validate")

	templateBase := "/v1/inventory/" + models.ProductTemplateSchemaName
	assert.Less(t, indexOf(routes, "POST "+templateBase+"/resolve_selection"), indexOf(routes, "GET "+templateBase+"/:id"))
	assert.Contains(t, routes, "POST "+templateBase+"/:id/generate_variants")
	assert.Contains(t, routes, "POST "+templateBase+"/:id/duplicate")

	assert.Contains(t, routes, "GET /v1/inventory/"+models.ProductVariantSchemaName+"/:id/effective")
}

func restEngineOf(schemaName string, handlers composable.CrudRestHandlers, defs []composable.RouteDefinition) composable.RestEngine {
	restEngine := composable.NewRestEngine(schemaName, handlers).AddCrudRoutes()
	for _, def := range defs {
		restEngine.AddRoute(def)
	}
	return restEngine
}

func quantRoutes(rest *v1.StockQuantRest) []composable.RouteDefinition {
	return []composable.RouteDefinition{
		{Path: ":id/enter_count", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.EnterCount},
		{Path: "variant_stock_summary", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.VariantStockSummary},
	}
}

func transferRoutes(rest *v1.StockTransferRest) []composable.RouteDefinition {
	return []composable.RouteDefinition{
		{Path: ":id/validate", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Validate},
		{Path: "reserve_for_source", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReserveForSource},
	}
}

func templateRoutes(rest *v1.ProductTemplateRest) []composable.RouteDefinition {
	return []composable.RouteDefinition{
		{Path: ":id/generate_variants", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.GenerateVariants},
		{Path: "resolve_selection", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ResolveSelection},
		{Path: ":id/duplicate", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Duplicate},
	}
}

func variantRoutes(rest *v1.ProductVariantRest) []composable.RouteDefinition {
	return []composable.RouteDefinition{
		{Path: ":id/effective", ActionType: composable.ActionTypeRead, HandlerFn: rest.GetEffective},
	}
}

func errorsJoin(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func indexOf(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}

// The configuration rows are written only through the template's configuration payload, so their
// route tables carry no write method.
func TestConfigurationRowsServeReadsOnly(t *testing.T) {
	routes := registeredRoutes(t, func(route *echo.Group) error {
		rest := &v1.ProductTemplateCombinationExclusionRest{}
		return composable.NewRestEngine(models.ProductTemplateCombinationExclusionSchemaName, rest).
			AddCrudRoutes(configurationReadRoutes()...).
			RegisterRoutes(route, passThrough)
	})

	base := "/v1/inventory/" + models.ProductTemplateCombinationExclusionSchemaName
	require.Contains(t, routes, "GET "+base)
	for _, route := range routes {
		assert.NotRegexp(t, `^(PUT|PATCH|DELETE) `, route)
		assert.NotRegexp(t, `/(archived|import|bulk)$`, route)
		assert.NotEqual(t, "POST "+base, route)
	}
}

// A template value row serves PATCH :id for sales_price_extra and nothing else that writes.
func TestTemplateValueRowsServeReadsAndPatchOnly(t *testing.T) {
	routes := registeredRoutes(t, func(route *echo.Group) error {
		rest := &v1.ProductTemplateAttributeValueRest{}
		return composable.NewRestEngine(models.ProductTemplateAttributeValueSchemaName, rest).
			AddCrudRoutes(templateValueRoutes()...).
			RegisterRoutes(route, passThrough)
	})

	base := "/v1/inventory/" + models.ProductTemplateAttributeValueSchemaName
	require.Contains(t, routes, "GET "+base)
	require.Contains(t, routes, "PATCH "+base+"/:id")
	for _, route := range routes {
		assert.NotRegexp(t, `^(PUT|DELETE) `, route)
		assert.NotRegexp(t, `/(archived|import|bulk)$`, route)
		assert.NotEqual(t, "POST "+base, route)
	}
}
