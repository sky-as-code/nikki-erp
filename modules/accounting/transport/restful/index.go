package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	modconstants "github.com/sky-as-code/nikki-erp/modules/accounting/constants"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	v1 "github.com/sky-as-code/nikki-erp/modules/accounting/transport/restful/v1"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func InitRestfulHandlers() error {
	return stdErr.Join(
		deps.Register(
			v1.NewTaxCalculationRest,
			v1.NewTaxJurisdictionRest,
			v1.NewTaxGroupRest,
			v1.NewTaxRoundingPolicyRest,
			v1.NewTaxProductClassificationRest,
			v1.NewTaxRest,
			v1.NewTaxDefinitionVersionRest,
			v1.NewTaxRateVersionRest,
			v1.NewTaxComponentRest,
			v1.NewTaxMappingRest,
			v1.NewTaxMappingLineRest,
			v1.NewTaxRuleRest,
			v1.NewTaxRuleConditionRest,
			v1.NewTaxRuleResultRest,
		),
		initAccountingV1(),
		initSettingsV1(),
	)
}

type taxConfigRests struct {
	dig.In

	TaxJurisdiction          *v1.TaxJurisdictionRest
	TaxGroup                 *v1.TaxGroupRest
	TaxRoundingPolicy        *v1.TaxRoundingPolicyRest
	TaxProductClassification *v1.TaxProductClassificationRest
	Tax                      *v1.TaxRest
	TaxDefinitionVersion     *v1.TaxDefinitionVersionRest
	TaxRateVersion           *v1.TaxRateVersionRest
	TaxComponent             *v1.TaxComponentRest
	TaxMapping               *v1.TaxMappingRest
	TaxMappingLine           *v1.TaxMappingLineRest
	TaxRule                  *v1.TaxRuleRest
	TaxRuleCondition         *v1.TaxRuleConditionRest
	TaxRuleResult            *v1.TaxRuleResultRest
}

// initAccountingV1 serves every tax configuration resource at /v1/accounting/{schema_name} with
// the built-in route table.
func initAccountingV1() error {
	return deps.Invoke(func(route *echo.Group, rests taxConfigRests) error {
		routeV1 := route.Group(modconstants.AccountingRouteV1)
		return stdErr.Join(
			registerCrud(routeV1, models.TaxJurisdictionSchemaName, rests.TaxJurisdiction),
			registerCrud(routeV1, models.TaxGroupSchemaName, rests.TaxGroup),
			registerCrud(routeV1, models.TaxRoundingPolicySchemaName, rests.TaxRoundingPolicy),
			registerCrud(routeV1, models.TaxProductClassificationSchemaName, rests.TaxProductClassification),
			registerCrud(routeV1, models.TaxDefinitionVersionSchemaName, rests.TaxDefinitionVersion),
			registerCrud(routeV1, models.TaxRateVersionSchemaName, rests.TaxRateVersion),
			registerCrud(routeV1, models.TaxComponentSchemaName, rests.TaxComponent),
			registerCrud(routeV1, models.TaxMappingSchemaName, rests.TaxMapping),
			registerCrud(routeV1, models.TaxMappingLineSchemaName, rests.TaxMappingLine),
			registerCrud(routeV1, models.TaxRuleSchemaName, rests.TaxRule),
			registerCrud(routeV1, models.TaxRuleConditionSchemaName, rests.TaxRuleCondition),
			registerCrud(routeV1, models.TaxRuleResultSchemaName, rests.TaxRuleResult),
			initTaxV1(routeV1),
		)
	})
}

func registerCrud(route *echo.Group, schemaName string, handlers composable.CrudRestHandlers) error {
	return composable.NewRestEngine(schemaName, handlers).AddCrudRoutes().RegisterRoutes(route)
}

// initTaxV1 serves accounting_tax: the built-in CRUD plus the four calculation actions. They are
// collection-level (a calculation prices a document, not a tax record) and POST throughout,
// including the two that only read, because the request is a whole document of lines and party
// context that does not fit a query string.
func initTaxV1(routeV1 *echo.Group) error {
	return deps.Invoke(func(taxRest *v1.TaxRest, calcRest *v1.TaxCalculationRest) error {
		return composable.NewRestEngine(models.TaxSchemaName, taxRest).
			AddCrudRoutes().
			AddRoute(taxAction("calculate", calcRest.Calculate)).
			AddRoute(taxAction("simulate", calcRest.Simulate)).
			AddRoute(taxAction("reverse_full", calcRest.ReverseFull)).
			AddRoute(taxAction("reverse_partial", calcRest.ReversePartial)).
			RegisterRoutes(routeV1)
	})
}

// TaxCalculationActionPaths are the collection actions initTaxV1 adds to accounting_tax.
var TaxCalculationActionPaths = []string{"calculate", "simulate", "reverse_full", "reverse_partial"}

func taxAction(path string, handler composable.HandlerFn) composable.RouteDefinition {
	return composable.RouteDefinition{Path: path, ActionType: composable.ActionTypeGeneric, HandlerFn: handler}
}
