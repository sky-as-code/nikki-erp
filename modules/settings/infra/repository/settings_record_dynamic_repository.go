package repository

import (
	"encoding/json"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

const tenantIdColumn = "tenant_id"

// NewSettingsRecordRepository is handed the composable default by the settings_record onion.
func NewSettingsRecordRepository(base composable.CrudRepository) itRecord.SettingsRecordRepository {
	return &SettingsRecordRepositoryImpl{CrudRepository: base}
}

// SettingsRecordRepositoryImpl adds the set-based statements of the write path.
//
// They go around the row-at-a-time CRUD deliberately: a tenant enforcing a setting onto every user
// of a large tenant would otherwise be one statement per user, and seeding a new owner runs inside
// the transaction that creates it, where a per-row loop would hold that transaction open across
// one round trip per setting. The multi-tenant decorator does not see these statements, so each
// scopes itself to the tenant explicitly.
type SettingsRecordRepositoryImpl struct {
	composable.CrudRepository
}

func (this *SettingsRecordRepositoryImpl) HasTenantColumn() bool {
	return this.Schema().TenantKey() != ""
}

// FanOutValue is idempotent, so a retried transaction is harmless. Children are matched by
// tenant_id alone: settings must never join iam tables (iam depends on settings).
func (this *SettingsRecordRepositoryImpl) FanOutValue(
	ctx corectx.Context, param itRecord.FanOutValueParam,
) (int, error) {
	envelope, err := json.Marshal(map[string]any{models.ValueEnvelopeKey: param.Value})
	if err != nil {
		return 0, errors.Wrap(err, "FanOutValue")
	}

	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(this.Schema().TableName())
	// The flag travels with the value: a child left claiming it may override, while its value has
	// just been forcibly overwritten, would render editable and invite a save the tenant's next
	// enforcement silently discards.
	builder.Set(
		builder.Assign(models.SettingsRecordFieldValue, string(envelope)),
		builder.Assign(models.SettingsRecordFieldAllowOverride, param.AllowOverride),
	)
	builder.Where(
		builder.Equal(models.SettingsRecordFieldModuleKey, param.ModuleKey),
		builder.Equal(models.SettingsRecordFieldName, param.Name),
		builder.Equal(models.SettingsRecordFieldOwnerType, param.ChildOwnerType),
	)
	if param.TenantId != "" {
		builder.Where(builder.Equal(tenantIdColumn, param.TenantId))
	} else if this.HasTenantColumn() {
		// Without the tenant predicate an enforced setting would cross into other tenants.
		return 0, errors.New("FanOutValue: the request carries no tenant id")
	}

	query, args := builder.Build()
	return this.execCount(ctx, query, args, "FanOutValue")
}

// CopyTenantRows copies with ON CONFLICT DO NOTHING, which makes a retried owner creation
// harmless and leaves a value the owner already holds untouched.
func (this *SettingsRecordRepositoryImpl) CopyTenantRows(
	ctx corectx.Context, param itRecord.CopyTenantRowsParam,
) (int, error) {
	query, args, err := buildCopyTenantRowsSql(this.Schema().TableName(), this.HasTenantColumn(), param)
	if err != nil {
		return 0, err
	}
	return this.execCount(ctx, query, args, "CopyTenantRows")
}

func buildCopyTenantRowsSql(
	table string, hasTenantColumn bool, param itRecord.CopyTenantRowsParam,
) (string, []any, error) {
	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"gen_random_uuid()::text",
		models.SettingsRecordFieldSchemaId,
		models.SettingsRecordFieldModuleKey,
		selectBuilder.Var(param.Level),
		selectBuilder.Var(param.OwnerType),
		selectBuilder.Var(param.OwnerId),
		models.SettingsRecordFieldName,
		models.SettingsRecordFieldValue,
		// Carried forward so an owner created after an enforcement does not escape it.
		models.SettingsRecordFieldAllowOverride,
		"now()",
	)
	selectBuilder.From(table)
	selectBuilder.Where(selectBuilder.Equal(models.SettingsRecordFieldLevel, c.LevelTenant))

	// Positional against the select list above: the two must stay in the same order.
	columns := []string{
		models.SettingsRecordFieldId, models.SettingsRecordFieldSchemaId,
		models.SettingsRecordFieldModuleKey, models.SettingsRecordFieldLevel,
		models.SettingsRecordFieldOwnerType, models.SettingsRecordFieldOwnerId,
		models.SettingsRecordFieldName, models.SettingsRecordFieldValue,
		models.SettingsRecordFieldAllowOverride, models.SettingsRecordFieldCreatedAt,
	}
	if param.TenantId != "" {
		selectBuilder.Where(selectBuilder.Equal(tenantIdColumn, param.TenantId))
		// These rows are written outside the CRUD path, so nothing else would stamp the tenant.
		selectBuilder.SelectMore(selectBuilder.Var(param.TenantId))
		columns = append(columns, tenantIdColumn)
	} else if hasTenantColumn {
		return "", nil, errors.New("CopyTenantRows: the request carries no tenant id")
	}

	query, args := sqlbuilder.Buildf("INSERT INTO %v (%v) %v ON CONFLICT DO NOTHING",
		sqlbuilder.Raw(table), sqlbuilder.Raw(strings.Join(columns, ", ")), selectBuilder,
	).BuildWithFlavor(sqlbuilder.PostgreSQL)
	return query, args, nil
}

// execCount runs a statement on the request's client, which is the open transaction when the
// context carries one, and returns the rows it touched.
func (this *SettingsRecordRepositoryImpl) execCount(
	ctx corectx.Context, query string, args []any, action string,
) (int, error) {
	client := this.GetBaseRepo().ExtractClient(ctx)
	result, err := client.Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, action)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		// A driver that cannot report the count still applied the statement.
		return 0, nil
	}
	return int(affected), nil
}
