package repository

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itTag "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/tag"
)

// NewTagRepository is handed the composable default by the onion. The type exists so a bespoke
// query has a home; it adds nothing today.
func NewTagRepository(base composable.CrudRepository) itTag.TagRepository {
	return &TagRepositoryImpl{CrudRepository: base}
}

type TagRepositoryImpl struct {
	composable.CrudRepository
}
