package product

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// ProductTemplateCombinationExclusionRepository reads and writes the pairwise exclusion rows.
type ProductTemplateCombinationExclusionRepository interface {
	dynamicresource.CrudRepository
}

// ProductTemplateCombinationExclusionDomainService is the CRUD of the resource plus its own rules.
type ProductTemplateCombinationExclusionDomainService interface {
	dynamicresource.CrudDomainService
}

// ProductTemplateCombinationExclusionApplicationService is the authorized surface the REST reads serve.
type ProductTemplateCombinationExclusionApplicationService interface {
	dynamicresource.CrudApplicationService
}

// The CRUD commands, queries and results of the resource, as composable shapes under a
// resource-specific name.
type (
	CreateProductTemplateCombinationExclusionCommand      = dynamicresource.CreateCommand
	UpdateProductTemplateCombinationExclusionCommand      = dynamicresource.UpdateCommand
	DeleteProductTemplateCombinationExclusionCommand      = dynamicresource.DeleteCommand
	SetProductTemplateCombinationExclusionArchivedCommand = dynamicresource.SetArchivedCommand
	GetProductTemplateCombinationExclusionByIdQuery       = dynamicresource.GetByIdQuery
	SearchProductTemplateCombinationExclusionsQuery       = dynamicresource.SearchQuery
	ProductTemplateCombinationExclusionExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateProductTemplateCombinationExclusionResult      = dynamicresource.CreateResult
	UpdateProductTemplateCombinationExclusionResult      = dynamicresource.MutateResult
	DeleteProductTemplateCombinationExclusionResult      = dynamicresource.MutateResult
	SetProductTemplateCombinationExclusionArchivedResult = dynamicresource.MutateResult
	GetProductTemplateCombinationExclusionByIdResult     = dynamicresource.GetOneResult
	SearchProductTemplateCombinationExclusionsResult     = dynamicresource.SearchResult
	ProductTemplateCombinationExclusionExistsResult      = dynamicresource.ExistsResult
)
