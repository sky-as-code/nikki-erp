package uom

import (
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// The CRUD commands, queries and results of the UoM resource, as composable shapes under a
// resource-specific name.
type (
	CreateUomCommand      = dynamicresource.CreateCommand
	UpdateUomCommand      = dynamicresource.UpdateCommand
	DeleteUomCommand      = dynamicresource.DeleteCommand
	SetUomArchivedCommand = dynamicresource.SetArchivedCommand
	GetUomByIdQuery       = dynamicresource.GetByIdQuery
	SearchUomsQuery       = dynamicresource.SearchQuery
	UomExistsQuery        = dynamicresource.ExistsQuery
)

type (
	CreateUomResult      = dynamicresource.CreateResult
	UpdateUomResult      = dynamicresource.MutateResult
	DeleteUomResult      = dynamicresource.MutateResult
	SetUomArchivedResult = dynamicresource.MutateResult
	GetUomByIdResult     = dynamicresource.GetOneResult
	SearchUomsResult     = dynamicresource.SearchResult
	UomExistsResult      = dynamicresource.ExistsResult
)
