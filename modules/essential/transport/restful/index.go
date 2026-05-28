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
		// v1.NewContactRest,
		v1.NewEnumRest,
		v1.NewFieldMetadataRest,
		v1.NewLanguageRest,
		v1.NewModelMetadataRest,
		v1.NewModuleRest,
		v1.NewTagRest,
		v1.NewUnitRest,
		v1.NewUnitCategoryRest,
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
		// contactRest *v1.ContactRest,
		enumRest *v1.EnumRest,
		unitRest *v1.UnitRest,
		unitCategoryRest *v1.UnitCategoryRest,
		fieldMetadataRest *v1.FieldMetadataRest,
		languageRest *v1.LanguageRest,
		modelMetadataRest *v1.ModelMetadataRest,
		moduleRest *v1.ModuleRest,
		tagRest *v1.TagRest,
	) error {
		routeV1 := route.Group("/v1/essential")

		// routeV1.DELETE("/contacts/:id", contactRest.DeleteContact)
		// routeV1.GET("/contacts/:id", contactRest.GetContact)
		// routeV1.GET("/contacts", contactRest.SearchContacts)
		// routeV1.POST("/contacts/exists", contactRest.ContactExists)
		// routeV1.POST("/contacts", contactRest.CreateContact)
		// routeV1.PUT("/contacts/:id", contactRest.UpdateContact)

		routeV1.DELETE("/enums/:id", enumRest.Delete)
		routeV1.GET("/enums/:id", enumRest.GetOne)
		routeV1.GET("/enums", enumRest.Search)
		routeV1.POST("/enums/exists", enumRest.Exists)
		routeV1.POST("/enums", enumRest.Create)
		routeV1.PUT("/enums/:id", enumRest.Update)

		routeV1.DELETE("/field-metadata/:id", fieldMetadataRest.DeleteFieldMetadata)
		routeV1.GET("/field-metadata/:id", fieldMetadataRest.GetFieldMetadata)
		routeV1.GET("/field-metadata", fieldMetadataRest.SearchFieldMetadata)
		routeV1.POST("/field-metadata/exists", fieldMetadataRest.FieldMetadataExists)
		routeV1.POST("/field-metadata", fieldMetadataRest.CreateFieldMetadata)
		routeV1.PUT("/field-metadata/:id", fieldMetadataRest.UpdateFieldMetadata)

		routeV1.DELETE("/languages/:id", languageRest.DeleteLanguage)
		routeV1.GET("/languages/:id", languageRest.GetLanguage)
		routeV1.GET("/languages", languageRest.SearchLanguages)
		routeV1.POST("/languages/exists", languageRest.LanguageExists)
		routeV1.POST("/languages", languageRest.CreateLanguage)
		routeV1.PUT("/languages/:id", languageRest.UpdateLanguage)

		routeV1.DELETE("/model-metadata/:id", modelMetadataRest.DeleteModelMetadata)
		routeV1.GET("/model-metadata/:id", modelMetadataRest.GetModelMetadata)
		routeV1.GET("/model-metadata", modelMetadataRest.SearchModelMetadata)
		routeV1.POST("/model-metadata/exists", modelMetadataRest.ModelMetadataExists)
		routeV1.POST("/model-metadata", modelMetadataRest.CreateModelMetadata)
		routeV1.PUT("/model-metadata/:id", modelMetadataRest.UpdateModelMetadata)

		routeV1.DELETE("/modules/:id", moduleRest.DeleteModule)
		routeV1.GET("/modules/:id", moduleRest.GetModule)
		routeV1.GET("/modules", moduleRest.SearchModules)
		routeV1.POST("/modules/exists", moduleRest.ModuleExists)
		routeV1.POST("/modules", moduleRest.CreateModule)
		routeV1.PUT("/modules/:id", moduleRest.UpdateModule)

		routeV1.DELETE("/tags/:id", tagRest.Delete)
		routeV1.GET("/tags/:id", tagRest.GetOne)
		routeV1.GET("/tags", tagRest.Search)
		routeV1.POST("/tags/exists", tagRest.Exists)
		routeV1.POST("/tags", tagRest.Create)
		routeV1.PUT("/tags/:id", tagRest.Update)

		routeV1.DELETE("/units/:id", unitRest.Delete, m.SmokeAuthz())
		routeV1.GET("/units/:id", unitRest.GetOne, m.SmokeAuthz())
		routeV1.POST("/units/:id/exists", unitRest.Exists, m.SmokeAuthz())
		routeV1.POST("/units/:id", unitRest.Create, m.SmokeAuthz())
		routeV1.PUT("/units/:id", unitRest.Update, m.SmokeAuthz())

		routeV1.DELETE("/units-categories/:id", unitCategoryRest.Delete)
		routeV1.GET("/units-categories/:id", unitCategoryRest.GetOne)
		routeV1.POST("/units-categories/:id/exists", unitCategoryRest.Exists)
		routeV1.POST("/units-categories/:id", unitCategoryRest.Create)
		routeV1.PUT("/units-categories/:id", unitCategoryRest.Update)

		return nil
	})
}
