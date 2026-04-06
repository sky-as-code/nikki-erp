package baserepo

import (
	"database/sql"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type NewBaseRepositoryParam struct {
	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	Logger       logging.LoggerService
	QueryBuilder orm.QueryBuilder
	Schema       *dmodel.ModelSchema
}

func NewBaseDynamicRepository(param NewBaseRepositoryParam) dyn.BaseDynamicRepository {
	sqlDebugEnabled := param.ConfigSvc.GetBool(c.DbDebugEnabled)

	return &BaseDynamicRepositoryImpl{
		client:          param.Client,
		queryBuilder:    param.QueryBuilder,
		logger:          param.Logger,
		schema:          param.Schema,
		sqlDebugEnabled: sqlDebugEnabled,
	}
}

type BaseDynamicRepositoryImpl struct {
	client          orm.DbClient
	queryBuilder    orm.QueryBuilder
	logger          logging.LoggerService
	schema          *dmodel.ModelSchema
	sqlDebugEnabled bool
}

func (this *BaseDynamicRepositoryImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

func (this *BaseDynamicRepositoryImpl) ExtractClient(ctx corectx.Context) orm.DbClient {
	if ctx != nil {
		if tx := ctx.GetDbTranx(); tx != nil {
			if sqlClient, ok := tx.(orm.DbClient); ok {
				return sqlClient
			}
		}
	}
	return this.client
}

func (this *BaseDynamicRepositoryImpl) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.client.BeginTx(ctx.InnerContext(), nil)
}

func (this *BaseDynamicRepositoryImpl) ExecFunc(ctx corectx.Context, sqlFuncName string, sqlFuncArgs ...any) error {
	sqlFuncName = strings.TrimSpace(sqlFuncName)
	if sqlFuncName == "" {
		return errors.New("sql function name is required")
	}

	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString("SELECT ")
	sqlBuilder.WriteString(sqlFuncName)
	sqlBuilder.WriteString("(")
	for i := range sqlFuncArgs {
		if i > 0 {
			sqlBuilder.WriteString(", ")
		}
		sqlBuilder.WriteString(fmt.Sprintf("$p%d", i+1))
	}
	sqlBuilder.WriteString(")")

	sqlQuery := sqlBuilder.String()
	this.logQuery(sqlQuery)
	_, err := this.ExtractClient(ctx).Exec(ctx.InnerContext(), sqlQuery, sqlFuncArgs...)
	return err
}

// CheckUniqueCollisions returns unique key groups that have collisions. Empty slice means no collisions.
func (this *BaseDynamicRepositoryImpl) CheckUniqueCollisions(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[[][]string], error,
) {
	uniqueKeysToCheck := this.filterUniqueKeysWithValues(data)
	if len(uniqueKeysToCheck) == 0 {
		return &dyn.OpResult[[][]string]{Data: nil, HasData: false}, nil
	}

	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlCheckUniqueCollisions(this.schema, uniqueKeysToCheck, data)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[[][]string]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(sqlQuery.Sql)
	rows, err := this.ExtractClient(ctx).Query(ctx, sqlQuery.Sql, sqlQuery.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collidingKeys [][]string
	idx := 0
	for rows.Next() {
		var val int
		if err := rows.Scan(&val); err != nil {
			return nil, err
		}
		if val == 1 && idx < len(uniqueKeysToCheck) {
			collidingKeys = append(collidingKeys, uniqueKeysToCheck[idx])
		}
		idx++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &dyn.OpResult[[][]string]{Data: collidingKeys, HasData: len(collidingKeys) != 0}, nil
}

func (this *BaseDynamicRepositoryImpl) DeleteOne(
	ctx corectx.Context, keys dmodel.DynamicFields,
) (*dyn.OpResult[int], error) {
	if err := this.validateKeyMap(keys); err != nil {
		return nil, err
	}
	// if err := this.ensureTenantKeyIn(keys); err != nil {
	// 	return nil, err
	// }
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlDeleteEqual(this.schema, keys)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	result, err := this.ExtractClient(ctx).Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseDynamicRepositoryImpl) Exists(
	ctx corectx.Context, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	if len(keys) == 0 {
		return &dyn.OpResult[dyn.RepoExistsResult]{
			Data:    dyn.RepoExistsResult{},
			HasData: true,
		}, nil
	}
	return this.existsOnSchema(ctx, this.schema, keys)
}

func (this *BaseDynamicRepositoryImpl) existsOnSchema(
	ctx corectx.Context, schema *dmodel.ModelSchema, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	sqlRes, qbClientErrs, err := this.queryBuilder.SqlExistsMany(schema, keys)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.RepoExistsResult]{ClientErrors: *qbClientErrs}, nil
	}
	if sqlRes == nil {
		return &dyn.OpResult[dyn.RepoExistsResult]{
			Data:    dyn.RepoExistsResult{},
			HasData: true,
		}, nil
	}
	return this.runExistsManyQuery(ctx, *sqlRes, keys)
}

func (this *BaseDynamicRepositoryImpl) runExistsManyQuery(
	ctx corectx.Context, sqlData orm.SqlExistsManyData, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	this.logQuery(sqlData.Sql)
	rows, err := this.ExtractClient(ctx).Query(ctx, sqlData.Sql, sqlData.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanExistsManyRows(rows, keys)
}

func scanExistsManyRows(rows *sql.Rows, keys []dmodel.DynamicFields) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	var existing, notExisting []dmodel.DynamicFields
	i := 0
	for rows.Next() {
		var flag int
		if err := rows.Scan(&flag); err != nil {
			return nil, err
		}
		if i >= len(keys) {
			break
		}
		if flag == 1 {
			existing = append(existing, keys[i])
		} else {
			notExisting = append(notExisting, keys[i])
		}
		i++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := dyn.RepoExistsResult{Existing: existing, NotExisting: notExisting}
	return &dyn.OpResult[dyn.RepoExistsResult]{Data: out, HasData: true}, nil
}

// Insert inserts a record to database. If the schema defines "created_at", sets current UTC timestamp.
// On success, Data holds the inserted field map; HasData is true when Data is non-nil.
func (this *BaseDynamicRepositoryImpl) Insert(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[int], error,
) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	// this.trySetCreatedAt(data)
	// this.trySetEtag(data)
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlInsert(this.schema, data, false)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	result, err := this.ExtractClient(ctx).Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

// Implements BaseRepository interface
func (this *BaseDynamicRepositoryImpl) GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (
	*dyn.OpResult[dmodel.DynamicFields], error,
) {
	if vErr := this.validateGetOneColumnsAndFilter(param.Columns, param.Filter); vErr != nil {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: ft.ClientErrors{*vErr}}, nil
	}
	if this.hasNestedOrEdgeColumns(param.Columns) {
		return this.getOneWithNestedColumns(ctx, param)
	}
	// if err := this.ensureTenantKeyIn(param.Filter); err != nil {
	// 	return nil, err
	// }
	graph, err := this.buildFindOneGraph(param.Filter)
	if err != nil {
		return nil, err
	}
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(
		this.schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
			Columns: orm.ToSelectColumns(this.ensurePrimaryKeyColumns(param.Columns)),
		})
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	rows, err := this.queryAndScan(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{HasData: false}, nil
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: rows[0], HasData: true}, nil
}

