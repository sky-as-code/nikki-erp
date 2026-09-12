package repository

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// The physical tables these statements address. They go around the resource engine deliberately:
// every read here joins the recipient to its notification, and the engine reads one schema at a
// time, so going through it would mean a query per row to fill in the title and message.
const (
	recipientsTable    = "notification_recipients"
	notificationsTable = "notification_notifications"
)

func NewRecipientRepository(base composable.CrudRepository) it.RecipientRepository {
	return &RecipientRepositoryImpl{CrudRepository: base}
}

type RecipientRepositoryImpl struct {
	composable.CrudRepository
}

// NextStreamSeq draws the next ordering key.
//
// A sequence rather than max()+1 or a timestamp: it is monotonic without taking a lock, and the
// gaps a rolled-back transaction leaves are explicitly allowed (BR 27, BR-FS 7).
func (this *RecipientRepositoryImpl) NextStreamSeq(ctx corectx.Context) (int64, error) {
	client := this.GetBaseRepo().ExtractClient(ctx)

	var seq int64
	row := client.QueryRow(ctx, nextStreamSeqSql)
	if err := row.Scan(&seq); err != nil {
		return 0, errors.Wrap(err, "NextStreamSeq")
	}
	return seq, nil
}

// inboxColumns is the projection every read here shares, so that one scanner serves them all.
var inboxColumns = []string{
	"r.id", "r.notification_id", "r.stream_seq", "r.read_at", "r.created_at",
	"n.title", "n.message", "n.severity", "n.source_module",
	"n.source_resource_name", "n.source_resource_key", "n.metadata",
}

// selectInbox builds the shared skeleton: this user's rows in this organization, joined to their
// notification. The org and user predicates are not optional and not caller-supplied -- they come
// from the request context, which is what stops one person reading another's inbox (BR 18, BR 29).
func (this *RecipientRepositoryImpl) selectInbox(
	ctx corectx.Context, orgId model.Id, userId model.Id,
) *sqlbuilder.SelectBuilder {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select(inboxColumns...)
	builder.From(recipientsTable + " AS r")
	builder.JoinWithOption(sqlbuilder.InnerJoin, notificationsTable+" AS n", "n.id = r.notification_id")
	builder.Where(
		builder.Equal("r.org_id", string(orgId)),
		builder.Equal("r.recipient_user_id", string(userId)),
	)
	scopeSelectToTenant(ctx, builder, "r")
	return builder
}

func (this *RecipientRepositoryImpl) Inbox(
	ctx corectx.Context, orgId model.Id, userId model.Id, query it.InboxQuery,
) ([]it.InboxItem, error) {
	builder := this.selectInbox(ctx, orgId, userId)

	if query.IsRead != nil {
		if *query.IsRead {
			builder.Where(builder.IsNotNull("r.read_at"))
		} else {
			builder.Where(builder.IsNull("r.read_at"))
		}
	}
	if query.Severity != nil {
		builder.Where(builder.Equal("n.severity", *query.Severity))
	}
	if query.SourceModule != nil {
		builder.Where(builder.Equal("n.source_module", *query.SourceModule))
	}
	// The cursor is a position in the stream, not an offset: paging by offset shifts under a
	// reader whenever a notification arrives at the top, which is exactly when someone is reading.
	if query.Cursor != nil {
		builder.Where(builder.LessThan("r.stream_seq", *query.Cursor))
	}

	builder.OrderBy("r.stream_seq").Desc()
	builder.Limit(query.Limit)

	return this.queryItems(ctx, builder, "Inbox")
}

// Replay returns what a reconnecting client missed, oldest first, so it can be written to the
// stream in the order it would have arrived (BR-FS 10, BR-FS 11).
func (this *RecipientRepositoryImpl) Replay(
	ctx corectx.Context, orgId model.Id, userId model.Id, afterSeq *int64, limit int,
) ([]it.InboxItem, error) {
	builder := this.selectInbox(ctx, orgId, userId)
	if afterSeq != nil {
		builder.Where(builder.GreaterThan("r.stream_seq", *afterSeq))
	}
	builder.OrderBy("r.stream_seq").Asc()
	builder.Limit(limit)

	return this.queryItems(ctx, builder, "Replay")
}

func (this *RecipientRepositoryImpl) UnreadCount(
	ctx corectx.Context, orgId model.Id, userId model.Id,
) (int, error) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("COUNT(1)")
	builder.From(recipientsTable + " AS r")
	builder.Where(
		builder.Equal("r.org_id", string(orgId)),
		builder.Equal("r.recipient_user_id", string(userId)),
		builder.IsNull("r.read_at"),
	)
	scopeSelectToTenant(ctx, builder, "r")

	query, args := builder.Build()
	var count int
	row := this.GetBaseRepo().ExtractClient(ctx).QueryRow(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "UnreadCount")
	}
	return count, nil
}

