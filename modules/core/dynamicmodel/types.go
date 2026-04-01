package dynamicmodel

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

type BaseRepository interface {
	// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
	CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[[][]string], error)
	// DeleteOne deletes a single record by primary key then returns the number of affected rows.
	// If affected rows is 0, the record is not found.
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*OpResult[int], error)
	// ManageM2m inserts and/or deletes junction rows for a finalized many-to-many link to dest schema.
	// Source and destination are identified by id.
	ManageM2m(ctx corectx.Context, param RepoManageM2mParam) (*OpResult[int], error)
	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[int], error)
	GetOne(ctx corectx.Context, param RepoGetOneParam) (*OpResult[dmodel.DynamicFields], error)
	Search(ctx corectx.Context, param RepoSearchParam) (*OpResult[PagedResultData[dmodel.DynamicFields]], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[dmodel.DynamicFields], error)
	Exists(ctx corectx.Context, keys []dmodel.DynamicFields) (*OpResult[RepoExistsResult], error)
	GetSchema() *dmodel.ModelSchema
}

// RepoM2mAssociation is one row to insert into the M2M junction: source entity keys and peer entity keys.
type RepoM2mAssociation struct {
	SrcKeys  dmodel.DynamicFields
	DestKeys dmodel.DynamicFields
}

// RepoExistsResult is the raw batch existence outcome per filter map (same order as input keys).
type RepoExistsResult struct {
	Existing    []dmodel.DynamicFields `json:"existing"`
	NotExisting []dmodel.DynamicFields `json:"not_existing"`
}

type RepoGetOneParam struct {
	Filter  dmodel.DynamicFields
	Columns []string
}

type RepoSearchParam struct {
	Graph   *dmodel.SearchGraph
	Columns []string
	Filter  []dmodel.DynamicFields
	Page    int
	Size    int
}

type RepoManageM2mParam struct {
	DestSchemaName     string
	SrcId              model.Id
	SrcIdFieldForError string
	AssociatedIds      datastructure.Set[model.Id]
	DisassociatedIds   datastructure.Set[model.Id]
}

type BaseRepoGetter interface {
	GetBaseRepo() BaseRepository
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	dmodel.DynamicModel
}