func (this *BaseDynamicRepositoryImpl) getOneWithNestedColumns(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	graph, err := this.buildFindOneGraph(param.Filter)
	if err != nil {
		return nil, err
	}
	plan, cErrs := this.buildNestedSelectPlan(param.Columns)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: cErrs}, nil
	}
	rows, scanErrs, err := this.runSelectGraphScan(ctx, graph, dyn.RepoSearchParam{
		Columns: plan.MainColumns,
		Page:    0,
		Size:    1,
	})
	if err != nil {
		return nil, err
	}
	if len(scanErrs) > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: scanErrs}, nil
	}
	if len(rows) == 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{HasData: false}, nil
	}
	if err := this.hydrateNestedEdgesForRows(ctx, rows, plan.EdgeLeafColumns); err != nil {
		return nil, err
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: rows[0], HasData: true}, nil
}

func (this *BaseDynamicRepositoryImpl) ManageM2m(
	ctx corectx.Context, param dyn.RepoManageM2mParam,
) (*dyn.OpResult[int], error) {
	link, ok := this.schema.M2mPeerLinkForDest(param.DestSchemaName)
	if !ok {
		return nil, errors.Errorf(
			"ManageM2m: no M2M relation from '%s' to '%s'", this.schema.Name(), param.DestSchemaName,
		)
	}
	if overlapErrs := assertNoAssociationOverlap(param); overlapErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: overlapErrs}, nil
	}
	srcExistsErrs, err := this.assertSrcIdExists(ctx, param)
	if err != nil {
		return nil, err
	}
	if srcExistsErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: srcExistsErrs}, nil
	}
	associations, disassociations, prepClientErrs, err := this.prepareM2mAssociations(ctx, link, param)
	if err != nil {
		return nil, err
	}
	if prepClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: prepClientErrs}, nil
	}
	if len(associations) == 0 && len(disassociations) == 0 {
		return &dyn.OpResult[int]{Data: 0, HasData: false}, nil
	}
	return this.applyM2mAssociations(ctx, link, associations, disassociations, param.BeforeInsert)
}

func (this *BaseDynamicRepositoryImpl) ExistsM2m(ctx corectx.Context, param dyn.RepoExistsM2mParam) (bool, error) {
	link, ok := this.schema.M2mPeerLinkForEdge(param.M2mEdge)
	if !ok {
		return false, errors.Errorf(
			"ExistsM2m: no many-to-many edge '%s' on schema '%s'", param.M2mEdge, this.schema.Name(),
		)
	}
	srcKeys := dmodel.DynamicFields{basemodel.FieldId: param.SrcId}
	var row dmodel.DynamicFields
	if param.DestId == nil || *param.DestId == "" {
		filter, filled := this.m2mThroughFilterForLink(srcKeys, link)
		if !filled {
			return false, nil
		}
		row = filter
	} else {
		destKeys := dmodel.DynamicFields{basemodel.FieldId: *param.DestId}
		row = this.materializeM2mJunctionRow(link, dyn.RepoM2mAssociation{SrcKeys: srcKeys, DestKeys: destKeys})
	}
	graph := filterToAndGraph(row)
	sqlPtr, qbClientErrs, err := this.queryBuilder.SqlExistsGraph(
		link.ThroughSchema, dmodel.GetSchemaRegistry(), graph)
	if err != nil {
		return false, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return false, errors.Errorf("ExistsM2m: invalid query graph: %v", qbClientErrs)
	}
	if sqlPtr == nil {
		return false, nil
	}
	this.logQuery(*sqlPtr)
	return this.queryScalarBool(ctx, *sqlPtr)
}

func (this *BaseDynamicRepositoryImpl) CountM2m(
	ctx corectx.Context, param dyn.RepoCountM2mParam,
) (*dyn.OpResult[int], error) {
	link, ok := this.schema.M2mPeerLinkForEdge(param.M2mEdge)
	if !ok {
		return nil, errors.Errorf(
			"CountM2m: no many-to-many edge '%s' on schema '%s'", param.M2mEdge, this.schema.Name(),
		)
	}
	srcKeys := dmodel.DynamicFields{basemodel.FieldId: param.SrcId}
	filter, filled := this.m2mThroughFilterForLink(srcKeys, link)
	if !filled {
		return &dyn.OpResult[int]{Data: 0, HasData: false}, nil
	}
	graph := filterToAndGraph(filter)
	total, countClientErrs, err := this.countRowsMatchingGraphOnSchema(ctx, link.ThroughSchema, graph)
	if err != nil {
		return nil, err
	}
	if countClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: countClientErrs}, nil
	}
	return &dyn.OpResult[int]{Data: total, HasData: true}, nil
}

func assertNoAssociationOverlap(param dyn.RepoManageM2mParam) ft.ClientErrors {
	if len(param.AssociatedIds) == 0 || len(param.DisassociatedIds) == 0 {
		return nil
	}
	for id := range param.AssociatedIds {
		if param.DisassociatedIds.Contains(id) {
			return ft.ClientErrors{
				*ft.NewOverlappedFieldsError(
					[]string{basemodel.FieldAssociations, basemodel.FieldDesociations},
				),
			}
		}
	}
	return nil
}

func (this *BaseDynamicRepositoryImpl) prepareM2mAssociations(
	ctx corectx.Context, link *dmodel.M2mPeerLink, param dyn.RepoManageM2mParam,
) ([]dyn.RepoM2mAssociation, []dyn.RepoM2mAssociation, ft.ClientErrors, error) {
	validAssociatedIds, destClientErrs, err := this.validateAssociatedDestIds(ctx, link, param.AssociatedIds.ToSlice())
	if err != nil {
		return nil, nil, nil, err
	}
	if destClientErrs.Count() > 0 {
		return nil, nil, destClientErrs, nil
	}
	srcKeys := dmodel.DynamicFields{basemodel.FieldId: param.SrcId}
	associations := this.buildM2mAssociations(srcKeys, idsToDynamicFields(validAssociatedIds))
	disassociations := this.buildM2mAssociations(srcKeys, idsToDynamicFields(param.DisassociatedIds.ToSlice()))
	return associations, disassociations, nil, nil
}

