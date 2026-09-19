package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/inventory/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewProductTypeRest,
		v1.NewProductCategoryRest,
		v1.NewBrandRest,
		v1.NewProductAttributeRest,
		v1.NewProductAttributeValueRest,
		v1.NewProductTemplateRest,
		v1.NewProductTemplateAttributeRest,
		v1.NewProductTemplateAttributeValueRest,
		v1.NewProductVariantRest,
		v1.NewProductVariantAttributeValueRest,
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
			initProductTypeV1(routeV1),
			initProductCategoryV1(routeV1),
			initBrandV1(routeV1),
			initProductAttributeV1(routeV1),
			initProductAttributeValueV1(routeV1),
			initProductTemplateV1(routeV1),
			initProductTemplateAttributeV1(routeV1),
			initProductTemplateAttributeValueV1(routeV1),
			initProductVariantV1(routeV1),
			initProductVariantAttributeValueV1(routeV1),
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

func initProductTypeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTypeRest) error {
		return composable.NewRestEngine(models.ProductTypeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductCategoryV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductCategoryRest) error {
		return composable.NewRestEngine(models.ProductCategorySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initBrandV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.BrandRest) error {
		return composable.NewRestEngine(models.BrandSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductAttributeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductAttributeRest) error {
		return composable.NewRestEngine(models.ProductAttributeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductAttributeValueRest) error {
		return composable.NewRestEngine(models.ProductAttributeValueSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductTemplateV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateRest) error {
		return composable.NewRestEngine(models.ProductTemplateSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/generate_variants", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.GenerateVariants}).
			AddRoute(composable.RouteDefinition{Path: "resolve_selection", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ResolveSelection}).
			RegisterRoutes(route)
	})
}

func initProductTemplateAttributeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateAttributeRest) error {
		return composable.NewRestEngine(models.ProductTemplateAttributeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductTemplateAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductTemplateAttributeValueRest) error {
		return composable.NewRestEngine(models.ProductTemplateAttributeValueSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initProductVariantV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductVariantRest) error {
		return composable.NewRestEngine(models.ProductVariantSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/effective", ActionType: composable.ActionTypeRead, HandlerFn: rest.GetEffective}).
			RegisterRoutes(route)
	})
}

func initProductVariantAttributeValueV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.ProductVariantAttributeValueRest) error {
		return composable.NewRestEngine(models.ProductVariantAttributeValueSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initWarehouseV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.WarehouseRest) error {
		return composable.NewRestEngine(models.WarehouseSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/suspend", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Suspend}).
			AddRoute(composable.RouteDefinition{Path: ":id/resume", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Resume}).
			AddRoute(composable.RouteDefinition{Path: ":id/configure_incoming_flow", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ConfigureIncomingFlow}).
			AddRoute(composable.RouteDefinition{Path: ":id/configure_outgoing_flow", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ConfigureOutgoingFlow}).
			RegisterRoutes(route)
	})
}

func initStorageCategoryV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StorageCategoryRest) error {
		return composable.NewRestEngine(models.StorageCategorySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initInventoryLocationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.InventoryLocationRest) error {
		return composable.NewRestEngine(models.InventoryLocationSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/suspend", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Suspend}).
			AddRoute(composable.RouteDefinition{Path: ":id/resume", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Resume}).
			AddRoute(composable.RouteDefinition{Path: ":id/move", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Move}).
			RegisterRoutes(route)
	})
}

func initWarehouseSupplyRelationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.WarehouseSupplyRelationRest) error {
		return composable.NewRestEngine(models.WarehouseSupplyRelationSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initPutawayRuleV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.PutawayRuleRest) error {
		return composable.NewRestEngine(models.PutawayRuleSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: "suggest_location", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.SuggestLocation}).
			RegisterRoutes(route)
	})
}

