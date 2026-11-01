package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	itLang "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
)

// NewLanguageRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewLanguageRepository(base dynamicresource.CrudRepository) itLang.LanguageRepository {
	return &LanguageRepositoryImpl{CrudRepository: base}
}

type LanguageRepositoryImpl struct {
	dynamicresource.CrudRepository
}