func (this *BaseDynamicRepositoryImpl) applyM2mAssociations(
	ctx corectx.Context, link *dmodel.M2mPeerLink,
	associations []dyn.RepoM2mAssociation, disassociations []dyn.RepoM2mAssociation,
	beforeInsert dyn.RepoBeforeInsertM2mFn,
) (*dyn.OpResult[int], error) {
	total, cErrs, err := this.insertM2mAssociations(ctx, link, associations, beforeInsert)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: cErrs}, nil
	}
	deleted, cErrs, err := this.deleteM2mAssociations(ctx, link, disassociations)
	if err != nil {
		return nil, err
	}
	if cErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: cErrs}, nil
	}
	total += deleted
	return &dyn.OpResult[int]{Data: total, HasData: true}, nil
}

func (this *BaseDynamicRepositoryImpl) insertM2mAssociations(
	ctx corectx.Context, link *dmodel.M2mPeerLink, associations []dyn.RepoM2mAssociation,
	beforeInsert dyn.RepoBeforeInsertM2mFn,
) (int, ft.ClientErrors, error) {
	if len(associations) == 0 {
		return 0, nil, nil
	}
	insertRes, err := this.insertJunctionRows(ctx, link, associations, beforeInsert)
	if err != nil {
		return 0, nil, err
	}
	return insertRes.Data, insertRes.ClientErrors, nil
}

func (this *BaseDynamicRepositoryImpl) deleteM2mAssociations(
	ctx corectx.Context, link *dmodel.M2mPeerLink, disassociations []dyn.RepoM2mAssociation,
) (int, ft.ClientErrors, error) {
	if len(disassociations) == 0 {
		return 0, nil, nil
	}
	deleteRes, err := this.deleteJunctionRows(ctx, link, disassociations)
	if err != nil {
		return 0, nil, err
	}
	return deleteRes.Data, deleteRes.ClientErrors, nil
}

func (this *BaseDynamicRepositoryImpl) assertSrcIdExists(
	ctx corectx.Context, param dyn.RepoManageM2mParam,
) (ft.ClientErrors, error) {
	existsRes, err := this.Exists(ctx, []dmodel.DynamicFields{{basemodel.FieldId: param.SrcId}})
	if err != nil {
		return nil, err
	}
	if existsRes.ClientErrors.Count() > 0 {
		return existsRes.ClientErrors, nil
	}
	if len(existsRes.Data.NotExisting) == 0 {
		return nil, nil
	}
	errs := ft.ClientErrors{}
	fieldName := param.SrcIdFieldForError
	if fieldName == "" {
		fieldName = basemodel.FieldId
	}
	errs.Append(*ft.NewNotFoundError(fieldName))
	return errs, nil
}

func (this *BaseDynamicRepositoryImpl) validateAssociatedDestIds(
	ctx corectx.Context, link *dmodel.M2mPeerLink, associatedIds []model.Id,
) ([]model.Id, ft.ClientErrors, error) {
	if len(associatedIds) == 0 {
		return []model.Id{}, nil, nil
	}
	existsRes, err := this.existsOnSchema(ctx, link.DestSchema, idsToDynamicFields(associatedIds))
	if err != nil {
		return nil, nil, err
	}
	if existsRes.ClientErrors.Count() > 0 {
		return nil, existsRes.ClientErrors, nil
	}
	if len(existsRes.Data.NotExisting) > 0 {
		return nil, ft.ClientErrors{*ft.NewNotFoundValError(associatedIds)}, nil
	}
	return dynamicFieldsToIds(existsRes.Data.Existing), nil, nil
}

func (this *BaseDynamicRepositoryImpl) buildM2mAssociations(
	srcPk dmodel.DynamicFields, destPks []dmodel.DynamicFields,
) []dyn.RepoM2mAssociation {
	out := make([]dyn.RepoM2mAssociation, 0, len(destPks))
	for _, destPk := range destPks {
		out = append(out, dyn.RepoM2mAssociation{
			SrcKeys:  srcPk,
			DestKeys: destPk,
		})
	}
	return out
}

func idsToDynamicFields(ids []model.Id) []dmodel.DynamicFields {
	out := make([]dmodel.DynamicFields, 0, len(ids))
	for _, id := range ids {
		out = append(out, dmodel.DynamicFields{basemodel.FieldId: id})
	}
	return out
}

func dynamicFieldsToIds(fields []dmodel.DynamicFields) []model.Id {
	out := make([]model.Id, 0, len(fields))
	for _, item := range fields {
		raw, ok := item[basemodel.FieldId]
		if !ok || raw == nil {
			continue
		}
		if id, ok := raw.(model.Id); ok {
			out = append(out, id)
			continue
		}
		if id, ok := raw.(string); ok {
			out = append(out, model.Id(id))
		}
	}
	return out
}

func (this *BaseDynamicRepositoryImpl) deleteJunctionRows(ctx corectx.Context,
	link *dmodel.M2mPeerLink, desociations []dyn.RepoM2mAssociation,
) (*dyn.OpResult[int], error) {
	rows, clientErrs := this.assocToJuncRowArr(link, desociations)
	if len(clientErrs) > 0 {
		return &dyn.OpResult[int]{ClientErrors: clientErrs}, nil
	}
	sqlStr, qbClientErrs, err := this.queryBuilder.SqlDeleteOrAndEquals(link.ThroughSchema, rows)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}
	this.logQuery(*sqlStr)
	execRes, err := this.ExtractClient(ctx).Exec(ctx, *sqlStr)
	if err != nil {
		return nil, err
	}
	n, err := execRes.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseDynamicRepositoryImpl) insertJunctionRows(ctx corectx.Context,
	link *dmodel.M2mPeerLink, associations []dyn.RepoM2mAssociation,
	beforeInsert dyn.RepoBeforeInsertM2mFn,
) (*dyn.OpResult[int], error) {
	rows, clientErrs := this.assocToJuncRowArr(link, associations)
	if len(clientErrs) > 0 {
		return &dyn.OpResult[int]{ClientErrors: clientErrs}, nil
	}
	if beforeInsert != nil {
		err := beforeInsert(ctx, rows)
		if err != nil {
			return nil, err
		}
	}
	sqlRes, qbClientErrs, err := this.queryBuilder.SqlInsertBulk(link.ThroughSchema, rows, true)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[int]{ClientErrors: *qbClientErrs}, nil
	}
	this.logQuery(*sqlRes)
	execRes, err := this.ExtractClient(ctx).Exec(ctx, *sqlRes)
	if err != nil {
		return nil, err
	}
	n, err := execRes.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(n), HasData: n != 0}, nil
}

