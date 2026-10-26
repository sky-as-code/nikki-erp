package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/constants"
	"github.com/sky-as-code/nikki-erp/modules/commerce/purchase/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/commerce/purchase/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// Every Purchase resource declares its routes through the composable RestEngine: the built-in CRUD
// table the legacy engine served, plus one AddRoute per custom action, on the paths the legacy
// actions used.

const routeV1Prefix = "/v1/purchase"

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewConfigurationRest,
		v1.NewVendorProductPriceRest,
		v1.NewSourcingGroupRest,
		v1.NewAgreementRest,
		v1.NewAgreementLineRest,
		v1.NewPurchaseOrderRest,
		v1.NewPurchaseOrderLineRest,
		v1.NewAuditEventRest,
	)
	return stdErr.Join(err, initPurchaseV1())
}

func initPurchaseV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		configurations *v1.ConfigurationRest,
		vendorPrices *v1.VendorProductPriceRest,
		sourcingGroups *v1.SourcingGroupRest,
		agreements *v1.AgreementRest,
		agreementLines *v1.AgreementLineRest,
		orders *v1.PurchaseOrderRest,
		orderLines *v1.PurchaseOrderLineRest,
		auditEvents *v1.AuditEventRest,
	) error {
		routeV1 := route.Group(routeV1Prefix)
		return stdErr.Join(
			crudRoutes(models.ConfigurationSchemaName, configurations).RegisterRoutes(routeV1),
			crudRoutes(models.VendorProductPriceSchemaName, vendorPrices).RegisterRoutes(routeV1),
			crudRoutes(models.SourcingGroupSchemaName, sourcingGroups).RegisterRoutes(routeV1),
			agreementRoutes(agreements).RegisterRoutes(routeV1),
			crudRoutes(models.AgreementLineSchemaName, agreementLines).RegisterRoutes(routeV1),
			purchaseOrderRoutes(orders).RegisterRoutes(routeV1),
			crudRoutes(models.PurchaseOrderLineSchemaName, orderLines).RegisterRoutes(routeV1),
			crudRoutes(models.AuditEventSchemaName, auditEvents).RegisterRoutes(routeV1),
		)
	})
}

// legacyCrudActions is the built-in verb set the legacy engine served: every CRUD action except
// bulk create and import, which these resources never exposed. The resources a client may not
// write keep their write routes, which answer a business violation naming the reason.
func legacyCrudActions() []composable.CrudAction {
	return []composable.CrudAction{
		composable.CrudActionCreate, composable.CrudActionUpdate, composable.CrudActionDelete,
		composable.CrudActionSetArchived, composable.CrudActionGetById, composable.CrudActionSearch,
		composable.CrudActionExists, composable.CrudActionGetSchema, composable.CrudActionComputeField,
	}
}

func crudRoutes(schemaName string, rest composable.CrudRestHandlers) composable.RestEngine {
	return composable.NewRestEngine(schemaName, rest).AddCrudRoutes(legacyCrudActions()...)
}

func genericRoute(path string, handler composable.HandlerFn) composable.RouteDefinition {
	return composable.RouteDefinition{Path: path, ActionType: composable.ActionTypeGeneric, HandlerFn: handler}
}

// purchaseOrderRoutes adds the lifecycle actions. merge is collection-level: it names several
// orders in the body, so it is registered before the ":id" routes.
func purchaseOrderRoutes(rest *v1.PurchaseOrderRest) composable.RestEngine {
	return crudRoutes(models.PurchaseOrderSchemaName, rest).
		AddRoute(genericRoute(constants.ActionMerge, rest.Merge)).
		AddRoute(genericRoute(":id/"+constants.ActionConfirm, rest.Confirm)).
		AddRoute(genericRoute(":id/"+constants.ActionApprove, rest.Approve)).
		AddRoute(genericRoute(":id/"+constants.ActionCancel, rest.Cancel)).
		AddRoute(genericRoute(":id/"+constants.ActionSend, rest.Send)).
		AddRoute(genericRoute(":id/"+constants.ActionLock, rest.Lock)).
		AddRoute(genericRoute(":id/"+constants.ActionUnlock, rest.Unlock)).
		AddRoute(genericRoute(":id/"+constants.ActionAcknowledge, rest.Acknowledge)).
		AddRoute(genericRoute(":id/"+constants.ActionCreateAlternative, rest.CreateAlternative)).
		AddRoute(genericRoute(":id/"+constants.ActionCompareAlternatives, rest.CompareAlternatives)).
		AddRoute(genericRoute(":id/"+constants.ActionReprice, rest.Reprice)).
		AddRoute(genericRoute(":id/"+constants.ActionDuplicate, rest.Duplicate))
}

func agreementRoutes(rest *v1.AgreementRest) composable.RestEngine {
	return crudRoutes(models.AgreementSchemaName, rest).
		AddRoute(genericRoute(":id/"+constants.ActionConfirm, rest.Confirm)).
		AddRoute(genericRoute(":id/"+constants.ActionClose, rest.Close)).
		AddRoute(genericRoute(":id/"+constants.ActionCancel, rest.Cancel)).
		AddRoute(genericRoute(":id/"+constants.ActionCreateRfq, rest.CreateRfq))
}
