package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/dynamicengines"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		// v1.NewContactRest,
		v1.NewEnumRest,
		v1.NewFieldMetadataRest,
		v1.NewLanguageRest,
		v1.NewModelMetadataRest,
		v1.NewModuleRest,
		v1.NewTagRest,
		v1.NewUomConversionRest,
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

		// UoM and UoM Category are served entirely by the dynamic resource engine, at
		// /v1/essential/{schema_name}. They have no hand-written REST layer.
		registerEngineRoutes(routeV1)

		return stdErr.Join(
			initEnumV1(routeV1),
			initFieldMetadataV1(routeV1),
			initLanguageV1(routeV1),
			initModelMetadataV1(routeV1),
			initModuleV1(routeV1),
			initTagV1(routeV1),
			initUomConversionV1(routeV1),
		)
	})
}

// registerEngineRoutes exposes every Essential resource engine over HTTP.
// A missing engine is skipped, so that a build which drops one still starts.
func registerEngineRoutes(routeV1 *echo.Group) {
	for _, schemaName := range dynamicengines.EngineSchemaNames() {
		engine, exists := dynamicresource.Registry().GetEngine(schemaName)
		if !exists {
			continue
		}
		engine.RestApi().RegisterRoutes(routeV1, m.SmokeAuthz())
	}
}

func initEnumV1(route *echo.Group) error {
	return deps.Invoke(func(
		enumRest *v1.EnumRest,
	) {
		route.DELETE("/enums/:id", enumRest.Delete)
		route.GET("/enums/:id", enumRest.GetOne)
		route.GET("/enums", enumRest.Search)
		route.POST("/enums/exists", enumRest.Exists)
		route.POST("/enums", enumRest.Create)
		route.PUT("/enums/:id", enumRest.Update)
	})
}

func initFieldMetadataV1(route *echo.Group) error {
	return deps.Invoke(func(
		fieldMetadataRest *v1.FieldMetadataRest,
	) {
		route.DELETE("/field-metadata/:id", fieldMetadataRest.DeleteFieldMetadata)
		route.GET("/field-metadata/:id", fieldMetadataRest.GetFieldMetadata)
		route.GET("/field-metadata", fieldMetadataRest.SearchFieldMetadata)
		route.POST("/field-metadata/exists", fieldMetadataRest.FieldMetadataExists)
		route.POST("/field-metadata", fieldMetadataRest.CreateFieldMetadata)
		route.PUT("/field-metadata/:id", fieldMetadataRest.UpdateFieldMetadata)
	})
}

func initLanguageV1(route *echo.Group) error {
	return deps.Invoke(func(
		languageRest *v1.LanguageRest,
	) {
		route.DELETE("/languages/:id", languageRest.DeleteLanguage)
		route.GET("/languages/json", languageRest.GetLanguageJson)
		route.GET("/languages/:id", languageRest.GetLanguage)
		route.GET("/languages", languageRest.SearchLanguages)
		route.POST("/languages/exists", languageRest.LanguageExists)
		route.POST("/languages", languageRest.CreateLanguage)
		route.PUT("/languages/:id", languageRest.UpdateLanguage)
	})
}

func initModelMetadataV1(route *echo.Group) error {
	return deps.Invoke(func(
		modelMetadataRest *v1.ModelMetadataRest,
	) {
		route.DELETE("/model-metadata/:id", modelMetadataRest.DeleteModelMetadata)
		route.GET("/model-metadata/:id", modelMetadataRest.GetModelMetadata)
		route.GET("/model-metadata", modelMetadataRest.SearchModelMetadata)
		route.POST("/model-metadata/exists", modelMetadataRest.ModelMetadataExists)
		route.POST("/model-metadata", modelMetadataRest.CreateModelMetadata)
		route.PUT("/model-metadata/:id", modelMetadataRest.UpdateModelMetadata)
	})
}

func initModuleV1(route *echo.Group) error {
	return deps.Invoke(func(
		moduleRest *v1.ModuleRest,
	) {
		route.DELETE("/modules/:id", moduleRest.DeleteModule)
		route.GET("/modules/meta/schema", moduleRest.GetModelSchema)
		route.GET("/modules/:id", moduleRest.GetModule)
		route.GET("/modules", moduleRest.SearchModules)
		route.POST("/modules/exists", moduleRest.ModuleExists)
		route.POST("/modules", moduleRest.CreateModule)
		route.PUT("/modules/:id", moduleRest.UpdateModule)
	})
}

func initTagV1(route *echo.Group) error {
	return deps.Invoke(func(
		tagRest *v1.TagRest,
	) {
		route.DELETE("/tags/:id", tagRest.Delete)
		route.GET("/tags/:id", tagRest.GetOne)
		route.GET("/tags", tagRest.Search)
		route.POST("/tags/exists", tagRest.Exists)
		route.POST("/tags", tagRest.Create)
		route.PUT("/tags/:id", tagRest.Update)
	})
}

// initUomConversionV1 registers the one UoM endpoint the resource engine cannot express:
// conversion is a calculation over two records, not a CRUD action on one.
func initUomConversionV1(route *echo.Group) error {
	return deps.Invoke(func(
		uomConversionRest *v1.UomConversionRest,
	) {
		route.POST("/uoms/convert", uomConversionRest.Convert, m.SmokeAuthz())
		route.POST("/uoms/to_reference", uomConversionRest.ToReference, m.SmokeAuthz())
	})
}