func (this *BaseDynamicRepositoryImpl) assocToJuncRowArr(
	link *dmodel.M2mPeerLink, associations []dyn.RepoM2mAssociation,
) ([]dmodel.DynamicFields, ft.ClientErrors) {
	var errs ft.ClientErrors
	out := make([]dmodel.DynamicFields, 0, len(associations))
	for i := range associations {
		row, rowErrs := this.assocToJuncRow(link, i, associations[i])
		if rowErrs.Count() > 0 {
			for _, e := range rowErrs {
				errs.Append(e)
			}
			continue
		}
		out = append(out, row)
	}
	if errs.Count() > 0 {
		return nil, errs
	}
	return out, nil
}

func (this *BaseDynamicRepositoryImpl) assocToJuncRow(
	link *dmodel.M2mPeerLink, idx int, assoc dyn.RepoM2mAssociation,
) (dmodel.DynamicFields, ft.ClientErrors) {
	var errs ft.ClientErrors
	prefix := fmt.Sprintf("associations[%d]", idx)
	if assoc.SrcKeys == nil || assoc.DestKeys == nil {
		errs.Append(*ft.NewValidationError(
			prefix,
			"common.err_missing_required_field",
			"src and dest key maps are required",
		))
		return nil, errs
	}
	appendMissingKeyErrors(&errs, prefix+".srcKeys", assoc.SrcKeys, this.schema.KeyColumns())
	appendMissingKeyErrors(&errs, prefix+".destKeys", assoc.DestKeys, link.DestSchema.KeyColumns())
	if errs.Count() > 0 {
		return nil, errs
	}
	return this.materializeM2mJunctionRow(link, assoc), nil
}

func (this *BaseDynamicRepositoryImpl) materializeM2mJunctionRow(
	link *dmodel.M2mPeerLink, assoc dyn.RepoM2mAssociation,
) dmodel.DynamicFields {
	row := make(dmodel.DynamicFields)
	for _, k := range this.schema.PrimaryKeys() {
		row[dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, k)] = assoc.SrcKeys[k]
	}
	if tk := this.schema.TenantKey(); tk != "" {
		col := dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, tk)
		row[col] = assoc.SrcKeys[tk]
	}
	for _, k := range link.DestSchema.PrimaryKeys() {
		row[dmodel.PrefixedThroughColumn(link.DestFieldPrefix, k)] = assoc.DestKeys[k]
	}
	return row
}

func appendMissingKeyErrors(errs *ft.ClientErrors, fieldPrefix string, keys dmodel.DynamicFields, required []string) {
	for _, k := range required {
		if _, ok := keys[k]; !ok || keys[k] == nil {
			errs.Append(*dmodel.NewMissingFieldErr(fieldPrefix + "." + k))
		}
	}
}

// Search fetches records matching the SearchGraph criteria.
// When the schema is tenant-scoped, filter must be provided and contain the tenant key.
// Data uses PagedResult: Total is from COUNT when Size > 0, otherwise len(Items).
func (this *BaseDynamicRepositoryImpl) Search(ctx corectx.Context, param dyn.RepoSearchParam) (
	*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error,
) {
	if this.hasNestedOrEdgeColumns(param.Columns) {
		return this.searchWithNestedColumns(ctx, param)
	}
	// merged, err := this.mergeFilterIntoGraph(param.Graph, param.Filter)
	// if err != nil {
	// 	return nil, err
	// }
	merged := param.Graph
	page := param.Page
	size := param.Size
	var total int
	total, countClientErrs, err := this.countRowsMatchingGraph(ctx, merged)
	if err != nil {
		return nil, err
	}
	if len(countClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: countClientErrs,
		}, nil
	}
	rows, scanClientErrs, err := this.runSelectGraphScan(ctx, merged, dyn.RepoSearchParam{
		Columns: this.ensurePrimaryKeyColumns(param.Columns),
		Page:    param.Page,
		Size:    param.Size,
	})
	if err != nil {
		return nil, err
	}
	if len(scanClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: scanClientErrs,
		}, nil
	}
	if size <= 0 {
		total = len(rows)
	}
	paged := dyn.PagedResultData[dmodel.DynamicFields]{
		Items: rows,
		Total: total,
		Page:  page,
		Size:  size,
	}
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    paged,
		HasData: len(rows) != 0,
	}, nil
}

func (this *BaseDynamicRepositoryImpl) searchWithNestedColumns(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	plan, cErrs := this.buildNestedSelectPlan(param.Columns)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{ClientErrors: cErrs}, nil
	}
	merged := param.Graph
	total, countClientErrs, err := this.countRowsMatchingGraph(ctx, merged)
	if err != nil {
		return nil, err
	}
	if len(countClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: countClientErrs,
		}, nil
	}
	rows, scanClientErrs, err := this.runSelectGraphScan(ctx, merged, dyn.RepoSearchParam{
		Columns: plan.MainColumns,
		Page:    param.Page,
		Size:    param.Size,
	})
	if err != nil {
		return nil, err
	}
	if len(scanClientErrs) > 0 {
		return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
			ClientErrors: scanClientErrs,
		}, nil
	}
	if err := this.hydrateNestedEdgesForRows(ctx, rows, plan.EdgeLeafColumns); err != nil {
		return nil, err
	}
	if param.Size <= 0 {
		total = len(rows)
	}
	paged := dyn.PagedResultData[dmodel.DynamicFields]{
		Items: rows,
		Total: total,
		Page:  param.Page,
		Size:  param.Size,
	}
	return &dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]]{
		Data:    paged,
		HasData: len(rows) != 0,
	}, nil
}

type nestedSelectPlan struct {
	MainColumns     []string
	EdgeLeafColumns map[string][]string
}

func (this *BaseDynamicRepositoryImpl) hasNestedOrEdgeColumns(columns []string) bool {
	for _, col := range columns {
		if strings.Contains(col, ".") {
			return true
		}
		field, ok := this.schema.Field(col)
		if ok && field.IsVirtualModelField() {
			return true
		}
	}
	return false
}

