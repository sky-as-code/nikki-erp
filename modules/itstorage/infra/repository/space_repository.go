package repository

import (
	"unicode/utf8"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itSpace "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/space"
)

// NewSpaceRepository is handed the composable default by the itstorage_space onion.
func NewSpaceRepository(base composable.CrudRepository) itSpace.SpaceRepository {
	return &SpaceRepositoryImpl{CrudRepository: base}
}

// SpaceRepositoryImpl adds the tree statements. They go around the row-at-a-time CRUD because a
// size roll-up or a rename touches a whole chain of rows, which must change together.
type SpaceRepositoryImpl struct {
	composable.CrudRepository
}

func (this *SpaceRepositoryImpl) AddSize(ctx corectx.Context, spaceId model.Id, delta int64) error {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "AddSize")
	if err != nil {
		return err
	}
	query, args := buildAddSizeSql(this.Schema().TableName(), spaceId, delta, tenantColumn, tenantId)
	_, err = execCount(ctx, this, query, args, "AddSize")
	return err
}

// buildAddSizeSql walks the parent chain with a recursive CTE and updates every space on it in
// the same statement, so a concurrent upload elsewhere in the tree cannot interleave half-way.
func buildAddSizeSql(
	table string, spaceId model.Id, delta int64, tenantColumn string, tenantId string,
) (string, []any) {
	var anchorTenant, updateTenant sqlbuilder.Builder = sqlbuilder.Buildf(""), sqlbuilder.Buildf("")
	if tenantColumn != "" {
		anchorTenant = sqlbuilder.Buildf(" AND %v = %v", sqlbuilder.Raw(tenantColumn), tenantId)
		updateTenant = sqlbuilder.Buildf(" AND %v = %v", sqlbuilder.Raw(tenantColumn), tenantId)
	}
	return sqlbuilder.Buildf(
		"WITH RECURSIVE chain(id, parent_id) AS ("+
			"SELECT id, parent_id FROM %v WHERE id = %v%v"+
			" UNION ALL SELECT s.id, s.parent_id FROM %v s JOIN chain ON s.id = chain.parent_id"+
			") UPDATE %v SET %v = GREATEST(%v + %v, 0) WHERE id IN (SELECT id FROM chain)%v",
		sqlbuilder.Raw(table), spaceId, anchorTenant,
		sqlbuilder.Raw(table),
		sqlbuilder.Raw(table), sqlbuilder.Raw(models.SpaceFieldSizeBytes), sqlbuilder.Raw(models.SpaceFieldSizeBytes),
		delta, updateTenant,
	).BuildWithFlavor(sqlbuilder.PostgreSQL)
}

func (this *SpaceRepositoryImpl) RewriteDescendantPaths(
	ctx corectx.Context, oldPath string, newPath string,
) (int, error) {
	if oldPath == "" || newPath == "" {
		return 0, errors.New("RewriteDescendantPaths: empty path")
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "RewriteDescendantPaths")
	if err != nil {
		return 0, err
	}
	query, args := buildRewritePathsSql(this.Schema().TableName(), oldPath, newPath, tenantColumn, tenantId)
	return execCount(ctx, this, query, args, "RewriteDescendantPaths")
}

// buildRewritePathsSql matches descendants by a literal prefix comparison rather than LIKE, so a
// '_' in a space name is not a wildcard.
func buildRewritePathsSql(
	table string, oldPath string, newPath string, tenantColumn string, tenantId string,
) (string, []any) {
	oldPrefix := oldPath + models.PathSeparator
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(table)
	builder.Set(builder.Assign(models.SpaceFieldPath, sqlbuilder.Buildf(
		"%v || substr(%v, %v)", newPath, sqlbuilder.Raw(models.SpaceFieldPath), utf8.RuneCountInString(oldPath)+1,
	)))
	builder.Where(builder.Var(sqlbuilder.Buildf(
		"substr(%v, 1, %v) = %v", sqlbuilder.Raw(models.SpaceFieldPath), utf8.RuneCountInString(oldPrefix), oldPrefix,
	)))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	return builder.Build()
}

func (this *SpaceRepositoryImpl) ListSubtreeIds(ctx corectx.Context, path string) ([]model.Id, error) {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "ListSubtreeIds")
	if err != nil {
		return nil, err
	}
	query, args := buildListSubtreeSql(this.Schema().TableName(), path, tenantColumn, tenantId)
	return queryStrings(ctx, this, query, args, "ListSubtreeIds")
}

func buildListSubtreeSql(table string, path string, tenantColumn string, tenantId string) (string, []any) {
	prefix := path + models.PathSeparator
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(models.SpaceFieldId).From(table)
	builder.Where(builder.Or(
		builder.Equal(models.SpaceFieldPath, path),
		builder.Var(sqlbuilder.Buildf(
			"substr(%v, 1, %v) = %v", sqlbuilder.Raw(models.SpaceFieldPath), utf8.RuneCountInString(prefix), prefix,
		)),
	))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	// Deepest first: a parent can only go once its children have.
	builder.OrderBy("length(" + models.SpaceFieldPath + ") DESC")
	return builder.Build()
}

func (this *SpaceRepositoryImpl) DeleteSpaces(ctx corectx.Context, spaceIds []model.Id) (int, error) {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "DeleteSpaces")
	if err != nil {
		return 0, err
	}
	// One statement per space, in the given order: the parent edge is RESTRICT, which is checked
	// row by row, so deleting a parent and its child in one statement can fail on the parent.
	total := 0
	for _, spaceId := range spaceIds {
		builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
		builder.DeleteFrom(this.Schema().TableName())
		builder.Where(builder.Equal(models.SpaceFieldId, spaceId))
		if tenantColumn != "" {
			builder.Where(builder.Equal(tenantColumn, tenantId))
		}
		query, args := builder.Build()
		count, err := execCount(ctx, this, query, args, "DeleteSpaces")
		if err != nil {
			return total, err
		}
		total += count
	}
	return total, nil
}
