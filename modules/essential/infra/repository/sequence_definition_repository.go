package repository

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"time"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

// registrationLockSpace keeps the advisory keys of UpsertDefinition apart from any other module's
// advisory locks: the two-int form of pg_advisory_xact_lock has its own key space per first int.
const registrationLockSpace = int32(0x5345_5144) // "SEQD"

func NewSequenceDefinitionRepository(base dynamicresource.CrudRepository) itSeq.SequenceDefinitionRepository {
	return &SequenceDefinitionRepositoryImpl{CrudRepository: base}
}

type SequenceDefinitionRepositoryImpl struct {
	dynamicresource.CrudRepository
}

func (this *SequenceDefinitionRepositoryImpl) LockDefinition(
	ctx corectx.Context, id model.Id, mode itSeq.LockMode,
) (bool, error) {
	if ctx.GetDbTranx() == nil {
		return false, errNoTransaction("LockDefinition")
	}
	if mode != itSeq.LockModeShare && mode != itSeq.LockModeUpdate {
		return false, errors.Errorf("LockDefinition: unknown lock mode '%s'", mode)
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "LockDefinition")
	if err != nil {
		return false, err
	}
	args := &sqlArgs{}
	query := "SELECT " + models.SeqDefFieldId + " FROM " + this.Schema().TableName() +
		" WHERE " + models.SeqDefFieldId + " = " + args.bind(string(id)) + tenant.predicate(args) +
		" FOR " + string(mode)
	rows, err := this.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args.values...)
	if err != nil {
		return false, errors.Wrap(err, "LockDefinition")
	}
	defer rows.Close()
	found := rows.Next()
	return found, errors.Wrap(rows.Err(), "LockDefinition")
}

func (this *SequenceDefinitionRepositoryImpl) StampFirstIssued(
	ctx corectx.Context, id model.Id, at time.Time,
) (bool, error) {
	if ctx.GetDbTranx() == nil {
		return false, errNoTransaction("StampFirstIssued")
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "StampFirstIssued")
	if err != nil {
		return false, err
	}
	args := &sqlArgs{}
	query := "UPDATE " + this.Schema().TableName() +
		" SET " + models.SeqDefFieldFirstIssuedAt + " = " + args.bind(at.UTC()) +
		" WHERE " + models.SeqDefFieldId + " = " + args.bind(string(id)) + tenant.predicate(args) +
		" AND " + models.SeqDefFieldFirstIssuedAt + " IS NULL"
	result, err := this.GetBaseRepo().ExtractClient(ctx).Exec(ctx, query, args.values...)
	if err != nil {
		return false, errors.Wrap(err, "StampFirstIssued")
	}
	affected, err := result.RowsAffected()
	return affected > 0, errors.Wrap(err, "StampFirstIssued.RowsAffected")
}

// LockRegistration takes pg_advisory_xact_lock on a hash of the registration key. A hash collision
// only serializes two unrelated registrations; it never merges them, because the lookup that follows
// compares the real columns.
func (this *SequenceDefinitionRepositoryImpl) LockRegistration(ctx corectx.Context, key string) error {
	if ctx.GetDbTranx() == nil {
		return errNoTransaction("LockRegistration")
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "LockRegistration")
	if err != nil {
		return err
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	if tenant.column != "" {
		_, _ = hasher.Write([]byte("|" + fmt.Sprint(tenant.value)))
	}
	args := &sqlArgs{}
	query := "SELECT pg_advisory_xact_lock(" + args.bind(registrationLockSpace) + "::int4, " +
		args.bind(int32(hasher.Sum32())) + "::int4)"
	_, err = this.GetBaseRepo().ExtractClient(ctx).Exec(ctx, query, args.values...)
	return errors.Wrap(err, "LockRegistration")
}

func (this *SequenceDefinitionRepositoryImpl) FindIdByCode(
	ctx corectx.Context, code string, orgId *model.Id,
) (model.Id, error) {
	tenant, err := resolveTenant(ctx, this.Schema(), "FindIdByCode")
	if err != nil {
		return "", err
	}
	args := &sqlArgs{}
	orgPredicate := " AND " + models.SeqDefFieldOrgId + " IS NULL"
	if orgId != nil && *orgId != "" {
		orgPredicate = " AND " + models.SeqDefFieldOrgId + " = " + args.bind(string(*orgId))
	}
	query := "SELECT " + models.SeqDefFieldId + " FROM " + this.Schema().TableName() +
		" WHERE " + models.SeqDefFieldCode + " = " + args.bind(code) + orgPredicate + tenant.predicate(args)
	var id string
	err = this.GetBaseRepo().ExtractClient(ctx).QueryRow(ctx, query, args.values...).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return model.Id(id), errors.Wrap(err, "FindIdByCode")
}
