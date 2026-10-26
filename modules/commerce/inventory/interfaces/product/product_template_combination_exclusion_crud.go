package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// ProductTemplateCombinationExclusionRepository reads and writes the pairwise exclusion rows.
type ProductTemplateCombinationExclusionRepository interface {
	composable.CrudRepository
}

// ProductTemplateCombinationExclusionDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateCombinationExclusionDomainService interface {
	composable.CrudDomainService
}

// ProductTemplateCombinationExclusionApplicationService is the authorized surface the REST reads serve.
type ProductTemplateCombinationExclusionApplicationService interface {
	composable.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateCombinationExclusionCommand      = composable.CreateCommand
	UpdateProductTemplateCombinationExclusionCommand      = composable.UpdateCommand
	DeleteProductTemplateCombinationExclusionCommand      = composable.DeleteCommand
	SetProductTemplateCombinationExclusionArchivedCommand = composable.SetArchivedCommand
	GetProductTemplateCombinationExclusionByIdQuery       = composable.GetByIdQuery
	SearchProductTemplateCombinationExclusionsQuery       = composable.SearchQuery
	ProductTemplateCombinationExclusionExistsQuery        = composable.ExistsQuery
)

type (
	CreateProductTemplateCombinationExclusionResult      = composable.CreateResult
	UpdateProductTemplateCombinationExclusionResult      = composable.MutateResult
	DeleteProductTemplateCombinationExclusionResult      = composable.MutateResult
	SetProductTemplateCombinationExclusionArchivedResult = composable.MutateResult
	GetProductTemplateCombinationExclusionByIdResult     = composable.GetOneResult
	SearchProductTemplateCombinationExclusionsResult     = composable.SearchResult
	ProductTemplateCombinationExclusionExistsResult      = composable.ExistsResult
)
