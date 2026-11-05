package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/commerce/sales/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The promotion routes (plan §3.2). Custom actions are added before the CRUD set's :id routes
// cannot shadow them, and every path segment is snake_case.

// initSalesPromotionReadOnlyV1 serves the read-only promotion resources: the selectors and the
// records of what a confirmed order received.
func initSalesPromotionReadOnlyV1(route *echo.Group) error {
	schemaNames := []string{
		models.SalesPromotionRuleProductSchemaName,
		models.SalesPromotionRuleCategorySchemaName,
		models.SalesPromotionRewardProductSchemaName,
		models.SalesPromotionRewardCategorySchemaName,
		models.SalesOrderVoucherSchemaName,
		models.SalesOrderPromotionSelectionSchemaName,
		models.SalesPromotionApplicationSchemaName,
		models.SalesPromotionAllocationSchemaName,
		models.SalesVoucherIssuanceSchemaName,
	}
	errs := make([]error, 0, len(schemaNames))
	for _, schemaName := range schemaNames {
		errs = append(errs, readOnlyPromotionRoutes(route, schemaName))
	}
	return stdErr.Join(errs...)
}

func initSalesPromotionV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionRest) error {
		engine := dynamicresource.NewRestEngine(models.SalesPromotionSchemaName, rest).AddCrudRoutes()
		for path, handler := range map[string]dynamicresource.HandlerFn{
			":id/activate": rest.Activate, ":id/suspend": rest.Suspend, ":id/duplicate": rest.Duplicate,
			":id/manage_channels": rest.ManageChannels, ":id/manage_points": rest.ManagePoints,
			":id/manage_pricelists": rest.ManagePricelists,
		} {
			engine = engine.AddRoute(dynamicresource.RouteDefinition{
				Path: path, ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: handler,
			})
		}
		return engine.RegisterRoutes(route)
	})
}

func initSalesPromotionRuleV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionRuleRest) error {
		return selectorOwnerRoutes(models.SalesPromotionRuleSchemaName, rest.SalesPromotionSelectorOwnerRest).
			RegisterRoutes(route)
	})
}

func initSalesPromotionRewardV1(route *echo.Group) error {
	return deps.Invoke(func(rest *v1.SalesPromotionRewardRest) error {
		return selectorOwnerRoutes(models.SalesPromotionRewardSchemaName, rest.SalesPromotionSelectorOwnerRest).
			RegisterRoutes(route)
	})
}

func selectorOwnerRoutes(schemaName string, rest *v1.SalesPromotionSelectorOwnerRest) dynamicresource.RestEngine {
	return dynamicresource.NewRestEngine(schemaName, rest).
		AddCrudRoutes().
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_products", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageProducts,
		}).
		AddRoute(dynamicresource.RouteDefinition{
			Path: ":id/manage_categories", ActionType: dynamicresource.ActionTypeGeneric, HandlerFn: rest.ManageCategories,
		})
}

func readOnlyPromotionRoutes(route *echo.Group, schemaName string) error {
	engines, err := dynamicresource.ResolveEngines(schemaName)
	if err != nil {
		return err
	}
	rest := &dynamicresource.CrudRestBase{}
	rest.SetApplicationService(engines[0].ApplicationService())
	return dynamicresource.NewRestEngine(schemaName, rest).
		AddCrudRoutes(readOnlyRoutes()...).
		RegisterRoutes(route)
}