func initStockOperationTypeV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockOperationTypeRest) error {
		return composable.NewRestEngine(models.StockOperationTypeSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockQuantV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockQuantRest) error {
		return composable.NewRestEngine(models.StockQuantSchemaName, rest).
			AddCrudRoutes(composable.ReadCrudActions()...).
			AddRoute(composable.RouteDefinition{Path: ":id/enter_count", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.EnterCount}).
			AddRoute(composable.RouteDefinition{Path: ":id/reset_count", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ResetCount}).
			AddRoute(composable.RouteDefinition{Path: ":id/apply_adjustment", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ApplyAdjustment}).
			AddRoute(composable.RouteDefinition{Path: ":id/schedule_count", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ScheduleCount}).
			AddRoute(composable.RouteDefinition{Path: ":id/assign_counter", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.AssignCounter}).
			AddRoute(composable.RouteDefinition{Path: "variant_stock_summary", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.VariantStockSummary}).
			AddRoute(composable.RouteDefinition{Path: "variant_stock_summaries", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.VariantStockSummaries}).
			AddRoute(composable.RouteDefinition{Path: "template_stock_summary", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.TemplateStockSummary}).
			AddRoute(composable.RouteDefinition{Path: "stock_by_warehouse", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.StockByWarehouse}).
			AddRoute(composable.RouteDefinition{Path: "stock_by_location", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.StockByLocation}).
			AddRoute(composable.RouteDefinition{Path: "product_usage", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ProductUsage}).
			RegisterRoutes(route)
	})
}

func initStockTransferV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockTransferRest) error {
		return composable.NewRestEngine(models.StockTransferSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/confirm", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Confirm}).
			AddRoute(composable.RouteDefinition{Path: ":id/check_availability", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CheckAvailability}).
			AddRoute(composable.RouteDefinition{Path: ":id/reserve", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Reserve}).
			AddRoute(composable.RouteDefinition{Path: ":id/unreserve", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Unreserve}).
			AddRoute(composable.RouteDefinition{Path: ":id/validate", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Validate}).
			AddRoute(composable.RouteDefinition{Path: ":id/cancel", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.Cancel}).
			AddRoute(composable.RouteDefinition{Path: ":id/create_return", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CreateReturn}).
			AddRoute(composable.RouteDefinition{Path: "reserve_for_source", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReserveForSource}).
			AddRoute(composable.RouteDefinition{Path: "release_for_source", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReleaseForSource}).
			AddRoute(composable.RouteDefinition{Path: "reallocate_reservation", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReallocateReservation}).
			AddRoute(composable.RouteDefinition{Path: "apply_fulfillment_result", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ApplyFulfillmentResult}).
			RegisterRoutes(route)
	})
}

func initStockMoveV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveRest) error {
		return composable.NewRestEngine(models.StockMoveSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockMoveLineV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveLineRest) error {
		return composable.NewRestEngine(models.StockMoveLineSchemaName, rest).
			AddCrudRoutes(composable.ReadCrudActions()...).
			RegisterRoutes(route)
	})
}

func initStockMoveDependencyV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockMoveDependencyRest) error {
		return composable.NewRestEngine(models.StockMoveDependencySchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initStockScrapV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockScrapRest) error {
		return composable.NewRestEngine(models.StockScrapSchemaName, rest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{Path: ":id/do_scrap", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.DoScrap}).
			RegisterRoutes(route)
	})
}

func initStockProductConfigV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockProductConfigRest) error {
		return composable.NewRestEngine(models.StockProductConfigSchemaName, rest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// initStockReservationV1 serves the reads only: a reservation is written by its operations, which
// register as custom routes as they land, never by the built-in create, update or delete.
func initStockReservationV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.StockReservationRest) error {
		return composable.NewRestEngine(models.StockReservationSchemaName, rest).
			AddCrudRoutes(
				composable.CrudActionGetById,
				composable.CrudActionGetByUnique,
				composable.CrudActionSearch,
				composable.CrudActionExists,
				composable.CrudActionGetSchema,
				composable.CrudActionComputeField,
			).
			AddRoute(composable.RouteDefinition{Path: "reserve_warehouse_stock", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReserveWarehouseStock}).
			AddRoute(composable.RouteDefinition{Path: "check_warehouse_availability", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.CheckWarehouseAvailability}).
			AddRoute(composable.RouteDefinition{Path: ":id/consume", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ConsumeReservation}).
			AddRoute(composable.RouteDefinition{Path: ":id/release", ActionType: composable.ActionTypeGeneric, HandlerFn: rest.ReleaseReservation}).
			RegisterRoutes(route)
	})
}
