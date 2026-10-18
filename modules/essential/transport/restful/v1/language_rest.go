package v1

import (
	"github.com/labstack/echo/v5"
	"go.uber.org/dig"

	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// LanguageEngineName is the container name of the language onion, checked by a test.
const LanguageEngineName = "dynengine_essential_language"

type languageRestParams struct {
	dig.In

	Engine composable.DynamicResourceEngineOnion `name:"dynengine_essential_language"`
}

func NewLanguageRest(params languageRestParams) *LanguageRest {
	rest := &LanguageRest{languageSvc: params.Engine.ApplicationService().(it.LanguageApplicationService)}
	rest.SetApplicationService(params.Engine.ApplicationService())
	return rest
}

// LanguageRest serves the built-in CRUD of the language catalogue plus its two public reads.
type LanguageRest struct {
	composable.CrudRestBase
	languageSvc it.LanguageApplicationService
}

// GetLanguageJson serves one module's translation bundle (?language_code=&module_name=).
func (this *LanguageRest) GetLanguageJson(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "get language json", payload,
		this.languageSvc.GetLanguageJsonAction, composable.Identity)
}

// ListEnabledLanguages serves the languages a user may pick.
func (this *LanguageRest) ListEnabledLanguages(echoCtx *echo.Context, payload map[string]any) error {
	return composable.ServeAction(echoCtx, "list enabled languages", payload,
		this.languageSvc.ListEnabledLanguagesAction, func(data []models.Language) any {
			return NewListEnabledLanguagesResponse(data)
		})
}
