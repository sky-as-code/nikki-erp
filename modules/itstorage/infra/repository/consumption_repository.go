package repository

import (
	"time"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itConsumption "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/consumption"
)

// NewConsumptionRepository is handed the composable default by the itstorage_consumption onion.
func NewConsumptionRepository(base composable.CrudRepository) itConsumption.ConsumptionRepository {
	return &ConsumptionRepositoryImpl{CrudRepository: base}
}

// ConsumptionRepositoryImpl keeps the counters with single conditional statements, so the database
// serializes concurrent uploads on the owner's row without a transaction held across the upload.
type ConsumptionRepositoryImpl struct {
	composable.CrudRepository
}

func (this *ConsumptionRepositoryImpl) Ensure(ctx corectx.Context, owners []models.ConsumptionOwner) error {
	if len(owners) == 0 {
		return nil
	}
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "Ensure")
	if err != nil {
		return err
	}
	query, args, err := buildEnsureSql(this.Schema().TableName(), owners, tenantColumn, tenantId)
	if err != nil {
		return err
	}
	_, err = execCount(ctx, this, query, args, "Ensure")
	return err
}

// buildEnsureSql relies on the (owner_type, owner_id) unique key: ON CONFLICT DO NOTHING leaves an
// existing row, and a concurrent first upload, untouched.
func buildEnsureSql(
	table string, owners []models.ConsumptionOwner, tenantColumn string, tenantId string,
) (string, []any, error) {
	columns := []string{
		models.ConsumptionFieldId, models.ConsumptionFieldOwnerType, models.ConsumptionFieldOwnerId,
		models.ConsumptionFieldUsedBytes, models.ConsumptionFieldReservedBytes,
		models.ConsumptionFieldCreatedAt, models.ConsumptionFieldEtag,
	}
	if tenantColumn != "" {
		columns = append(columns, tenantColumn)
	}
	builder := sqlbuilder.PostgreSQL.NewInsertBuilder()
	builder.InsertInto(table).Cols(columns...)
	now := time.Now().UTC()
	for _, owner := range owners {
		rowId, err := model.NewId()
		if err != nil {
			return "", nil, errors.Wrap(err, "Ensure: new consumption id")
		}
		values := []any{*rowId, owner.Type, owner.Id, max(owner.InitialUsedBytes, 0), int64(0), now, *model.NewEtag()}
		if tenantColumn != "" {
			values = append(values, tenantId)
		}
		builder.Values(values...)
	}
	builder.SQL("ON CONFLICT DO NOTHING")
	query, args := builder.Build()
	return query, args, nil
}

func (this *ConsumptionRepositoryImpl) Reserve(
	ctx corectx.Context, owner models.ConsumptionOwner, bytes int64, capBytes *int64,
) (bool, error) {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "Reserve")
	if err != nil {
		return false, err
	}
	query, args := buildReserveSql(this.Schema().TableName(), owner, bytes, capBytes, tenantColumn, tenantId)
	count, err := execCount(ctx, this, query, args, "Reserve")
	return count > 0, err
}

// buildReserveSql checks and reserves in one statement: two uploads racing for the last bytes
// queue on the row lock, and the second re-evaluates the cap against the first one's reservation.
func buildReserveSql(
	table string, owner models.ConsumptionOwner, bytes int64, capBytes *int64, tenantColumn string, tenantId string,
) (string, []any) {
	builder := newCounterUpdate(table, owner, tenantColumn, tenantId)
	builder.SetMore(builder.Assign(models.ConsumptionFieldReservedBytes, sqlbuilder.Buildf(
		"%v + %v", sqlbuilder.Raw(models.ConsumptionFieldReservedBytes), bytes)))
	if capBytes != nil {
		builder.Where(builder.Var(sqlbuilder.Buildf("%v + %v + %v <= %v",
			sqlbuilder.Raw(models.ConsumptionFieldUsedBytes), sqlbuilder.Raw(models.ConsumptionFieldReservedBytes),
			bytes, *capBytes)))
	}
	return builder.Build()
}

