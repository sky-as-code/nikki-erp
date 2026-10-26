package repository

import (
	"time"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
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

func (this *FileRepositoryImpl) SumSizesByUploader(
	ctx corectx.Context, spaceIds []model.Id,
) (map[model.Id]int64, error) {
	sums := map[model.Id]int64{}
	if len(spaceIds) == 0 {
		return sums, nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "SumSizesByUploader")
	if err != nil {
		return nil, err
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(models.FileFieldUploadedBy, "SUM("+models.FileFieldSizeBytes+")").From(this.Schema().TableName())
	builder.Where(builder.In(models.FileFieldSpaceId, toArgs(spaceIds)...), builder.IsNotNull(models.FileFieldUploadedBy))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	builder.GroupBy(models.FileFieldUploadedBy)
	query, args := builder.Build()
	rows, err := this.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "SumSizesByUploader")
	}
	defer rows.Close()
	for rows.Next() {
		var uploader string
		var total int64
		if err := rows.Scan(&uploader, &total); err != nil {
			return nil, errors.Wrap(err, "SumSizesByUploader")
		}
		sums[model.Id(uploader)] = total
	}
	return sums, errors.Wrap(rows.Err(), "SumSizesByUploader")
}

func (this *FileRepositoryImpl) SetStatus(
	ctx corectx.Context, ids []model.Id, fromStatus string, toStatus string,
) (map[model.Id]string, error) {
	if len(ids) == 0 {
		return map[model.Id]string{}, nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "SetStatus")
	if err != nil {
		return nil, err
	}
	query, args := buildSetStatusSql(this.Schema().TableName(), ids, fromStatus, toStatus, tenantColumn, tenantId)
	return queryIdKeys(ctx, this, query, args, "SetStatus")
}

// buildSetStatusSql returns the rows it moved: a concurrent writer that moved a file first leaves
// it out, which is how the caller learns the file was not where it expected.
func buildSetStatusSql(
	table string, ids []model.Id, fromStatus string, toStatus string, tenantColumn string, tenantId string,
) (string, []any) {
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(table)
	now := time.Now().UTC()
	builder.Set(
		builder.Assign(models.FileFieldStatus, toStatus),
		builder.Assign(models.FileFieldEtag, *model.NewEtag()),
		builder.Assign("updated_at", now),
	)
	builder.Where(builder.In(models.FileFieldId, toArgs(ids)...), builder.Equal(models.FileFieldStatus, fromStatus))
	if toStatus == storage.FileStatusInUse {
		// A released file is pending with in_use_at set; it must never come back into use.
		builder.SetMore(builder.Assign(models.FileFieldInUseAt, now))
		builder.Where(builder.IsNull(models.FileFieldInUseAt))
	}
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	builder.Returning(models.FileFieldId, models.FileFieldObjectKey)
	return builder.Build()
}

func (this *FileRepositoryImpl) ListObjectKeysWithStatus(
	ctx corectx.Context, ids []model.Id, status string,
) (map[model.Id]string, error) {
	if len(ids) == 0 {
		return map[model.Id]string{}, nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "ListObjectKeysWithStatus")
	if err != nil {
		return nil, err
	}
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(models.FileFieldId, models.FileFieldObjectKey).From(this.Schema().TableName())
	builder.Where(builder.In(models.FileFieldId, toArgs(ids)...), builder.Equal(models.FileFieldStatus, status))
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	query, args := builder.Build()
	return queryIdKeys(ctx, this, query, args, "ListObjectKeysWithStatus")
}

func queryIdKeys(
	ctx corectx.Context, repo composable.CrudRepository, query string, args []any, action string,
) (map[model.Id]string, error) {
	rows, err := repo.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, action)
	}
	defer rows.Close()
	keys := map[model.Id]string{}
	for rows.Next() {
		var id, objectKey string
		if err := rows.Scan(&id, &objectKey); err != nil {
			return nil, errors.Wrap(err, action)
		}
		keys[model.Id(id)] = objectKey
	}
	return keys, errors.Wrap(rows.Err(), action)
}
