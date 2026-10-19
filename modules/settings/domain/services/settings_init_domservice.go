package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
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
func (this *SettingsDomainServiceImpl) InitOwnerSettings(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand,
) (*it.InitOwnerSettingsResult, error) {
	level, err := levelForOwnerType(cmd.OwnerType)
	if err != nil {
		return nil, err
	}
	if cmd.OwnerId == "" {
		vErrs := ft.ClientErrors{}
		vErrs.Append(*ft.NewValidationError("owner_id", ft.ErrorKey("settings.owner_id_required", "settings"),
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

// copyTenantRows seeds the owner through the repository's single INSERT ... SELECT.
//
// Set-based rather than read-then-loop because it runs inside the transaction that creates the
// owner, and idempotent (ON CONFLICT DO NOTHING), so a retried creation cannot violate the unique
// key and an owner that already holds a value keeps it.
func (this *SettingsDomainServiceImpl) copyTenantRows(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand, level string,
) (int, error) {
	tenantId := string(cmd.TenantId)
	if tenantId == "" {
		tenantId = actingTenantId(ctx)
	}
	return this.recordRepo.CopyTenantRows(ctx, itRecord.CopyTenantRowsParam{
		Level:     level,
		OwnerType: cmd.OwnerType,
		OwnerId:   string(cmd.OwnerId),
		TenantId:  tenantId,
	})
}

func (this *SettingsDomainServiceImpl) DeleteOwnerSettings(
	ctx corectx.Context, ownerType string, ownerId model.Id,
) (int, error) {
	return this.recordRepo.DeleteOwnerRows(ctx, itRecord.DeleteOwnerRowsParam{
		OwnerType: ownerType,
		OwnerId:   string(ownerId),
		TenantId:  actingTenantId(ctx),
	})
}
