package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewLanguageRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewLanguageRepository(base composable.CrudRepository) itLang.LanguageRepository {
	return &LanguageRepositoryImpl{CrudRepository: base}
}

type LanguageRepositoryImpl struct {
	composable.CrudRepository
}
