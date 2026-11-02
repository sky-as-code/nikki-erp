package uomcat

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The CRUD commands, queries and results of the UoM Category resource, as composable shapes
// under a resource-specific name.
type (
	CreateUomCatCommand      = dynamicresource.CreateCommand
	UpdateUomCatCommand      = dynamicresource.UpdateCommand
	DeleteUomCatCommand      = dynamicresource.DeleteCommand
	SetUomCatArchivedCommand = dynamicresource.SetArchivedCommand
	GetUomCatByIdQuery       = dynamicresource.GetByIdQuery
	SearchUomCatsQuery       = dynamicresource.SearchQuery
	UomCatExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateUomCatResult      = dynamicresource.CreateResult
	UpdateUomCatResult      = dynamicresource.MutateResult
	DeleteUomCatResult      = dynamicresource.MutateResult
	SetUomCatArchivedResult = dynamicresource.MutateResult
	GetUomCatByIdResult     = dynamicresource.GetOneResult
	SearchUomCatsResult     = dynamicresource.SearchResult
	UomCatExistsResult      = dynamicresource.ExistsResult
)
