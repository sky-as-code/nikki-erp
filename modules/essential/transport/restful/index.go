package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	coreSettings "github.com/sky-as-code/nikki-erp/modules/core/settings"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/dynamicengines"
	itSettings "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/settings"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		// v1.NewContactRest,
		v1.NewCurrencyRest,
		v1.NewEnumRest,
		v1.NewFieldMetadataRest,
		v1.NewLanguageRest,
		v1.NewModelMetadataRest,
		v1.NewModuleRest,
		v1.NewTagRest,
		v1.NewUomCatRest,
		v1.NewUomConversionRest,
		v1.NewUomRest,
	)
	err = stdErr.Join(
		err,
		initEssentialV1(),
	)
	return err
}

func initEssentialV1() error {
	return deps.Invoke(func(
		route *echo.Group,
	) error {
		routeV1 := route.Group("/v1/essential")

		return stdErr.Join(
			// Composable resources are served at /v1/essential/{schema_name}.
			initCurrencyV1(routeV1),
			initUomV1(routeV1),
			initUomCatV1(routeV1),
			initEnumV1(routeV1),
			initFieldMetadataV1(routeV1),
			initLanguageV1(routeV1),
			initModelMetadataV1(routeV1),
			initModuleV1(routeV1),
			initTagV1(routeV1),
			initSettingsV1(routeV1),
		)
	})
}

func initCurrencyV1(route *echo.Group) error {
	return deps.Invoke(func(currencyRest *v1.CurrencyRest) error {
		return composable.NewRestEngine(models.CurrencySchemaName, currencyRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// initUomV1 serves the UoM CRUD plus the two conversion actions (BR-UOM-ESS-013): a
// calculation over two records rather than an action on one, so they are collection-level.
func initUomV1(route *echo.Group) error {
	return deps.Invoke(func(uomRest *v1.UomRest, conversionRest *v1.UomConversionRest) error {
		return composable.NewRestEngine(models.UomSchemaName, uomRest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "convert", ActionType: composable.ActionTypeGeneric, HandlerFn: conversionRest.Convert,
			}).
			AddRoute(composable.RouteDefinition{
				Path: "to_reference", ActionType: composable.ActionTypeGeneric, HandlerFn: conversionRest.ToReference,
			}).
			RegisterRoutes(route)
	})
}

func initEnumV1(route *echo.Group) error {
	return deps.Invoke(func(enumRest *v1.EnumRest) error {
		return composable.NewRestEngine(models.EnumSchemaName, enumRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initTagV1(route *echo.Group) error {
	return deps.Invoke(func(tagRest *v1.TagRest) error {
		return composable.NewRestEngine(models.TagSchemaName, tagRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// initLanguageV1 serves the catalogue at /v1/essential/essential_language. The translation
// bundles (json) and the enabled list are public: they render the sign-in page, before anyone has
// a session.
func initLanguageV1(route *echo.Group) error {
	return deps.Invoke(func(languageRest *v1.LanguageRest) error {
		return composable.NewRestEngine(models.LanguageSchemaName, languageRest).
			AddCrudRoutes().
			AddRoute(composable.RouteDefinition{
				Path: "json", ActionType: composable.ActionTypeRead, HandlerFn: languageRest.GetLanguageJson,
				IsAuthorized: util.ToPtr(false),
			}).
			AddRoute(composable.RouteDefinition{
				Path: "enabled", ActionType: composable.ActionTypeRead, HandlerFn: languageRest.ListEnabledLanguages,
				IsAuthorized: util.ToPtr(false),
			}).
			RegisterRoutes(route)
	})
}

func initFieldMetadataV1(route *echo.Group) error {
	return deps.Invoke(func(fieldMetadataRest *v1.FieldMetadataRest) error {
		return composable.NewRestEngine(models.FieldMetadataSchemaName, fieldMetadataRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initModelMetadataV1(route *echo.Group) error {
	return deps.Invoke(func(modelMetadataRest *v1.ModelMetadataRest) error {
		return composable.NewRestEngine(models.ModelMetadataSchemaName, modelMetadataRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

// initModuleV1 serves the module catalogue read-only at /v1/essential/essential_module_metadata.
func initModuleV1(route *echo.Group) error {
	return deps.Invoke(func(moduleRest *v1.ModuleRest) error {
		return composable.NewRestEngine(models.ModuleMetadataSchemaName, moduleRest).
			AddCrudRoutes(dynamicengines.ModuleReadActions()...).
			RegisterRoutes(route)
	})
}

func initUomCatV1(route *echo.Group) error {
	return deps.Invoke(func(uomCatRest *v1.UomCatRest) error {
		return composable.NewRestEngine(models.UomCatSchemaName, uomCatRest).
			AddCrudRoutes().
			RegisterRoutes(route)
	})
}

func initSettingsV1(route *echo.Group) error {
	return deps.Invoke(func(settingsSvc itSettings.SettingsAppService) {
		route.PATCH(coreSettings.UpdateSettingsRoutePath, coreSettings.UpdateSettingsHandler(settingsSvc.UpdateSettings),
			m.UserSmokeAuthz())
	})
}
