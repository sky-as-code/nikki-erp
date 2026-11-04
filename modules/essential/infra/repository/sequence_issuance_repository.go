package repository

import (
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func NewSequenceIssuanceRepository(base dynamicresource.CrudRepository) itSeq.SequenceIssuanceRepository {
	return &SequenceIssuanceRepositoryImpl{CrudRepository: base}
}

type SequenceIssuanceRepositoryImpl struct {
	dynamicresource.CrudRepository
}

// ReserveIssuance claims the request key before any counter moves. On a conflict Postgres waits for
// the transaction holding the key: a commit leaves the row to replay, a rollback lets this insert in.
func (this *SequenceIssuanceRepositoryImpl) ReserveIssuance(
	ctx corectx.Context, param itSeq.ReserveIssuanceParam,
) (bool, error) {
	if ctx.GetDbTranx() == nil {
		return false, errNoTransaction("ReserveIssuance")
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "ReserveIssuance")
	if err != nil {
		return false, err
	}
	query, args := buildReserveIssuanceSql(this.Schema().TableName(), tenant, param, actingPrincipal(ctx))
	rows, err := this.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return false, errors.Wrap(err, "ReserveIssuance")
	}
	defer rows.Close()
	inserted := rows.Next()
	return inserted, errors.Wrap(rows.Err(), "ReserveIssuance")
}

func buildReserveIssuanceSql(
	table string, tenant tenantScope, param itSeq.ReserveIssuanceParam, principal any,
) (string, []any) {
	args := &sqlArgs{}
	columns := []string{
		models.SeqIssFieldId, models.SeqIssFieldDefinitionId, models.SeqIssFieldOrgId,
		models.SeqIssFieldIdempotencyId, models.SeqIssFieldOperationKey, models.SeqIssFieldRequestFingerprint,
		models.SeqIssFieldCallerModule, models.SeqIssFieldCallerPrincipalId, models.SeqIssFieldAcceptedAt,
		models.SeqIssFieldCreatedAt, models.SeqIssFieldCreatedBy,
	}
	values := []string{
		args.bind(string(param.NewId)), args.bind(string(param.DefinitionId)), args.bind(nullableId(param.OrgId)),
		args.bind(param.IdempotencyId) + "::uuid", args.bind(param.OperationKey), args.bind(param.RequestFingerprint),
		args.bind(param.CallerModule), args.bind(nullableId(param.CallerPrincipalId)),
		args.bind(param.AcceptedAt.UTC()), args.bind(time.Now().UTC()), args.bind(principal),
	}
	columns, values = tenant.insertColumns(columns, values, args)
	query := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES (" +
		strings.Join(values, ", ") + ")" +
		" ON CONFLICT (" + tenant.conflictColumns(models.SeqIssFieldDefinitionId,
		models.SeqIssFieldIdempotencyId, models.SeqIssFieldOperationKey) + ") DO NOTHING" +
		" RETURNING " + models.SeqIssFieldId
	return query, args.values
}

func (this *SequenceIssuanceRepositoryImpl) CompleteIssuance(
	ctx corectx.Context, param itSeq.CompleteIssuanceParam,
) error {
	if ctx.GetDbTranx() == nil {
		return errNoTransaction("CompleteIssuance")
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "CompleteIssuance")
	if err != nil {
		return err
	}
	args := &sqlArgs{}
	query := "UPDATE " + this.Schema().TableName() + " SET " +
		models.SeqIssFieldCounterId + " = " + args.bind(string(param.CounterId)) + ", " +
		models.SeqIssFieldEffectiveAt + " = " + args.bind(param.EffectiveAt.UTC()) + ", " +
		models.SeqIssFieldAllocatedValue + " = " + args.bind(param.AllocatedValue) + ", " +
		models.SeqIssFieldFormattedValue + " = " + args.bind(param.FormattedValue) + ", " +
		models.SeqIssFieldIssuedAt + " = " + args.bind(param.IssuedAt.UTC()) + ", " +
		models.SeqIssFieldExpiresAt + " = " + args.bind(param.ExpiresAt.UTC()) +
		" WHERE " + models.SeqIssFieldId + " = " + args.bind(string(param.Id)) + tenant.predicate(args)
	result, err := this.GetBaseRepo().ExtractClient(ctx).Exec(ctx, query, args.values...)
	if err != nil {
		return errors.Wrap(err, "CompleteIssuance")
	}
	if affected, err := result.RowsAffected(); err == nil && affected != 1 {
		return errors.Errorf("CompleteIssuance: expected one reserved issuance '%s', updated %d", param.Id, affected)
	}
	return nil
}
