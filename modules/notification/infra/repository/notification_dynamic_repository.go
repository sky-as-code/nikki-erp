package repository

import (
	"database/sql"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

func NewNotificationRepository(base composable.CrudRepository) it.NotificationRepository {
	return &NotificationRepositoryImpl{CrudRepository: base}
}

type NotificationRepositoryImpl struct {
	composable.CrudRepository
}

// FindByIdempotencyKey answers "has this send already been handled?".
//
// It is half of the replay guard: the lookup catches the common case, and the partial unique index
// catches the race the lookup cannot: two deliveries of one command arriving together both see no
// row, and the second insert fails on the constraint instead of creating a twin (BR 11.15, BR 12).
// Neither half is sufficient alone.
func (this *NotificationRepositoryImpl) FindByIdempotencyKey(
	ctx corectx.Context, orgId model.Id, sourceModule string, key string,
) (*model.Id, error) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("n.id")
	builder.From(notificationsTable + " AS n")
	builder.Where(
		builder.Equal("n.org_id", string(orgId)),
		builder.Equal("n.source_module", sourceModule),
		builder.Equal("n.idempotency_key", key),
	)
	scopeSelectToTenant(ctx, builder, "n")
	builder.Limit(1)

	query, args := builder.Build()

	var id string
	row := this.GetBaseRepo().ExtractClient(ctx).QueryRow(ctx, query, args...)
	switch err := row.Scan(&id); {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, errors.Wrap(err, "FindByIdempotencyKey")
	}

	found := model.Id(id)
	return &found, nil
}
