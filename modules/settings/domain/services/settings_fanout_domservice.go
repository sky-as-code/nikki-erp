package services

import (
	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
)

// fanOutToChildren writes an enforced tenant value onto every child record in the tenant.
//
// This is what "allow_override = false" means in this design: enforcement is a physical write, not
// a read-time precedence rule, so a child's stored value simply becomes the tenant's. It overwrites
// whatever the child had chosen, with no record of the previous value — that is intended, and it is
// why the UI warns before saving an enforcement.
//
// The statement itself is the repository's (one set-based UPDATE scoped to the acting tenant).
func (this *SettingsDomainServiceImpl) fanOutToChildren(
	ctx corectx.Context, param applyParam, item validatedItem,
) (int, error) {
	childOwnerType, err := childOwnerTypeFor(param.Level)
	if err != nil {
		return 0, err
	}
	return this.recordRepo.FanOutValue(ctx, itRecord.FanOutValueParam{
		ModuleKey:      param.ModuleKey,
		Name:           item.Name,
		ChildOwnerType: childOwnerType,
		Value:          item.Value,
		AllowOverride:  item.AllowOverride,
		TenantId:       actingTenantId(ctx),
	})
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
