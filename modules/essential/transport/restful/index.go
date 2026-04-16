package restful

import (
	stdErr "errors"

	"github.com/labstack/echo/v5"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
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

		routeV1.DELETE("/contacts/:id", contactRest.DeleteContact)
		routeV1.GET("/contacts/:id", contactRest.GetContact)
		routeV1.GET("/contacts", contactRest.SearchContacts)
		routeV1.POST("/contacts/exists", contactRest.ContactExists)
		routeV1.POST("/contacts", contactRest.CreateContact)
		routeV1.PUT("/contacts/:id", contactRest.UpdateContact)
	})
}

func initEssentialV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		moduleRest *v1.ModuleRest,
	) {
		routeV1 := route.Group("/v1/essential")

		routeV1.DELETE("/modules/:id", moduleRest.DeleteModule)
		routeV1.GET("/modules/:id", moduleRest.GetModule)
		routeV1.GET("/modules", moduleRest.SearchModules)
		routeV1.POST("/modules/exists", moduleRest.ModuleExists)
		routeV1.POST("/modules", moduleRest.CreateModule)
		routeV1.PUT("/modules/:id", moduleRest.UpdateModule)
	})
}

func initUnitV1() error {
	return deps.Invoke(func(
		route *echo.Group,
		unitRest *v1.UnitRest,
		unitCategoryRest *v1.UnitCategoryRest,
	) {
		routeV1 := route.Group("/v1/:org_id/inventory")

		routeV1.DELETE("/units/:id", unitRest.Delete)
		routeV1.GET("/units/:id", unitRest.GetOne)
		routeV1.POST("/units/:id/exists", unitRest.Exists)
		routeV1.POST("/units/:id", unitRest.Create)
		routeV1.PUT("/units/:id", unitRest.Update)

		routeV1.DELETE("/units-categories/:id", unitCategoryRest.Delete)
		routeV1.GET("/units-categories/:id", unitCategoryRest.GetOne)
		routeV1.POST("/units-categories/:id/exists", unitCategoryRest.Exists)
		routeV1.POST("/units-categories/:id", unitCategoryRest.Create)
		routeV1.PUT("/units-categories/:id", unitCategoryRest.Update)
	})
}

func initModelMetadataV1() error {
	return deps.Invoke(func(route *echo.Group, modelMetadataRest *v1.ModelMetadataRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/model-metadata/:id", modelMetadataRest.DeleteModelMetadata)
		routeV1.GET("/model-metadata/:id", modelMetadataRest.GetModelMetadata)
		routeV1.GET("/model-metadata", modelMetadataRest.SearchModelMetadata)
		routeV1.POST("/model-metadata/exists", modelMetadataRest.ModelMetadataExists)
		routeV1.POST("/model-metadata", modelMetadataRest.CreateModelMetadata)
		routeV1.PUT("/model-metadata/:id", modelMetadataRest.UpdateModelMetadata)
	})
}

func initFieldMetadataV1() error {
	return deps.Invoke(func(route *echo.Group, fieldMetadataRest *v1.FieldMetadataRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/field-metadata/:id", fieldMetadataRest.DeleteFieldMetadata)
		routeV1.GET("/field-metadata/:id", fieldMetadataRest.GetFieldMetadata)
		routeV1.GET("/field-metadata", fieldMetadataRest.SearchFieldMetadata)
		routeV1.POST("/field-metadata/exists", fieldMetadataRest.FieldMetadataExists)
		routeV1.POST("/field-metadata", fieldMetadataRest.CreateFieldMetadata)
		routeV1.PUT("/field-metadata/:id", fieldMetadataRest.UpdateFieldMetadata)
	})
}

func initLanguageV1() error {
	return deps.Invoke(func(route *echo.Group, languageRest *v1.LanguageRest) {
		routeV1 := route.Group("/v1/essential")
		routeV1.DELETE("/languages/:id", languageRest.DeleteLanguage)
		routeV1.GET("/languages/:id", languageRest.GetLanguage)
		routeV1.GET("/languages", languageRest.SearchLanguages)
		routeV1.POST("/languages/exists", languageRest.LanguageExists)
		routeV1.POST("/languages", languageRest.CreateLanguage)
		routeV1.PUT("/languages/:id", languageRest.UpdateLanguage)
	})
}
