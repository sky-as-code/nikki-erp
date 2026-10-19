package language

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The CRUD commands, queries and results of the language resource: the composable shapes under a
// resource-specific name.
type (
	CreateLanguageCommand      = composable.CreateCommand
	UpdateLanguageCommand      = composable.UpdateCommand
	DeleteLanguageCommand      = composable.DeleteCommand
	SetLanguageArchivedCommand = composable.SetArchivedCommand
	GetLanguageByIdQuery       = composable.GetByIdQuery
	SearchLanguagesQuery       = composable.SearchQuery
	LanguageExistsQuery        = composable.ExistsQuery
)

type (
	CreateLanguageResult  = composable.CreateResult
	UpdateLanguageResult  = composable.MutateResult
	DeleteLanguageResult  = composable.MutateResult
	GetLanguageByIdResult = composable.GetOneResult
	SearchLanguagesResult = composable.SearchResult
	LanguageExistsResult  = composable.ExistsResult
)

type GetLanguageJsonResult = dyn.OpResult[model.LanguageTranslationItem]

// GetLanguageByIsoCodeQuery resolves a language by the BCP47 code every consumer names it with.
// The only caller is another module reading the acting user's formatting rules.
type GetLanguageByIsoCodeQuery struct {
	IsoCode string
}

// GetLanguageByIsoCodeResult carries the whole record rather than a narrowed projection: a caller
// formatting numbers and dates wants every field, and this table is global reference data with
// nothing org- or tenant-scoped to withhold.
type GetLanguageByIsoCodeResult = dyn.OpResult[models.Language]

// ListEnabledLanguagesQuery lists the languages a user may pick, ordered by name.
type ListEnabledLanguagesQuery struct{}

type ListEnabledLanguagesResult = dyn.OpResult[[]models.Language]

// GetLanguageJsonQuery asks for one module's translation bundle in one language.
type GetLanguageJsonQuery struct {
	LanguageCode string `json:"language_code" query:"language_code"`
	ModuleName   string `json:"module_name" query:"module_name"`
}
