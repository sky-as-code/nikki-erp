package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewProductCategoryRest,
		v1.NewBrandRest,
		v1.NewProductAttributeRest,
		v1.NewProductAttributeValueRest,
		v1.NewProductTemplateRest,
		v1.NewProductTemplateAttributeRest,
		v1.NewProductTemplateAttributeValueRest,
		v1.NewProductVariantRest,
		v1.NewProductVariantAttributeValueRest,
		v1.NewProductTemplateCombinationExclusionRest,
		v1.NewWarehouseRest,
		v1.NewStorageCategoryRest,
		v1.NewInventoryLocationRest,
		v1.NewWarehouseSupplyRelationRest,
		v1.NewPutawayRuleRest,
		v1.NewStockOperationTypeRest,
		v1.NewStockQuantRest,
		v1.NewStockTransferRest,
		v1.NewStockMoveRest,
		v1.NewStockMoveLineRest,
		v1.NewStockMoveDependencyRest,
		v1.NewStockScrapRest,
		v1.NewStockProductConfigRest,
		v1.NewStockReservationRest,
	)
	return stdErr.Join(err, initInventoryV1())
}

// initInventoryV1 registers every resource at /v1/inventory/{schema_name}: the built-in route
// table of the composable engine, plus the custom routes a resource declares.
func initInventoryV1() error {
	return deps.Invoke(func(route *echo.Group) error {
		routeV1 := route.Group("/v1/inventory")
		return stdErr.Join(
			initProductCategoryV1(routeV1),
			initBrandV1(routeV1),
			initProductAttributeV1(routeV1),
			initProductAttributeValueV1(routeV1),
			initProductTemplateV1(routeV1),
			initProductTemplateAttributeV1(routeV1),
			initProductTemplateAttributeValueV1(routeV1),
			initProductVariantV1(routeV1),
			initProductVariantAttributeValueV1(routeV1),
			initProductTemplateCombinationExclusionV1(routeV1),
			initWarehouseV1(routeV1),
			initStorageCategoryV1(routeV1),
			initInventoryLocationV1(routeV1),
			initWarehouseSupplyRelationV1(routeV1),
			initPutawayRuleV1(routeV1),
			initStockOperationTypeV1(routeV1),
			initStockQuantV1(routeV1),
			initStockTransferV1(routeV1),
			initStockMoveV1(routeV1),
			initStockMoveLineV1(routeV1),
			initStockMoveDependencyV1(routeV1),
			initStockScrapV1(routeV1),
			initStockProductConfigV1(routeV1),
			initStockReservationV1(routeV1),
		)
	})
}

func initProductCategoryV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductCategoryRest) error {
		return dynamicresource.NewRestEngine(models.ProductCategorySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initBrandV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.BrandRest) error {
		return dynamicresource.NewRestEngine(models.BrandSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductAttributeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductAttributeRest) error {
		return dynamicresource.NewRestEngine(models.ProductAttributeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductAttributeValueRest) error {
		return dynamicresource.NewRestEngine(models.ProductAttributeValueSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductTemplateV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateRest) error {
		return dynamicresource.NewRestEngine(models.ProductTemplateSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/generate_variants", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.GenerateVariants}).
			AddRoute(dynamicresource.RouteDefinition{Path: "resolve_selection", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ResolveSelection}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/duplicate", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Duplicate}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/preview_attribute_configuration", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.PreviewAttributeConfiguration}).
			RegisterRoutes(route)
	})
}

func initProductTemplateAttributeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateAttributeRest) error {
		return dynamicresource.NewRestEngine(models.ProductTemplateAttributeSchemaName, rest).
			AddCrudRoutes(configurationReadRoutes()...).
			RegisterRoutes(route)
	})
}

func initProductTemplateAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateAttributeValueRest) error {
		return dynamicresource.NewRestEngine(models.ProductTemplateAttributeValueSchemaName, rest).
			AddCrudRoutes(templateValueRoutes()...).
			RegisterRoutes(route)
	})
}

// initProductTemplateCombinationExclusionV1 serves the reads only, as the junctions above do.
func initProductTemplateCombinationExclusionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateCombinationExclusionRest) error {
		return dynamicresource.NewRestEngine(models.ProductTemplateCombinationExclusionSchemaName, rest).
			AddCrudRoutes(configurationReadRoutes()...).
			RegisterRoutes(route)
	})
}

// configurationReadRoutes is the route table of the template configuration rows: template lines,
// their values, variant values and exclusions. They are written only through the template's own
// configuration payload, so the UI, the API and an import share one path that reconciles the
// variants; a direct write here would change the configuration without the reconcile. Import is
// withheld for the same reason.
func configurationReadRoutes() []dynamicresource.CrudAction {
	return []dynamicresource.CrudAction{
		dynamicresource.CrudActionGetById,
		dynamicresource.CrudActionGetByUnique,
		dynamicresource.CrudActionSearch,
		dynamicresource.CrudActionExists,
		dynamicresource.CrudActionGetSchema,
		dynamicresource.CrudActionComputeField,
	}
}

// templateValueRoutes adds PATCH :id to the reads: the domain service accepts only
// sales_price_extra there, which changes no variant identity and so needs no reconcile.
func templateValueRoutes() []dynamicresource.CrudAction {
	return append(configurationReadRoutes(), dynamicresource.CrudActionUpdate)
}

func initProductVariantV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductVariantRest) error {
		return dynamicresource.NewRestEngine(models.ProductVariantSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/effective", ActionType: dynamicresource.ActionTypeRead, HandlerFn: rest.GetEffective}).
			RegisterRoutes(route)
	})
}

func initProductVariantAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductVariantAttributeValueRest) error {
		return dynamicresource.NewRestEngine(models.ProductVariantAttributeValueSchemaName, rest).
			AddCrudRoutes(configurationReadRoutes()...).
			RegisterRoutes(route)
	})
}

func initWarehouseV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.WarehouseRest) error {
		return dynamicresource.NewRestEngine(models.WarehouseSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/suspend", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Suspend}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/resume", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Resume}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/configure_incoming_flow", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ConfigureIncomingFlow}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/configure_outgoing_flow", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ConfigureOutgoingFlow}).
			RegisterRoutes(route)
	})
}

func initStorageCategoryV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StorageCategoryRest) error {
		return dynamicresource.NewRestEngine(models.StorageCategorySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initInventoryLocationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.InventoryLocationRest) error {
		return dynamicresource.NewRestEngine(models.InventoryLocationSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/suspend", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Suspend}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/resume", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Resume}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/move", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Move}).
			RegisterRoutes(route)
	})
}

func initWarehouseSupplyRelationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.WarehouseSupplyRelationRest) error {
		return dynamicresource.NewRestEngine(models.WarehouseSupplyRelationSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initPutawayRuleV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.PutawayRuleRest) error {
		return dynamicresource.NewRestEngine(models.PutawayRuleSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: "suggest_location", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.SuggestLocation}).
			RegisterRoutes(route)
	})
}

func initStockOperationTypeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockOperationTypeRest) error {
		return dynamicresource.NewRestEngine(models.StockOperationTypeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockQuantV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockQuantRest) error {
		return dynamicresource.NewRestEngine(models.StockQuantSchemaName, rest).
			AddCrudRoutes(dynamicresource.ReadCrudActions()...).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/enter_count", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.EnterCount}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/reset_count", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ResetCount}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/apply_adjustment", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ApplyAdjustment}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/schedule_count", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ScheduleCount}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/assign_counter", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.AssignCounter}).
			AddRoute(dynamicresource.RouteDefinition{Path: "variant_stock_summary", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.VariantStockSummary}).
			AddRoute(dynamicresource.RouteDefinition{Path: "variant_stock_summaries", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.VariantStockSummaries}).
			AddRoute(dynamicresource.RouteDefinition{Path: "template_stock_summary", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.TemplateStockSummary}).
			AddRoute(dynamicresource.RouteDefinition{Path: "stock_by_warehouse", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.StockByWarehouse}).
			AddRoute(dynamicresource.RouteDefinition{Path: "stock_by_location", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.StockByLocation}).
			AddRoute(dynamicresource.RouteDefinition{Path: "product_usage", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ProductUsage}).
			RegisterRoutes(route)
	})
}

func initStockTransferV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockTransferRest) error {
		return dynamicresource.NewRestEngine(models.StockTransferSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/confirm", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Confirm}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/check_availability", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CheckAvailability}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/reserve", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Reserve}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/unreserve", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Unreserve}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/validate", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Validate}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/cancel", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.Cancel}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/create_return", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CreateReturn}).
			AddRoute(dynamicresource.RouteDefinition{Path: "reserve_for_source", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReserveForSource}).
			AddRoute(dynamicresource.RouteDefinition{Path: "release_for_source", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReleaseForSource}).
			AddRoute(dynamicresource.RouteDefinition{Path: "reallocate_reservation", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReallocateReservation}).
			AddRoute(dynamicresource.RouteDefinition{Path: "apply_fulfillment_result", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ApplyFulfillmentResult}).
			RegisterRoutes(route)
	})
}

func initStockMoveV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveRest) error {
		return dynamicresource.NewRestEngine(models.StockMoveSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockMoveLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveLineRest) error {
		return dynamicresource.NewRestEngine(models.StockMoveLineSchemaName, rest).
			AddCrudRoutes(dynamicresource.ReadCrudActions()...).
			RegisterRoutes(route)
	})
}

func initStockMoveDependencyV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveDependencyRest) error {
		return dynamicresource.NewRestEngine(models.StockMoveDependencySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockScrapV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockScrapRest) error {
		return dynamicresource.NewRestEngine(models.StockScrapSchemaName, rest).
			AddCrudRoutes().
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/do_scrap", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.DoScrap}).
			RegisterRoutes(route)
	})
}

func initStockProductConfigV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockProductConfigRest) error {
		return dynamicresource.NewRestEngine(models.StockProductConfigSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// initStockReservationV1 serves the reads only: a reservation is written by its operations, which
// register as custom routes as they land, never by the built-in create, update or delete.
func initStockReservationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockReservationRest) error {
		return dynamicresource.NewRestEngine(models.StockReservationSchemaName, rest).
			AddCrudRoutes(
				dynamicresource.CrudActionGetById,
				dynamicresource.CrudActionGetByUnique,
				dynamicresource.CrudActionSearch,
				dynamicresource.CrudActionExists,
				dynamicresource.CrudActionGetSchema,
				dynamicresource.CrudActionComputeField,
			).
			AddRoute(dynamicresource.RouteDefinition{Path: "reserve_warehouse_stock", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReserveWarehouseStock}).
			AddRoute(dynamicresource.RouteDefinition{Path: "check_warehouse_availability", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.CheckWarehouseAvailability}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/consume", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ConsumeReservation}).
			AddRoute(dynamicresource.RouteDefinition{Path: ":id/release", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ReleaseReservation}).
			RegisterRoutes(route)
	})
}
