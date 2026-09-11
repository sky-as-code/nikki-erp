package uomcat

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD commands, queries and results of the UoM Category resource, as composable shapes
// under a resource-specific name.
type (
	CreateUomCatCommand      = composable.CreateCommand
	UpdateUomCatCommand      = composable.UpdateCommand
	DeleteUomCatCommand      = composable.DeleteCommand
	SetUomCatArchivedCommand = composable.SetArchivedCommand
	GetUomCatByIdQuery       = composable.GetByIdQuery
	SearchUomCatsQuery       = composable.SearchQuery
	UomCatExistsQuery        = composable.ExistsQuery
)

type (
	CreateUomCatResult      = composable.CreateResult
	UpdateUomCatResult      = composable.MutateResult
	DeleteUomCatResult      = composable.MutateResult
	SetUomCatArchivedResult = composable.MutateResult
	GetUomCatByIdResult     = composable.GetOneResult
	SearchUomCatsResult     = composable.SearchResult
	UomCatExistsResult      = composable.ExistsResult
)
