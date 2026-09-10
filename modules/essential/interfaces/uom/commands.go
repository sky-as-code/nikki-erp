package uom

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// The CRUD commands, queries and results of the UoM resource, as composable shapes under a
// resource-specific name.
type (
	CreateUomCommand      = composable.CreateCommand
	UpdateUomCommand      = composable.UpdateCommand
	DeleteUomCommand      = composable.DeleteCommand
	SetUomArchivedCommand = composable.SetArchivedCommand
	GetUomByIdQuery       = composable.GetByIdQuery
	SearchUomsQuery       = composable.SearchQuery
	UomExistsQuery        = composable.ExistsQuery
)

type (
	CreateUomResult      = composable.CreateResult
	UpdateUomResult      = composable.MutateResult
	DeleteUomResult      = composable.MutateResult
	SetUomArchivedResult = composable.MutateResult
	GetUomByIdResult     = composable.GetOneResult
	SearchUomsResult     = composable.SearchResult
	UomExistsResult      = composable.ExistsResult
)
