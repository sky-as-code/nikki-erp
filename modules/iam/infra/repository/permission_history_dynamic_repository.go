package repository

import (
	"fmt"
	"strings"

	"go.uber.org/dig"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

type PermissionHistoryRepositoryParam struct {
	dig.In

	Client        orm.DbClient
	ConfigSvc     config.ConfigService
	QueryBuilder  orm.QueryBuilder
	Logger        logging.LoggerService
	NewBaseRepoFn dyn.NewBaseDynamicRepositoryFn
}

func NewPermissionHistoryDynamicRepository(param PermissionHistoryRepositoryParam) itPerm.PermissionHistoryRepository {
	dynamicRepo := param.NewBaseRepoFn(
		dyn.NewBaseRepoParam{
			Client:       param.Client,
			ConfigSvc:    param.ConfigSvc,
			QueryBuilder: param.QueryBuilder,
			Logger:       param.Logger,
			Schema:       dmodel.MustGetSchema(models.PermissionHistorySchemaName),
		},
	)
	return &PermissionHistoryDynamicRepository{dynamicRepo: dynamicRepo}
}

type PermissionHistoryDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *PermissionHistoryDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *PermissionHistoryDynamicRepository) BeginTransaction(ctx corectx.Context) (database.DbTransaction, error) {
	return this.dynamicRepo.BeginTransaction(ctx)
}

// Insert appends one audit row.
//
// Written as raw SQL rather than through the generic insert path: `created_at` is
// declared auto-generated, so the schema mapper blanks it for its own generator to
// fill, which never runs on a direct insert and leaves the NOT NULL column null.
// An append-only table with no update path has little to gain from the mapper
// anyway, and this keeps the audit write independent of it.
func (this *PermissionHistoryDynamicRepository) Insert(
	ctx corectx.Context, entry models.PermissionHistory,
) (*dyn.OpResult[int], error) {
	fields := entry.GetFieldData()

	columns := []string{
		basemodel.FieldId, models.PermHistoryFieldEffect, models.PermHistoryFieldReason,
		models.PermHistoryFieldRoleId, models.PermHistoryFieldRoleName,
		models.PermHistoryFieldReceiverId, models.PermHistoryFieldApproverId,
		models.PermHistoryFieldEntitlementId, models.PermHistoryFieldEntitlementExpr,
		models.PermHistoryFieldAssignmentId,
	}
	placeholders := make([]string, 0, len(columns)+2)
	values := make([]any, 0, len(columns)+2)
	for index, column := range columns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		values = append(values, normalizeAuditValue(fields[column]))
	}

	// created_at is stamped by the database so the trail cannot carry a clock the
	// caller chose. Tenant scoping is added only where the column exists, which is
	// what keeps this one file working across both the single- and multi-tenant schemas.
	columnList := strings.Join(columns, ", ") + ", created_at"
	valueList := strings.Join(placeholders, ", ") + ", NOW()"
	if tenantId := tenantIdOf(ctx); tenantId != "" {
		columnList += ", tenant_id"
		values = append(values, tenantId)
		valueList += fmt.Sprintf(", $%d", len(values))
	}

	query := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		dmodel.MustGetSchema(models.PermissionHistorySchemaName).TableName(), columnList, valueList,
	)
	result, err := this.dynamicRepo.ExtractClient(ctx).Exec(ctx.InnerContext(), query, values...)
	if err != nil {
		return nil, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dyn.OpResult[int]{Data: int(rows), HasData: rows != 0}, nil
}

// normalizeAuditValue turns the model's typed pointers into something the driver
// accepts, and an absent value into a real NULL rather than a typed nil.
func normalizeAuditValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case *string:
		if typed == nil {
			return nil
		}
		return *typed
	case model.Id:
		return string(typed)
	}
	return value
}

func tenantIdOf(ctx corectx.Context) string {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil {
		return ""
	}
	if raw, ok := constraints["tenant_id"]; ok {
		return fmt.Sprintf("%v", raw)
	}
	return ""
}

func (this *PermissionHistoryDynamicRepository) Search(
	ctx corectx.Context, param dyn.RepoSearchParam,
) (*dyn.OpResult[dyn.PagedResultData[models.PermissionHistory]], error) {
	return baserepo.Search[models.PermissionHistory](ctx, this.dynamicRepo, param)
}
