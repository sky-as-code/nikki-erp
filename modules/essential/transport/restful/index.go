package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	m "github.com/sky-as-code/nikki-erp/modules/core/httpserver/middlewares"
	v1 "github.com/sky-as-code/nikki-erp/modules/essential/transport/restful/v1"
)

func InitRestfulHandlers() error {
	err := deps.Register(
		v1.NewContactRest,
		v1.NewFieldMetadataRest,
		v1.NewLanguageRest,
		v1.NewModelMetadataRest,
		v1.NewModuleRest,
		v1.NewUnitRest,
		v1.NewUnitCategoryRest,
	)
	return stdErr.Join(
		err,
		initContactV1(),
		initEssentialV1(),
		initFieldMetadataV1(),
		initLanguageV1(),
		initModelMetadataV1(),
		initUnitV1(),
	)
}

func initContactV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		contactRest *v1.ContactRest,
	) {
		routeV1 := route.Group("/v1/essential")

		routeV1.DELETE("/contacts/:id", contactRest.DeleteContact, m.SmokeAuthz())
		routeV1.GET("/contacts/:id", contactRest.GetContact, m.SmokeAuthz())
		routeV1.GET("/contacts", contactRest.SearchContacts, m.SmokeAuthz())
		routeV1.POST("/contacts/exists", contactRest.ContactExists, m.SmokeAuthz())
		routeV1.POST("/contacts", contactRest.CreateContact, m.SmokeAuthz())
		routeV1.PUT("/contacts/:id", contactRest.UpdateContact, m.SmokeAuthz())
	})
}

func initEssentialV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		moduleRest *v1.ModuleRest,
	) {
		routeV1 := route.Group("/v1/essential")

		routeV1.DELETE("/modules/:id", moduleRest.DeleteModule, m.SmokeAuthz())
		routeV1.GET("/modules/meta/schema", moduleRest.GetModelSchema, m.SmokeAuthz())
		routeV1.GET("/modules/:id", moduleRest.GetModule, m.SmokeAuthz())
		routeV1.GET("/modules", moduleRest.SearchModules, m.SmokeAuthz())
		routeV1.POST("/modules/exists", moduleRest.ModuleExists, m.SmokeAuthz())
		routeV1.POST("/modules", moduleRest.CreateModule, m.SmokeAuthz())
		routeV1.PUT("/modules/:id", moduleRest.UpdateModule, m.SmokeAuthz())
	})
}

func initFieldMetadataV1() error {
	return deps.Invoke(func(route *echo.Group, fieldMetadataRest *v1.FieldMetadataRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/field-metadata/:id", fieldMetadataRest.DeleteFieldMetadata, m.SmokeAuthz())
		routeV1.GET("/field-metadata/:id", fieldMetadataRest.GetFieldMetadata, m.SmokeAuthz())
		routeV1.GET("/field-metadata", fieldMetadataRest.SearchFieldMetadata, m.SmokeAuthz())
		routeV1.POST("/field-metadata/exists", fieldMetadataRest.FieldMetadataExists, m.SmokeAuthz())
		routeV1.POST("/field-metadata", fieldMetadataRest.CreateFieldMetadata, m.SmokeAuthz())
		routeV1.PUT("/field-metadata/:id", fieldMetadataRest.UpdateFieldMetadata, m.SmokeAuthz())
	})
}

func initLanguageV1() error {
	return deps.Invoke(func(route *echo.Group, languageRest *v1.LanguageRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/languages/:id", languageRest.DeleteLanguage, m.SmokeAuthz())
		routeV1.GET("/languages/:id", languageRest.GetLanguage, m.SmokeAuthz())
		routeV1.GET("/languages", languageRest.SearchLanguages, m.SmokeAuthz())
		routeV1.POST("/languages/exists", languageRest.LanguageExists, m.SmokeAuthz())
		routeV1.POST("/languages", languageRest.CreateLanguage, m.SmokeAuthz())
		routeV1.PUT("/languages/:id", languageRest.UpdateLanguage, m.SmokeAuthz())
	})
}

func initModelMetadataV1() error {
	return deps.Invoke(func(route *echo.Group, modelMetadataRest *v1.ModelMetadataRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/model-metadata/:id", modelMetadataRest.DeleteModelMetadata, m.SmokeAuthz())
		routeV1.GET("/model-metadata/:id", modelMetadataRest.GetModelMetadata, m.SmokeAuthz())
		routeV1.GET("/model-metadata", modelMetadataRest.SearchModelMetadata, m.SmokeAuthz())
		routeV1.POST("/model-metadata/exists", modelMetadataRest.ModelMetadataExists, m.SmokeAuthz())
		routeV1.POST("/model-metadata", modelMetadataRest.CreateModelMetadata, m.SmokeAuthz())
		routeV1.PUT("/model-metadata/:id", modelMetadataRest.UpdateModelMetadata, m.SmokeAuthz())
	})
}

func initUnitV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		unitRest *v1.UnitRest,
		unitCategoryRest *v1.UnitCategoryRest,
	) {
		routeV1 := route.Group("/v1/:org_id/inventory")

		routeV1.DELETE("/units/:id", unitRest.Delete, m.SmokeAuthz())
		routeV1.GET("/units/:id", unitRest.GetOne, m.SmokeAuthz())
		routeV1.POST("/units/:id/exists", unitRest.Exists, m.SmokeAuthz())
		routeV1.POST("/units/:id", unitRest.Create, m.SmokeAuthz())
		routeV1.PUT("/units/:id", unitRest.Update, m.SmokeAuthz())

		routeV1.DELETE("/units-categories/:id", unitCategoryRest.Delete, m.SmokeAuthz())
		routeV1.GET("/units-categories/:id", unitCategoryRest.GetOne, m.SmokeAuthz())
		routeV1.POST("/units-categories/:id/exists", unitCategoryRest.Exists, m.SmokeAuthz())
		routeV1.POST("/units-categories/:id", unitCategoryRest.Create, m.SmokeAuthz())
		routeV1.PUT("/units-categories/:id", unitCategoryRest.Update, m.SmokeAuthz())
	})
}
