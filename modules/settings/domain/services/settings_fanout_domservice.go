package services

import (
	"encoding/json"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
)

// settingsRecordsTable is the physical table the set-based statements below address.
//
// They go around the resource engine deliberately: the engine writes one row at a time, and a
// tenant enforcing a setting onto every user of a large tenant would otherwise be one statement per
// user. These are the only two places in the module that write SQL directly.
//
// KNOWN DEBT: query building belongs in infra/repository/, not in a domain service — see
// docs/wiki/07. ERP backend module.md §5. This module has no infra/repository/ yet. The set-based
// rationale above stands; the placement does not. Do not copy this pattern into a new service.
const settingsRecordsTable = "settings_records"

// fanOutToChildren writes an enforced tenant value onto every child record in the tenant.
//
// This is what "allow_override = false" means in this design: enforcement is a physical write, not
// a read-time precedence rule, so a child's stored value simply becomes the tenant's. It overwrites
// whatever the child had chosen, with no record of the previous value — that is intended, and it is
// why the UI warns before saving an enforcement.
//
// Children are matched by tenant_id alone. Settings must never join iam tables: iam already depends
// on settings, so reaching back into iam_org or iam_user would close a cycle that aborts start-up.
//
// It is a single set-based UPDATE rather than a per-row loop, because a tenant may hold very many
// users, and it is idempotent, so a retried transaction is harmless.
func (this *SettingsDomainServiceImpl) fanOutToChildren(
	ctx corectx.Context, param applyParam, item validatedItem,
) (int, error) {
	childOwnerType, err := childOwnerTypeFor(param.Level)
	if err != nil {
		return 0, err
	}

	envelope, err := json.Marshal(map[string]any{models.ValueEnvelopeKey: item.Value})
	if err != nil {
		return 0, errors.Wrap(err, "fanOutToChildren")
	}

	builder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	builder.Update(settingsRecordsTable)
	// The flag travels with the value (D19). A child left claiming it may override, while its
	// value has just been forcibly overwritten, would render editable and invite a save the
	// tenant's next enforcement silently discards -- and `isEditable` reads the child's own row,
	// so the row has to tell the truth about itself.
	builder.Set(
		builder.Assign(models.SettingsRecordFieldValue, string(envelope)),
		builder.Assign(models.SettingsRecordFieldAllowOverride, item.AllowOverride),
	)
	builder.Where(
		builder.Equal(models.SettingsRecordFieldModuleKey, param.ModuleKey),
		builder.Equal(models.SettingsRecordFieldName, item.Name),
		builder.Equal(models.SettingsRecordFieldOwnerType, childOwnerType),
	)
	if err := scopeToTenant(ctx, builder); err != nil {
		return 0, err
	}

	query, args := builder.Build()
	return execCount(ctx, query, args, "fanOutToChildren")
}

// childOwnerTypeFor maps the level being enforced onto the kind of owner that receives the value.
// A tenant-level setting has no children: there is no layer above the tenant to enforce it, and no
// layer below that holds a tenant-level value.
func childOwnerTypeFor(level string) (string, error) {
	switch level {
	case c.LevelOrg:
		return c.OwnerTypeOrg, nil
	case c.LevelUser:
		return c.OwnerTypeUser, nil
	}
	return "", errors.Errorf("childOwnerTypeFor: level '%s' has no child records", level)
}

// scopeToTenant adds the tenant predicate to a set-based statement.
//
// The multi-tenant repository decorator adds this automatically to everything that goes through the
// engine, but these statements do not, so the scope has to be explicit. Without it an enforced
// setting would cross into other tenants, which is why a missing tenant id fails the statement
// rather than widening it.
func scopeToTenant(ctx corectx.Context, builder *sqlbuilder.UpdateBuilder) error {
	tenantId := actingTenantId(ctx)
	if tenantId != "" {
		builder.Where(builder.Equal(tenantIdField, tenantId))
		return nil
	}
	// The nikkierp binary carries no tenant key, so its settings_records has no tenant_id column
	// and every row already belongs to the single implicit tenant.
	if hasTenantColumn() {
		return errors.New("scopeToTenant: the request carries no tenant id")
	}
	return nil
}

// hasTenantColumn reports whether this binary's settings_records carries a tenant column. It is
// true in coremart, where apptrait injects one, and false in the nikkierp binary.
func hasTenantColumn() bool {
	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsRecordSchemaName)
	if !ok {
		return false
	}
	return engine.Schema().TenantKey() != ""
}

// execCount runs a set-based statement on the request's client and returns the rows it touched.
// ExtractClient hands back the open transaction when the context carries one, so these statements
// join the same transaction as the engine writes around them rather than committing separately.
func execCount(ctx corectx.Context, query string, args []any, action string) (int, error) {
	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsRecordSchemaName)
	if !ok {
		return 0, errors.Errorf("%s: the '%s' engine is not registered",
			action, models.SettingsRecordSchemaName)
	}

	client := engine.ResourceRepository().GetBaseRepo().ExtractClient(ctx)
	result, err := client.Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, action)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		// A driver that cannot report the count still applied the statement, so this is not a
		// failure of the write itself.
		return 0, nil
	}
	return int(affected), nil
}
