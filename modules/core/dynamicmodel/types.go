package dynamicmodel

import (
	"database/sql"

	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type DynamicModelRepository interface {
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)
	GetBaseRepo() BaseDynamicRepository
}

type NewBaseRepoParam struct {
	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	Logger       logging.LoggerService
	QueryBuilder orm.QueryBuilder
	Schema       *dmodel.ModelSchema
}

type NewBaseDynamicRepositoryFn func(param NewBaseRepoParam) BaseDynamicRepository

type BaseDynamicRepository interface {
	Schema() *dmodel.ModelSchema
	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)
	ExtractClient(ctx corectx.Context) orm.DbClient
	ExecFunc(ctx corectx.Context, sqlFuncName string, sqlFuncArgs ...any) error
	QueryFunc(ctx corectx.Context, sqlFuncName string, sqlFuncArgs ...any) (*sql.Rows, error)

	CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[[][]string], error)
	CountM2m(ctx corectx.Context, param RepoCountM2mParam) (*OpResult[int], error)
	DeleteOne(ctx corectx.Context, keys dmodel.DynamicFields) (*OpResult[int], error)
	Exists(ctx corectx.Context, keys []dmodel.DynamicFields) (*OpResult[RepoExistsResult], error)
	ExistsM2m(ctx corectx.Context, param RepoExistsM2mParam) (bool, error)
	InsertBulk(ctx corectx.Context, data []dmodel.DynamicFields) (*OpResult[int], error)
	GetOne(ctx corectx.Context, param RepoGetOneParam) (*OpResult[dmodel.DynamicFields], error)
	Insert(ctx corectx.Context, data dmodel.DynamicFields) (*OpResult[int], error)
	// ManageM2m inserts and/or deletes junction rows for a finalized many-to-many link to dest schema.
	// Source and destination are identified by id.
	ManageM2m(ctx corectx.Context, param RepoManageM2mParam) (*OpResult[int], error)
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
	Filter dmodel.DynamicFields
	Fields []string
}

type RepoSearchParam struct {
	Fields   []string
	Filter   []dmodel.DynamicFields
	Page     int
	Size     int
	Graph    *dmodel.SearchGraph
	Language *model.LanguageCode

	// ComputedContext carries the request's whitelisted context values for SQL-computed
	// fields; the query builder binds them into "${ctx.key}" filter references.
	ComputedContext map[string]any

	// IncludeArchived controls whether archived records take part in the search.
	// It is tri-state on purpose:
	//   - nil: no filtering at all. This is the legacy contract that every internal lookup
	//     relies on, e.g. models.FindTemplateVariants which needs archived rows by design.
	//     Leave it nil when reusing Search to fetch a single known record.
	//   - false: the repository prepends "is_archived = false" to the search graph.
	//     crud.Search sets this when the caller omitted the query parameter, so the public
	//     HTTP API hides archived records by default.
	//   - true: no filtering; archived records are returned alongside active ones.
	//
	// The condition is only prepended on schemas that actually carry the is_archived column
	// (it comes from the optional basemodel.ArchivableModelSchemaBuilder mixin).
	//
	// Note this filters the searched schema only. Nested edge columns are hydrated by
	// follow-up queries that are not filtered, so requesting fields=groups.name with
	// IncludeArchived=false excludes archived rows of this schema but still hydrates
	// archived rows of the edge.
	IncludeArchived *bool
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
	BeforeInsert     RepoBeforeInsertM2mFn
}

type RepoBeforeInsertM2mFn func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error

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

type DynamicModelPtr[T any] interface {
	*T
	dmodel.DynamicModel
}

type DynamicModelSetterPtr[T any] interface {
	*T
	dmodel.DynamicModelSetter
}