// MarkRead stamps the unread rows of these notifications for this user.
//
// One statement rather than a read-then-write loop, because that is what makes it idempotent under
// concurrency: the read_at IS NULL predicate means a second call updates nothing and the first
// timestamp survives (BR 20.5, BR 20.6). Rows belonging to another user or organization are not
// matched at all, so a valid id someone else owns is silently skipped rather than refused (BR 21).
func (this *RecipientRepositoryImpl) MarkRead(
	ctx corectx.Context, orgId model.Id, userId model.Id, notificationIds []model.Id,
) (it.MarkReadResultData, error) {
	result := it.MarkReadResultData{RequestedCount: len(notificationIds)}
	if len(notificationIds) == 0 {
		return result, nil
	}

	ids := make([]any, 0, len(notificationIds))
	for _, id := range notificationIds {
		ids = append(ids, string(id))
	}

	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(recipientsTable)
	builder.Set(builder.Assign(models.RecipientFieldReadAt, sqlbuilder.Raw(nowUtcSqlExpr)))
	builder.Where(
		builder.Equal(models.RecipientFieldOrgId, string(orgId)),
		builder.Equal(models.RecipientFieldRecipientUserId, string(userId)),
		builder.In(models.RecipientFieldNotificationId, ids...),
		builder.IsNull(models.RecipientFieldReadAt),
	)
	scopeUpdateToTenant(ctx, builder)

	query, args := builder.Build()
	execResult, err := this.GetBaseRepo().ExtractClient(ctx).Exec(ctx, query, args...)
	if err != nil {
		return result, errors.Wrap(err, "MarkRead")
	}

	updated, err := execResult.RowsAffected()
	if err != nil {
		// A driver that cannot report the count still applied the statement. Reporting zero would
		// be a lie; leaving the counts at their defaults says "applied, count unknown".
		return result, nil
	}

	result.UpdatedCount = int(updated)
	matched, err := this.countMatching(ctx, orgId, userId, ids)
	if err != nil {
		return result, err
	}
	// Whatever this user owns and did not just change was already read. Ids belonging to someone
	// else never matched, so they are counted as neither.
	result.AlreadyReadCount = matched - result.UpdatedCount
	return result, nil
}

func (this *RecipientRepositoryImpl) countMatching(
	ctx corectx.Context, orgId model.Id, userId model.Id, ids []any,
) (int, error) {
	builder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	builder.Select("COUNT(1)")
	builder.From(recipientsTable + " AS r")
	builder.Where(
		builder.Equal("r.org_id", string(orgId)),
		builder.Equal("r.recipient_user_id", string(userId)),
		builder.In("r.notification_id", ids...),
	)
	scopeSelectToTenant(ctx, builder, "r")

	query, args := builder.Build()
	var count int
	row := this.GetBaseRepo().ExtractClient(ctx).QueryRow(ctx, query, args...)
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "countMatching")
	}
	return count, nil
}

func (this *RecipientRepositoryImpl) queryItems(
	ctx corectx.Context, builder *sqlbuilder.SelectBuilder, action string,
) ([]it.InboxItem, error) {
	query, args := builder.Build()
	rows, err := this.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, action)
	}
	defer rows.Close()

	items := make([]it.InboxItem, 0)
	for rows.Next() {
		item, err := scanInboxItem(rows)
		if err != nil {
			return nil, errors.Wrap(err, action)
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, action)
	}
	return items, nil
}

func scanInboxItem(rows *sql.Rows) (*it.InboxItem, error) {
	var (
		recipientId, notificationId            string
		title, message, severity, sourceModule string
		streamSeq                              int64
		readAt, createdAt                      sql.NullTime
		sourceResourceName                     sql.NullString
		sourceResourceKey, metadata            []byte
	)

	err := rows.Scan(
		&recipientId, &notificationId, &streamSeq, &readAt, &createdAt,
		&title, &message, &severity, &sourceModule,
		&sourceResourceName, &sourceResourceKey, &metadata,
	)
	if err != nil {
		return nil, err
	}

	item := it.InboxItem{
		NotificationId:    model.Id(notificationId),
		RecipientId:       model.Id(recipientId),
		StreamSeq:         streamSeq,
		Title:             title,
		Message:           message,
		Severity:          severity,
		SourceModule:      sourceModule,
		SourceResourceKey: decodeJsonMap(sourceResourceKey),
		Metadata:          decodeJsonMap(metadata),
	}
	if sourceResourceName.Valid {
		item.SourceResourceName = &sourceResourceName.String
	}
	if createdAt.Valid {
		item.CreatedAt = model.WrapModelDateTime(createdAt.Time).String()
	}
	if readAt.Valid {
		stamp := model.WrapModelDateTime(readAt.Time).String()
		item.ReadAt = &stamp
	}
	return &item, nil
}

// decodeJsonMap reads a jsonb column back into the shape the client is handed.
//
// A column that will not decode is data this module wrote and cannot read back. It is opaque
// pass-through for the client, so dropping it degrades one notification rather than failing a read
// the user is entitled to.
func decodeJsonMap(raw []byte) map[string]any {
	if len(raw) == 0 || strings.EqualFold(string(raw), "null") {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil
	}
	return value
}
