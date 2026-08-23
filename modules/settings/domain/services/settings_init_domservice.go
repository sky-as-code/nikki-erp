package services

import (
	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// InitOwnerSettings seeds a newly created organization or user with its own settings rows.
//
// The values are copied from the tenant's own level = 'tenant' rows, with level, owner_type and
// owner_id changed to suit the new owner. So a new owner starts from what the tenant has actually
// configured rather than from the schema defaults, which is what makes a tenant-wide choice apply
// to people who join afterwards.
//
// It is called by iam from inside the transaction that creates the organization or user, through a
// port iam owns. Settings must never import iam: iam already depends on settings, so the reverse
// edge would abort start-up. That is also why the copy reads and writes only settings_records,
// filtered by tenant, and never joins an iam table to find out who the new owner is.
//
// KNOWN DEBT: like settings_fanout_domservice, this builds SQL inside a domain service. Query
// building belongs in infra/repository/ — see docs/wiki/07. ERP backend module.md §5 — and this
// module has no infra/repository/ yet. Do not copy this pattern into a new service.
func (this *SettingsDomainServiceImpl) InitOwnerSettings(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand,
) (*it.InitOwnerSettingsResult, error) {
	level, err := levelForOwnerType(cmd.OwnerType)
	if err != nil {
		return nil, err
	}
	if cmd.OwnerId == "" {
		vErrs := ft.ClientErrors{}
		vErrs.Append(*ft.NewValidationError("owner_id", "settings.owner_id_required",
			"owner_id is required"))
		return &it.InitOwnerSettingsResult{ClientErrors: vErrs}, nil
	}

	created, err := this.copyTenantRows(ctx, cmd, level)
	if err != nil {
		return nil, err
	}

	return &it.InitOwnerSettingsResult{
		Data:    it.InitOwnerSettingsResultData{Created: created},
		HasData: true,
	}, nil
}

// levelForOwnerType maps a new owner to the level whose settings it should receive. A tenant is not
// initialized from anything: it is the source the others are copied from.
func levelForOwnerType(ownerType string) (string, error) {
	switch ownerType {
	case c.OwnerTypeOrg:
		return c.LevelOrg, nil
	case c.OwnerTypeUser:
		return c.LevelUser, nil
	}
	return "", errors.Errorf("levelForOwnerType: '%s' cannot be initialized", ownerType)
}

// copyTenantRows performs the copy as a single INSERT ... SELECT.
//
// Set-based rather than read-then-loop because it runs inside the transaction that creates the
// owner: a per-row loop would hold that transaction open across one round trip per setting, and a
// tenant may have many.
//
// The ON CONFLICT DO NOTHING clause is what makes it idempotent, so a retried creation cannot
// violate the unique key. It also means an owner that already holds a value keeps it, which is the
// right outcome for a re-run.
func (this *SettingsDomainServiceImpl) copyTenantRows(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand, level string,
) (int, error) {
	tenantId := string(cmd.TenantId)
	if tenantId == "" {
		tenantId = actingTenantId(ctx)
	}

	selectBuilder := sqlbuilder.PostgreSQL.NewSelectBuilder()
	selectBuilder.Select(
		"gen_random_uuid()::text",
		models.SettingsRecordFieldSchemaId,
		models.SettingsRecordFieldModuleKey,
		selectBuilder.Var(level),
		selectBuilder.Var(cmd.OwnerType),
		selectBuilder.Var(string(cmd.OwnerId)),
		models.SettingsRecordFieldName,
		models.SettingsRecordFieldValue,
		// Carried forward so an owner created after an enforcement does not escape it: without
		// this the new row's flag is NULL, which reads as overridable, and the org the tenant
		// just locked out is replaced by one that is not.
		models.SettingsRecordFieldAllowOverride,
		"now()",
	)
	selectBuilder.From(settingsRecordsTable)
	selectBuilder.Where(
		selectBuilder.Equal(models.SettingsRecordFieldLevel, c.LevelTenant),
	)

	columns := []string{
		models.SettingsRecordFieldId,
		models.SettingsRecordFieldSchemaId,
		models.SettingsRecordFieldModuleKey,
		models.SettingsRecordFieldLevel,
		models.SettingsRecordFieldOwnerType,
		models.SettingsRecordFieldOwnerId,
		models.SettingsRecordFieldName,
		models.SettingsRecordFieldValue,
		// Positional against the select list above -- the two must stay in the same order.
		models.SettingsRecordFieldAllowOverride,
		models.SettingsRecordFieldCreatedAt,
	}

	if tenantId != "" {
		selectBuilder.Where(selectBuilder.Equal(tenantIdField, tenantId))
		// The copy carries the tenant forward explicitly: these rows are written outside the
		// engine, so nothing else would stamp it.
		selectBuilder.SelectMore(selectBuilder.Var(tenantId))
		columns = append(columns, tenantIdField)
	} else if hasTenantColumn() {
		return 0, errors.New("copyTenantRows: the request carries no tenant id")
	}

	selectQuery, args := selectBuilder.Build()
	query := "INSERT INTO " + settingsRecordsTable + " (" + joinColumns(columns) + ") " +
		selectQuery + " ON CONFLICT DO NOTHING"

	return execCount(ctx, query, args, "copyTenantRows")
}

func joinColumns(columns []string) string {
	joined := ""
	for i, column := range columns {
		if i > 0 {
			joined += ", "
		}
		joined += column
	}
	return joined
}