func (this *BaseDynamicRepositoryImpl) buildNestedSelectPlan(columns []string) (nestedSelectPlan, ft.ClientErrors) {
	var errs ft.ClientErrors
	mainSet := make(map[string]struct{})
	for _, key := range this.schema.PrimaryKeys() {
		mainSet[key] = struct{}{}
	}
	edgeLeafSet := make(map[string]map[string]struct{})
	for _, col := range columns {
		if strings.Count(col, ".") == 0 {
			field, ok := this.schema.Field(col)
			if ok && field.IsVirtualModelField() {
				rel, hasRel := this.relationByEdge(col)
				if !hasRel {
					errs.Append(*ft.NewValidationError(
						col, ft.ErrorKey("err_unknown_schema_field"), "edge is not defined on this schema",
					))
					continue
				}
				destSchema := dmodel.GetSchemaRegistry().Get(rel.DestSchemaName)
				if destSchema == nil {
					errs.Append(*ft.NewAnonymousValidationError(
						ft.ErrorKey("err_schema_not_found"), "edge destination schema not found", nil,
					))
					continue
				}
				if edgeLeafSet[col] == nil {
					edgeLeafSet[col] = make(map[string]struct{})
				}
				for _, edgeCol := range physicalColumnNames(destSchema) {
					edgeLeafSet[col][edgeCol] = struct{}{}
				}
				continue
			}
			mainSet[col] = struct{}{}
			continue
		}
		parts, partErr := this.parseNestedColumn(col)
		if partErr != nil {
			errs.Append(*partErr)
			continue
		}
		if edgeLeafSet[parts[0]] == nil {
			edgeLeafSet[parts[0]] = make(map[string]struct{})
		}
		edgeLeafSet[parts[0]][parts[1]] = struct{}{}
	}
	if errs.Count() > 0 {
		return nestedSelectPlan{}, errs
	}
	mainCols := mapKeysSorted(mainSet)
	edgeCols := make(map[string][]string, len(edgeLeafSet))
	for edge, fields := range edgeLeafSet {
		edgeCols[edge] = mapKeysSorted(fields)
	}
	return nestedSelectPlan{
		MainColumns:     mainCols,
		EdgeLeafColumns: edgeCols,
	}, nil
}