func (this *ConsumptionRepositoryImpl) Commit(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error {
	return this.eachOwner(ctx, owners, "Commit", func(builder *sqlbuilder.UpdateBuilder) {
		builder.SetMore(
			builder.Assign(models.ConsumptionFieldUsedBytes, addClamped(models.ConsumptionFieldUsedBytes, bytes)),
			builder.Assign(models.ConsumptionFieldReservedBytes, addClamped(models.ConsumptionFieldReservedBytes, -bytes)),
		)
	})
}

func (this *ConsumptionRepositoryImpl) Release(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error {
	return this.eachOwner(ctx, owners, "Release", func(builder *sqlbuilder.UpdateBuilder) {
		builder.SetMore(builder.Assign(models.ConsumptionFieldReservedBytes,
			addClamped(models.ConsumptionFieldReservedBytes, -bytes)))
	})
}

func (this *ConsumptionRepositoryImpl) Subtract(ctx corectx.Context, owners []models.ConsumptionOwner, bytes int64) error {
	return this.eachOwner(ctx, owners, "Subtract", func(builder *sqlbuilder.UpdateBuilder) {
		builder.SetMore(builder.Assign(models.ConsumptionFieldUsedBytes, addClamped(models.ConsumptionFieldUsedBytes, -bytes)))
	})
}

func (this *ConsumptionRepositoryImpl) SubtractEach(
	ctx corectx.Context, ownerType string, bytesByOwner map[model.Id]int64,
) error {
	for ownerId, bytes := range bytesByOwner {
		owner := models.ConsumptionOwner{Type: ownerType, Id: ownerId}
		if err := this.Subtract(ctx, []models.ConsumptionOwner{owner}, bytes); err != nil {
			return err
		}
	}
	return nil
}

func (this *ConsumptionRepositoryImpl) DeleteOwner(ctx corectx.Context, owner models.ConsumptionOwner) error {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), "DeleteOwner")
	if err != nil {
		return err
	}
	builder := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	builder.DeleteFrom(this.Schema().TableName())
	builder.Where(
		builder.Equal(models.ConsumptionFieldOwnerType, owner.Type),
		builder.Equal(models.ConsumptionFieldOwnerId, owner.Id),
	)
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	query, args := builder.Build()
	_, err = execCount(ctx, this, query, args, "DeleteOwner")
	return err
}

func (this *ConsumptionRepositoryImpl) eachOwner(
	ctx corectx.Context, owners []models.ConsumptionOwner, action string, set func(*sqlbuilder.UpdateBuilder),
) error {
	tenantColumn, tenantId, err := tenantScope(ctx, this.Schema(), action)
	if err != nil {
		return err
	}
	for _, owner := range owners {
		builder := newCounterUpdate(this.Schema().TableName(), owner, tenantColumn, tenantId)
		set(builder)
		query, args := builder.Build()
		if _, err := execCount(ctx, this, query, args, action); err != nil {
			return err
		}
	}
	return nil
}

// newCounterUpdate targets one owner's row and bumps its etag and updated_at; the caller adds the
// counter assignments with SetMore.
func newCounterUpdate(
	table string, owner models.ConsumptionOwner, tenantColumn string, tenantId string,
) *sqlbuilder.UpdateBuilder {
	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(table)
	builder.Set(
		builder.Assign(models.ConsumptionFieldEtag, *model.NewEtag()),
		builder.Assign("updated_at", time.Now().UTC()),
	)
	builder.Where(
		builder.Equal(models.ConsumptionFieldOwnerType, owner.Type),
		builder.Equal(models.ConsumptionFieldOwnerId, owner.Id),
	)
	if tenantColumn != "" {
		builder.Where(builder.Equal(tenantColumn, tenantId))
	}
	return builder
}

// addClamped never takes a counter below zero: a counter that would go negative has already
// drifted, which the reconciler repairs (docs/problems itstorage-001), and the column refuses it.
func addClamped(column string, delta int64) sqlbuilder.Builder {
	return sqlbuilder.Buildf("GREATEST(%v + %v, 0)", sqlbuilder.Raw(column), delta)
}
