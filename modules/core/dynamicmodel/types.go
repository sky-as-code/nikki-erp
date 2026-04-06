package dynamicmodel

import (
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
)

type DynamicModelRepository interface {
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)
	GetBaseRepo() BaseDynamicRepository
}

type BaseDynamicRepository interface {
	Schema() *dmodel.ModelSchema
	ExtractClient(ctx corectx.Context) orm.DbClient
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)
	CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[[][]string], error)
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*OpResult[int], error)
	Exists(ctx corectx.Context, keys []dmodel.DynamicFields) (*OpResult[RepoExistsResult], error)
	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[int], error)
	GetOne(ctx corectx.Context, param RepoGetOneParam) (*OpResult[dmodel.DynamicFields], error)
	ManageM2m(ctx corectx.Context, param RepoManageM2mParam) (*OpResult[int], error)
	ExistsM2m(ctx corectx.Context, param RepoExistsM2mParam) (bool, error)
	CountM2m(ctx corectx.Context, param RepoCountM2mParam) (*OpResult[int], error)
	Search(ctx corectx.Context, param RepoSearchParam) (*OpResult[PagedResultData[dmodel.DynamicFields]], error)
	Update(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[dmodel.DynamicFields], error)
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
	DestSchemaName string
	SrcId          model.Id
	// Field name for the source ID used to include in the error message.
	SrcIdFieldForError string
	// M2M edge name on the source schema.
	SrcEdgeName      string
	AssociatedIds    datastructure.Set[model.Id]
	DisassociatedIds datastructure.Set[model.Id]
}

// RepoExistsM2mParam checks the junction for an outgoing many-to-many edge on the repository schema.
// When dest_id is omitted, null, or empty, checks that SrcId has at least one junction row; otherwise checks the (SrcId, DestId) pair.
type RepoExistsM2mParam struct {
	M2mEdge string    `json:"m2m_edge"`
	SrcId   model.Id  `json:"src_id"`
	DestId  *model.Id `json:"dest_id"`
}

// RepoCountM2mParam counts junction rows for one source record on an outgoing many-to-many edge.
type RepoCountM2mParam struct {
	M2mEdge string   `json:"m2m_edge"`
	SrcId   model.Id `json:"src_id"`
}

type DynamicModelPtr[TDomain any] interface {
	*TDomain
	dmodel.DynamicModel
}