func mapKeysSorted(in map[string]struct{}) []string {
	out := make([]string, 0, len(in))
	for key := range in {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (this *BaseDynamicRepositoryImpl) parseNestedColumn(col string) ([]string, *ft.ClientErrorItem) {
	if strings.Count(col, ".") > orm.MaxSelectGraphColumnDots {
		return nil, ft.NewValidationError(
			col, ft.ErrorKey("err_graph_field_path_too_deep"),
			fmt.Sprintf("field path exceeds maximum of %d dot separators", orm.MaxSelectGraphColumnDots),
		)
	}
	parts := strings.Split(col, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, ft.NewValidationError(
			col, ft.ErrorKey("err_invalid_graph_field_path"),
			"field path must be {edge}.{field}",
		)
	}
	rel, ok := this.relationByEdge(parts[0])
	if !ok {
		return nil, ft.NewValidationError(parts[0], ft.ErrorKey("err_unknown_schema_field"), "edge is not defined on this schema")
	}
	destSchema := dmodel.GetSchemaRegistry().Get(rel.DestSchemaName)
	if destSchema == nil {
		return nil, ft.NewAnonymousValidationError(ft.ErrorKey("err_schema_not_found"), "edge destination schema not found", nil)
	}
	f, ok := destSchema.Column(parts[1])
	if !ok || f.IsVirtualModelField() {
		return nil, ft.NewValidationError(col, ft.ErrorKey("err_unknown_schema_field"), "field is not defined on edge schema")
	}
	return parts, nil
}

func physicalColumnNames(schema *dmodel.ModelSchema) []string {
	cols := schema.Columns()
	out := make([]string, 0, len(cols))
	for _, col := range cols {
		out = append(out, col.Name())
	}
	return out
}

func (this *BaseDynamicRepositoryImpl) relationByEdge(edge string) (dmodel.ModelRelation, bool) {
	for _, rel := range this.schema.Relations() {
		if rel.Edge == edge {
			return rel, true
		}
	}
	return dmodel.ModelRelation{}, false
}

func (this *BaseDynamicRepositoryImpl) hydrateNestedEdgesForRows(
	ctx corectx.Context, rows []dmodel.DynamicFields, edgeLeafColumns map[string][]string,
) error {
	for i := range rows {
		for edge, leafCols := range edgeLeafColumns {
			val, err := this.fetchNestedEdgeValue(ctx, rows[i], edge, leafCols)
			if err != nil {
				return err
			}
			rows[i][edge] = val
		}
	}
	return nil
}

func (this *BaseDynamicRepositoryImpl) fetchNestedEdgeValue(
	ctx corectx.Context, srcRow dmodel.DynamicFields, edge string, leafColumns []string,
) (any, error) {
	rel, ok := this.relationByEdge(edge)
	if !ok {
		return nil, errors.Errorf("fetchNestedEdgeValue: unknown edge '%s'", edge)
	}
	destSchema := dmodel.GetSchemaRegistry().Get(rel.DestSchemaName)
	if destSchema == nil {
		return nil, errors.Errorf("fetchNestedEdgeValue: schema '%s' not found", rel.DestSchemaName)
	}
	destCols := withPrimaryKeys(destSchema.PrimaryKeys(), leafColumns)
	switch rel.RelationType {
	case dmodel.RelationTypeManyToMany:
		return this.fetchManyToManyEdgeRows(ctx, srcRow, rel, destSchema, destCols)
	case dmodel.RelationTypeOneToMany:
		return this.fetchOneToManyEdgeRows(ctx, srcRow, rel, destSchema, destCols)
	case dmodel.RelationTypeManyToOne, dmodel.RelationTypeOneToOne:
		return this.fetchSingleEdgeRow(ctx, srcRow, rel, destSchema, destCols)
	default:
		return nil, errors.Errorf("fetchNestedEdgeValue: unsupported relation type '%s'", rel.RelationType)
	}
}

func withPrimaryKeys(primaryKeys []string, columns []string) []string {
	set := make(map[string]struct{})
	for _, key := range primaryKeys {
		set[key] = struct{}{}
	}
	for _, col := range columns {
		set[col] = struct{}{}
	}
	return mapKeysSorted(set)
}

func (this *BaseDynamicRepositoryImpl) fetchSingleEdgeRow(
	ctx corectx.Context, srcRow dmodel.DynamicFields, rel dmodel.ModelRelation,
	destSchema *dmodel.ModelSchema, columns []string,
) (dmodel.DynamicFields, error) {
	filter, ok := this.filterForSingleEdge(srcRow, rel)
	if !ok {
		return nil, nil
	}
	rows, err := this.selectRowsByFilter(ctx, destSchema, filter, columns, 1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func (this *BaseDynamicRepositoryImpl) filterForSingleEdge(
	srcRow dmodel.DynamicFields, rel dmodel.ModelRelation,
) (dmodel.DynamicFields, bool) {
	filter := make(dmodel.DynamicFields)
	for _, pair := range rel.EffectiveForeignKeys() {
		srcVal, ok := srcRow[pair.FkColumn]
		if !ok || srcVal == nil {
			return nil, false
		}
		filter[pair.ReferencedColumn] = srcVal
	}
	return filter, true
}

func (this *BaseDynamicRepositoryImpl) fetchOneToManyEdgeRows(
	ctx corectx.Context, srcRow dmodel.DynamicFields, rel dmodel.ModelRelation,
	destSchema *dmodel.ModelSchema, columns []string,
) ([]dmodel.DynamicFields, error) {
	filter := make(dmodel.DynamicFields)
	for _, pair := range rel.EffectiveForeignKeys() {
		srcVal, ok := srcRow[pair.ReferencedColumn]
		if !ok || srcVal == nil {
			return []dmodel.DynamicFields{}, nil
		}
		filter[pair.FkColumn] = srcVal
	}
	return this.selectRowsByFilter(ctx, destSchema, filter, columns, 0)
}

func (this *BaseDynamicRepositoryImpl) fetchManyToManyEdgeRows(
	ctx corectx.Context, srcRow dmodel.DynamicFields, rel dmodel.ModelRelation,
	destSchema *dmodel.ModelSchema, columns []string,
) ([]dmodel.DynamicFields, error) {
	throughSchema := dmodel.GetSchemaRegistry().Get(rel.M2mThroughSchemaName)
	if throughSchema == nil {
		return nil, errors.Errorf("fetchManyToManyEdgeRows: through schema '%s' not found", rel.M2mThroughSchemaName)
	}
	filters, ok := this.buildM2mThroughFilter(srcRow, rel)
	if !ok {
		return []dmodel.DynamicFields{}, nil
	}
	destKeys, err := this.selectDestKeyFiltersFromThrough(ctx, throughSchema, rel, filters, destSchema)
	if err != nil {
		return nil, err
	}
	if len(destKeys) == 0 {
		return []dmodel.DynamicFields{}, nil
	}
	return this.selectRowsByAnyFilter(ctx, destSchema, destKeys, columns)
}

func (this *BaseDynamicRepositoryImpl) buildM2mThroughFilter(
	srcRow dmodel.DynamicFields, rel dmodel.ModelRelation,
) (dmodel.DynamicFields, bool) {
	filter := make(dmodel.DynamicFields)
	for _, srcPk := range this.schema.PrimaryKeys() {
		val, ok := srcRow[srcPk]
		if !ok || val == nil {
			return nil, false
		}
		filter[dmodel.PrefixedThroughColumn(rel.M2mSrcFieldPrefix, srcPk)] = val
	}
	srcTk := this.schema.TenantKey()
	if srcTk != "" {
		if val, ok := srcRow[srcTk]; ok && val != nil {
			filter[dmodel.PrefixedThroughColumn(rel.M2mSrcFieldPrefix, srcTk)] = val
		}
	}
	return filter, true
}

func (this *BaseDynamicRepositoryImpl) selectDestKeyFiltersFromThrough(
	ctx corectx.Context, throughSchema *dmodel.ModelSchema, rel dmodel.ModelRelation,
	throughFilter dmodel.DynamicFields, destSchema *dmodel.ModelSchema,
) ([]dmodel.DynamicFields, error) {
	destPrefixedPks := make([]string, 0, len(destSchema.PrimaryKeys()))
	for _, key := range destSchema.PrimaryKeys() {
		destPrefixedPks = append(destPrefixedPks, dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, key))
	}
	destTk := destSchema.TenantKey()
	if destTk != "" {
		destPrefixedPks = append(destPrefixedPks, dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, destTk))
	}
	throughRows, err := this.selectRowsByFilter(ctx, throughSchema, throughFilter, destPrefixedPks, 0)
	if err != nil {
		return nil, err
	}
	out := make([]dmodel.DynamicFields, 0, len(throughRows))
	for _, row := range throughRows {
		item := make(dmodel.DynamicFields)
		for _, key := range destSchema.PrimaryKeys() {
			v := row[dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, key)]
			item[key] = v
		}
		if destTk != "" {
			v, ok := row[dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, destTk)]
			if ok && v != nil {
				item[destTk] = v
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (this *BaseDynamicRepositoryImpl) selectRowsByFilter(
	ctx corectx.Context, schema *dmodel.ModelSchema, filter dmodel.DynamicFields, columns []string, size int,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	nodes := make([]dmodel.SearchNode, 0, len(filter))
	for field, value := range filter {
		nodes = append(nodes, *dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))
	}
	graph.And(nodes...)
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
		Columns: orm.ToSelectColumns(columns),
		Size:    size,
	})
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, errors.Errorf("selectRowsByFilter: invalid query graph")
	}
	this.logQuery(*sqlQuery)
	return this.queryAndScan(ctx, *sqlQuery)
}

func (this *BaseDynamicRepositoryImpl) selectRowsByAnyFilter(
	ctx corectx.Context, schema *dmodel.ModelSchema, filters []dmodel.DynamicFields, columns []string,
) ([]dmodel.DynamicFields, error) {
	graph := &dmodel.SearchGraph{}
	ors := make([]dmodel.SearchNode, 0, len(filters))
	for _, filter := range filters {
		andNodes := make([]dmodel.SearchNode, 0, len(filter))
		for field, value := range filter {
			andNodes = append(andNodes, *dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))
		}
		node := dmodel.NewSearchNode()
		node.And(andNodes...)
		ors = append(ors, *node)
	}
	graph.Or(ors...)
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
		Columns: orm.ToSelectColumns(columns),
	})
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, errors.Errorf("selectRowsByAnyFilter: invalid query graph")
	}
	this.logQuery(*sqlQuery)
	return this.queryAndScan(ctx, *sqlQuery)
}

func (this *BaseDynamicRepositoryImpl) countRowsMatchingGraph(
	ctx corectx.Context, graph *dmodel.SearchGraph,
) (int, ft.ClientErrors, error) {
	return this.countRowsMatchingGraphOnSchema(ctx, this.schema, graph)
}

