package repository

import (
	"github.com/huandu/go-sqlbuilder"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
)

// NewFileRepository is handed the composable default by the itstorage_file onion.
func NewFileRepository(base composable.CrudRepository) itFile.FileRepository {
	return &FileRepositoryImpl{CrudRepository: base}
}

// FileRepositoryImpl adds the statements that remove a whole space's files at once, for the org
// cleanup, where one statement per file would hold the transaction open across every round trip.
type FileRepositoryImpl struct {
	composable.CrudRepository
}

func (this *FileRepositoryImpl) ListObjectKeysInSpaces(
	ctx corectx.Context, spaceIds []model.Id,
) ([]string, error) {
	if len(spaceIds) == 0 {
		return []string{}, nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "ListObjectKeysInSpaces")
	if err != nil {
		return nil, err
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(models.FileFieldObjectKey).From(this.Schema().TableName())
	builder.Where(builder.In(models.FileFieldSpaceId, toArgs(spaceIds)...))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	query, args := builder.Build()
	return queryStrings(ctx, this, query, args, "ListObjectKeysInSpaces")
}

func (this *FileRepositoryImpl) DeleteInSpaces(ctx corectx.Context, spaceIds []model.Id) (int, error) {
	if len(spaceIds) == 0 {
		return 0, nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "DeleteInSpaces")
	if err != nil {
		return 0, err
	}
	builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	builder.DeleteFrom(this.Schema().TableName())
	builder.Where(builder.In(models.FileFieldSpaceId, toArgs(spaceIds)...))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	query, args := builder.Build()
	return execCount(ctx, this, query, args, "DeleteInSpaces")
}

func toArgs(ids []model.Id) []any {
	return array.Map(ids, func(id model.Id) any { return id })
}
