package repository

import (
	"database/sql"
	"encoding/json"
	"math"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itSeq "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/sequence"
)

func NewSequenceCounterRepository(base dynamicresource.CrudRepository) itSeq.SequenceCounterRepository {
	return &SequenceCounterRepositoryImpl{CrudRepository: base}
}

type SequenceCounterRepositoryImpl struct {
	dynamicresource.CrudRepository
}

// IncrementCounter is one statement: the first call of a scope and period inserts the counter at
// initial_value, every later one adds increment under the row lock the conflict takes. The WHERE on
// the update refuses an addition that would pass MaxInt64, and RETURNING then yields no row.
// algorithm_params is written on insert only, so a counter keeps the tweak it was born with.
func (this *SequenceCounterRepositoryImpl) IncrementCounter(
	ctx corectx.Context, param itSeq.IncrementCounterParam,
) (*itSeq.IncrementedCounter, error) {
	if ctx.GetDbTranx() == nil {
		return nil, errNoTransaction("IncrementCounter")
	}
	if param.Increment <= 0 {
		return nil, errors.Errorf("IncrementCounter: increment must be positive, got %d", param.Increment)
	}
	tenant, err := resolveTenant(ctx, this.Schema(), "IncrementCounter")
	if err != nil {
		return nil, err
	}
	query, args, err := buildIncrementCounterSql(this.Schema().TableName(), tenant, param, actingPrincipal(ctx))
	if err != nil {
		return nil, err
	}
	row := this.GetBaseRepo().ExtractClient(ctx).QueryRow(ctx, query, args...)
	var (
		id         string
		lastNumber int64
		params     []byte
	)
	err = row.Scan(&id, &lastNumber, &params)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "IncrementCounter")
	}
	return &itSeq.IncrementedCounter{Id: model.Id(id), LastNumber: lastNumber, AlgorithmParams: params}, nil
}

func buildIncrementCounterSql(
	table string, tenant tenantScope, param itSeq.IncrementCounterParam, principal any,
) (string, []any, error) {
	paramsJson, err := marshalJsonMap(param.AlgorithmParams)
	if err != nil {
		return "", nil, err
	}
	now := time.Now().UTC()
	args := &sqlArgs{}
	columns := []string{
		models.SeqCtrFieldId, models.SeqCtrFieldDefinitionId, models.SeqCtrFieldOrgId,
		models.SeqCtrFieldScopeObjectId, models.SeqCtrFieldScopeKey, models.SeqCtrFieldPeriodKey,
		models.SeqCtrFieldLastNumber, models.SeqCtrFieldAlgorithmParams,
		models.SeqCtrFieldCreatedAt, models.SeqCtrFieldUpdatedAt, models.SeqCtrFieldCreatedBy,
	}
	values := []string{
		args.bind(string(param.NewId)), args.bind(string(param.DefinitionId)), args.bind(nullableId(param.OrgId)),
		args.bind(nullableId(param.ScopeObjectId)), args.bind(param.ScopeKey), args.bind(param.PeriodKey),
		args.bind(param.InitialValue), args.bind(paramsJson) + "::jsonb",
		args.bind(now), args.bind(now), args.bind(principal),
	}
	columns, values = tenant.insertColumns(columns, values, args)
	last := table + "." + models.SeqCtrFieldLastNumber
	increment := args.bind(param.Increment)
	query := "INSERT INTO " + table + " (" + strings.Join(columns, ", ") + ") VALUES (" +
		strings.Join(values, ", ") + ")" +
		" ON CONFLICT (" + tenant.conflictColumns(models.SeqCtrFieldDefinitionId,
		models.SeqCtrFieldScopeKey, models.SeqCtrFieldPeriodKey) + ")" +
		" DO UPDATE SET " + models.SeqCtrFieldLastNumber + " = " + last + " + " + increment + "::bigint" +
		", " + models.SeqCtrFieldUpdatedAt + " = " + args.bind(now) +
		" WHERE " + last + " <= " + args.bind(int64(math.MaxInt64)) + "::bigint - " + increment + "::bigint" +
		" RETURNING " + models.SeqCtrFieldId + ", " + models.SeqCtrFieldLastNumber + ", " +
		"COALESCE(" + models.SeqCtrFieldAlgorithmParams + "::text, '{}')"
	return query, args.values, nil
}

func marshalJsonMap(fields dmodel.DynamicFields) (any, error) {
	if len(fields) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(fields)
	if err != nil {
		return nil, errors.Wrap(err, "marshal algorithm_params")
	}
	return string(raw), nil
}