func (this *BaseDynamicRepositoryImpl) countRowsMatchingGraphOnSchema(
	ctx corectx.Context, schema *dmodel.ModelSchema, graph *dmodel.SearchGraph,
) (int, ft.ClientErrors, error) {
	qbRes, qbClientErrs, err := this.queryBuilder.SqlCountGraph(schema, dmodel.GetSchemaRegistry(), graph)
	if err != nil {
		return 0, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return 0, *qbClientErrs, nil
	}
	countSql := *qbRes
	this.logQuery(countSql)
	n, ierr := this.queryScalarInt(ctx, countSql)
	return n, nil, ierr
}

func (this *BaseDynamicRepositoryImpl) runSelectGraphScan(
	ctx corectx.Context, graph *dmodel.SearchGraph, param dyn.RepoSearchParam,
) ([]dmodel.DynamicFields, ft.ClientErrors, error) {
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlSelectGraph(
		this.schema, dmodel.GetSchemaRegistry(), graph, orm.SqlSelectGraphOpts{
			Columns: orm.ToSelectColumns(param.Columns),
			Page:    param.Page,
			Size:    param.Size,
		})
	if err != nil {
		return nil, nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return nil, *qbClientErrs, nil
	}

	this.logQuery(*sqlQuery)
	rows, err := this.queryAndScan(ctx, *sqlQuery)
	if err != nil {
		return nil, nil, err
	}
	if rows == nil {
		return []dmodel.DynamicFields{}, nil, nil
	}
	return rows, nil, nil
}

func (this *BaseDynamicRepositoryImpl) queryScalarInt(ctx corectx.Context, query string) (int, error) {
	row := this.ExtractClient(ctx).QueryRow(ctx, query)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func (this *BaseDynamicRepositoryImpl) queryScalarBool(ctx corectx.Context, query string) (bool, error) {
	row := this.ExtractClient(ctx).QueryRow(ctx, query)
	var v bool
	if err := row.Scan(&v); err != nil {
		return false, err
	}
	return v, nil
}

func filterToAndGraph(filter dmodel.DynamicFields) *dmodel.SearchGraph {
	graph := &dmodel.SearchGraph{}
	nodes := make([]dmodel.SearchNode, 0, len(filter))
	for field, value := range filter {
		nodes = append(nodes, *dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))
	}
	graph.And(nodes...)
	return graph
}

func (this *BaseDynamicRepositoryImpl) m2mThroughFilterForLink(
	srcKeys dmodel.DynamicFields, link *dmodel.M2mPeerLink,
) (dmodel.DynamicFields, bool) {
	filter := make(dmodel.DynamicFields)
	for _, srcPk := range this.schema.PrimaryKeys() {
		val, ok := srcKeys[srcPk]
		if !ok || val == nil {
			return nil, false
		}
		filter[dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, srcPk)] = val
	}
	srcTk := this.schema.TenantKey()
	if srcTk != "" {
		if val, ok := srcKeys[srcTk]; ok && val != nil {
			filter[dmodel.PrefixedThroughColumn(link.SrcFieldPrefix, srcTk)] = val
		}
	}
	return filter, true
}

// Update updates a record. The data map must contain primary keys and tenant key.
// If the schema defines "updated_at", sets current UTC timestamp.
func (this *BaseDynamicRepositoryImpl) Update(ctx corectx.Context, data dmodel.DynamicFields) (
	*dyn.OpResult[dmodel.DynamicFields], error,
) {
	// TODO: Extract later
	// if err := this.ensureTenantKeyIn(data); err != nil {
	// 	return nil, err
	// }
	filters := this.extractKeyMap(data)
	this.trySetUpdatedAt(data)
	prevEtag := this.trySetEtag(data)
	if len(prevEtag) > 0 {
		filters[basemodel.FieldEtag] = prevEtag
	}
	sqlQuery, qbClientErrs, err := this.queryBuilder.SqlUpdateEqual(this.schema, data, filters)
	if err != nil {
		return nil, err
	}
	if qbClientErrs != nil && qbClientErrs.Count() > 0 {
		return &dyn.OpResult[dmodel.DynamicFields]{ClientErrors: *qbClientErrs}, nil
	}

	this.logQuery(*sqlQuery)
	_, err = this.ExtractClient(ctx).Exec(ctx, *sqlQuery)
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[dmodel.DynamicFields]{Data: data, HasData: true}, nil
}

func (this *BaseDynamicRepositoryImpl) logQuery(query string) {
	if this.sqlDebugEnabled {
		this.logger.Debug(query, nil)
	}
}

func (this *BaseDynamicRepositoryImpl) filterUniqueKeysWithValues(data dmodel.DynamicFields) [][]string {
	var result [][]string
	for _, uniqueFields := range this.schema.AllUniques() {
		if len(uniqueFields) == 0 {
			continue
		}
		hasAll := true
		for _, f := range uniqueFields {
			if _, ok := data[f]; !ok || data[f] == nil {
				hasAll = false
				break
			}
		}
		if this.schema.TenantKey() != "" {
			if _, ok := data[this.schema.TenantKey()]; !ok || data[this.schema.TenantKey()] == nil {
				hasAll = false
			}
		}
		if hasAll {
			result = append(result, uniqueFields)
		}
	}
	return result
}

// func (this *BaseRepositoryImpl) trySetArchivedAt(data dmodel.DynamicFields) error {
// 	if _, ok := this.schema.Column(basemodel.FieldArchivedAt); !ok {
// 		return errors.Errorf(
// 			"Archive: schema '%s' does not define field '%s'", this.schema.Name(), basemodel.FieldArchivedAt)
// 	}
// 	field := this.schema.MustField(basemodel.FieldArchivedAt)
// 	data[basemodel.FieldArchivedAt] = *field.DataType().DefaultValue().Get()
// 	return nil
// }

func (this *BaseDynamicRepositoryImpl) trySetCreatedAt(data dmodel.DynamicFields) {
	if _, ok := this.schema.Column(basemodel.FieldCreatedAt); !ok {
		return
	}
	field := this.schema.MustField(basemodel.FieldCreatedAt)
	data[basemodel.FieldCreatedAt] = *field.DataType().DefaultValue().Get()
}

func (this *BaseDynamicRepositoryImpl) trySetUpdatedAt(data dmodel.DynamicFields) {
	if _, ok := this.schema.Column(basemodel.FieldUpdatedAt); !ok {
		return
	}
	field := this.schema.MustField(basemodel.FieldUpdatedAt)
	data[basemodel.FieldUpdatedAt] = *field.DataType().DefaultValue().Get()
}

func (this *BaseDynamicRepositoryImpl) trySetEtag(data dmodel.DynamicFields) (prevEtag string) {
	if _, ok := this.schema.Column(basemodel.FieldEtag); !ok {
		return ""
	}
	field, ok := this.schema.Field(basemodel.FieldEtag)
	if !ok {
		return ""
	}
	et := data[basemodel.FieldEtag]
	if et != nil {
		prevEtag = et.(string)
	}
	data[basemodel.FieldEtag] = *field.DataType().DefaultValue().Get()
	return
}

func (this *BaseDynamicRepositoryImpl) extractKeyMap(data dmodel.DynamicFields) dmodel.DynamicFields {
	if data == nil {
		return nil
	}
	result := make(dmodel.DynamicFields)
	for _, key := range this.schema.KeyColumns() {
		if v, ok := data[key]; ok {
			result[key] = v
			delete(data, key)
		}
	}
	return result
}

func (this *BaseDynamicRepositoryImpl) validateKeyMap(keys dmodel.DynamicFields) error {
	if len(keys) == 0 {
		return errors.New("validateKeyMap: keys map is required")
	}
	for _, key := range this.schema.KeyColumns() {
		if _, ok := keys[key]; !ok {
			return errors.Errorf("validateKeyMap: missing required key '%s'", key)
		}
	}
	return nil
}

func (this *BaseDynamicRepositoryImpl) validateGetOneColumnsAndFilter(
	columns []string, filter dmodel.DynamicFields,
) *ft.ClientErrorItem {
	for _, col := range columns {
		field, hasField := this.schema.Field(col)
		if hasField && field.IsVirtualModelField() {
			if _, ok := this.relationByEdge(col); !ok {
				return ft.NewValidationError(
					col,
					ft.ErrorKey("err_unknown_schema_field"),
					"edge is not defined on this schema",
				)
			}
			continue
		}
		if strings.Contains(col, ".") {
			if _, err := this.parseNestedColumn(col); err != nil {
				return err
			}
			continue
		}
		field, ok := this.schema.Column(col)
		if !ok || field.IsVirtualModelField() {
			return ft.NewValidationError(
				col,
				ft.ErrorKey("err_unknown_schema_field"),
				"field is not defined on this schema",
			)
		}
	}
	if filter == nil {
		return nil
	}
	keys := make([]string, 0, len(filter))
	for k := range filter {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, ok := this.schema.Column(k); !ok {
			return ft.NewValidationError(
				k,
				ft.ErrorKey("err_unknown_schema_field"),
				"field is not defined on this schema",
			)
		}
	}
	return nil
}

func (this *BaseDynamicRepositoryImpl) ensurePrimaryKeyColumns(columns []string) []string {
	if len(columns) == 0 {
		return columns
	}
	set := make(map[string]struct{})
	for _, col := range columns {
		set[col] = struct{}{}
	}
	for _, key := range this.schema.PrimaryKeys() {
		set[key] = struct{}{}
	}
	return mapKeysSorted(set)
}

func (this *BaseDynamicRepositoryImpl) ensureTenantKeyIn(values dmodel.DynamicFields) error {
	key := this.schema.TenantKey()
	if key == "" {
		return nil
	}
	if _, ok := values[key]; !ok {
		return errors.Errorf("ensureTenantKeyIn: missing tenant key '%s'", key)
	}
	return nil
}

func (this *BaseDynamicRepositoryImpl) mergeFilterIntoGraph(
	graph *dmodel.SearchGraph, filter []dmodel.DynamicFields,
) (*dmodel.SearchGraph, error) {
	if len(filter) == 0 {
		if key := this.schema.TenantKey(); key != "" {
			return nil, errors.Errorf(
				"mergeFilterIntoGraph: filter required for tenant-scoped schema, must contain '%s'", key)
		}
		return graph, nil
	}
	f := filter[0]
	if err := this.ensureTenantKeyIn(f); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(f))
	for k := range f {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	filterNodes := make([]dmodel.SearchNode, 0, len(keys))
	for _, k := range keys {
		n := dmodel.NewSearchNode().NewCondition(k, dmodel.Equals, f[k])
		filterNodes = append(filterNodes, *n)
	}
	merged := graph
	merged.And(filterNodes...)
	return merged, nil
}

func isNilAnyValue(val any) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

func (this *BaseDynamicRepositoryImpl) shouldIncludeEqualFilterField(field string, val any) bool {
	if isNilAnyValue(val) {
		return false
	}
	_, ok := this.schema.Column(field)
	return ok
}

func missingKeyColumnNames(found map[string]bool, keyColumns []string) []string {
	missing := make([]string, 0, len(keyColumns))
	for _, k := range keyColumns {
		if !found[k] {
			missing = append(missing, k)
		}
	}
	return missing
}

// buildEqualNodes builds an Equals SearchNode for each defined, non-nil filter field.
// Unknown keys and nil values are ignored. Key columns must still be present (non-nil).
func (this *BaseDynamicRepositoryImpl) buildEqualNodes(filter dmodel.DynamicFields) ([]dmodel.SearchNode, error) {
	keyColumns := this.schema.KeyColumns()
	found := make(map[string]bool, len(keyColumns))
	for _, k := range keyColumns {
		found[k] = false
	}

	nodes := make([]dmodel.SearchNode, 0, len(filter))
	for field, val := range filter {
		if !this.shouldIncludeEqualFilterField(field, val) {
			continue
		}
		n := dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, val)
		nodes = append(nodes, *n)
		if _, isKey := found[field]; isKey {
			found[field] = true
		}
	}

	missing := missingKeyColumnNames(found, keyColumns)
	if len(missing) > 0 {
		return nil, errors.Errorf(
			"buildEqualNodes: missing required key columns: %s", strings.Join(missing, ", "))
	}
	return nodes, nil
}

func (this *BaseDynamicRepositoryImpl) buildFindOneGraph(filter dmodel.DynamicFields) (*dmodel.SearchGraph, error) {
	allNodes, err := this.buildEqualNodes(filter)
	if err != nil {
		return &dmodel.SearchGraph{}, err
	}
	g := &dmodel.SearchGraph{}
	g.And(allNodes...)
	return g, nil
}

func (this *BaseDynamicRepositoryImpl) queryAndScan(ctx corectx.Context, query string) ([]dmodel.DynamicFields, error) {
	rows, err := this.ExtractClient(ctx).Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []dmodel.DynamicFields
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(dmodel.DynamicFields)
		for i, col := range columns {
			val := convertDbValue(values[i])
			if val != nil {
				row[col] = val
			}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func convertDbValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	default:
		return v
	}
}
