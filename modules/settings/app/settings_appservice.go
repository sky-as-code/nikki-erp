package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// The three services below are the module's public surface, one per level.
//
// They are separate types rather than one service taking a level argument on purpose: a consumer
// holding the organization contract must not be able to reach tenant or user rows, and a level
// parameter would let it do exactly that by passing a different string. Each binds its level once,
// here, and the compiler keeps a consumer inside it.
//
// All three delegate to the same domain service. The split is about the shape of the contract, not
// about three storage paths.

// --- Tenant ---

func NewTenantSettingsAppServiceImpl(settingsSvc it.SettingsDomainService) it.TenantSettingsAppService {
	return &TenantSettingsAppServiceImpl{settingsSvc: settingsSvc}
}

type TenantSettingsAppServiceImpl struct {
	settingsSvc it.SettingsDomainService
}

// RegisterSchema is reached during start-up rather than through a request, so it asserts no
// permission: there is no acting user to assert against, and a module registering its own
// declaration is not an operation an end user performs.
func (this *TenantSettingsAppServiceImpl) RegisterSchema(
	ctx corectx.Context, cmd it.RegisterSchemaCommand,
) (*it.RegisterSchemaResult, error) {
	return this.settingsSvc.RegisterSchema(ctx, cmd)
}

func (this *TenantSettingsAppServiceImpl) GetTenantSettings(
	ctx corectx.Context, query it.GetSettingsQuery,
) (*it.GetSettingsResult, error) {
	if cErr := assertPermission(ctx, actionRead, settingsRecordResource); cErr != nil {
		return &it.GetSettingsResult{ClientErrors: *cErr}, nil
	}
	return this.settingsSvc.GetSettings(ctx, c.LevelTenant, c.OwnerTypeTenant, query)
}

func (this *TenantSettingsAppServiceImpl) SetTenantSettings(
	ctx corectx.Context, cmd it.SetSettingsCommand,
) (*it.SetSettingsResult, error) {
	if cErr := assertPermission(ctx, actionUpdate, settingsRecordResource); cErr != nil {
		return &it.SetSettingsResult{ClientErrors: *cErr}, nil
	}
	return this.settingsSvc.SetSettings(ctx, c.LevelTenant, c.OwnerTypeTenant, cmd)
}

// --- Organization ---

func NewOrgSettingsAppServiceImpl(settingsSvc it.SettingsDomainService) it.OrgSettingsAppService {
	return &OrgSettingsAppServiceImpl{settingsSvc: settingsSvc}
}

type OrgSettingsAppServiceImpl struct {
	settingsSvc it.SettingsDomainService
}

func (this *OrgSettingsAppServiceImpl) GetOrgSettings(
	ctx corectx.Context, query it.GetSettingsQuery,
) (*it.GetSettingsResult, error) {
	if cErr := assertPermission(ctx, actionRead, settingsRecordResource); cErr != nil {
		return &it.GetSettingsResult{ClientErrors: *cErr}, nil
	}
	return this.settingsSvc.GetSettings(ctx, c.LevelOrg, c.OwnerTypeOrg, query)
}

func (this *OrgSettingsAppServiceImpl) SetOrgSettings(
	ctx corectx.Context, cmd it.SetSettingsCommand,
) (*it.SetSettingsResult, error) {
	if cErr := assertPermission(ctx, actionUpdate, settingsRecordResource); cErr != nil {
		return &it.SetSettingsResult{ClientErrors: *cErr}, nil
	}
	return this.settingsSvc.SetSettings(ctx, c.LevelOrg, c.OwnerTypeOrg, cmd)
}

// InitOrgSettings runs inside iam's organization-creating transaction, before the organization is
// visible to anyone. It asserts no permission of its own: the caller already proved it may create
// the organization, and there is no separate right to seed the rows that come with one.
func (this *OrgSettingsAppServiceImpl) InitOrgSettings(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand,
) (*it.InitOwnerSettingsResult, error) {
	cmd.OwnerType = c.OwnerTypeOrg
	return this.settingsSvc.InitOwnerSettings(ctx, cmd)
}

// --- User ---

func NewUserPreferencesAppServiceImpl(settingsSvc it.SettingsDomainService) it.UserPreferencesAppService {
	return &UserPreferencesAppServiceImpl{settingsSvc: settingsSvc}
}

type UserPreferencesAppServiceImpl struct {
	settingsSvc it.SettingsDomainService
}

// GetUserPreferences asserts no permission: every account may read its own preferences, and the
// owner is the request's own user id rather than anything the caller names, so there is nothing
// here one user could use to reach another's row.
func (this *UserPreferencesAppServiceImpl) GetUserPreferences(
	ctx corectx.Context, query it.GetSettingsQuery,
) (*it.GetSettingsResult, error) {
	return this.settingsSvc.GetSettings(ctx, c.LevelUser, c.OwnerTypeUser, query)
}

// SetUserPreferences likewise writes only the acting user's own rows, and the domain service
// refuses any item the tenant has locked.
func (this *UserPreferencesAppServiceImpl) SetUserPreferences(
	ctx corectx.Context, cmd it.SetSettingsCommand,
) (*it.SetSettingsResult, error) {
	return this.settingsSvc.SetSettings(ctx, c.LevelUser, c.OwnerTypeUser, cmd)
}

// InitUserPreferences runs inside iam's user-creating transaction, on the same reasoning as
// InitOrgSettings.
func (this *UserPreferencesAppServiceImpl) InitUserPreferences(
	ctx corectx.Context, cmd it.InitOwnerSettingsCommand,
) (*it.InitOwnerSettingsResult, error) {
	cmd.OwnerType = c.OwnerTypeUser
	return this.settingsSvc.InitOwnerSettings(ctx, cmd)
}

// --- Effective (all levels) ---

func NewEffectiveSettingsAppServiceImpl(
	settingsSvc it.SettingsDomainService,
) it.EffectiveSettingsAppService {
	return &EffectiveSettingsAppServiceImpl{settingsSvc: settingsSvc}
}

type EffectiveSettingsAppServiceImpl struct {
	settingsSvc it.SettingsDomainService
}

// GetEffectiveSettings asserts no permission, on the same reasoning as GetUserPreferences: the
// owner of every row it reads is taken from the request context, so a caller can only ever see the
// values that already apply to them.
func (this *EffectiveSettingsAppServiceImpl) GetEffectiveSettings(
	ctx corectx.Context, query it.GetEffectiveSettingsQuery,
) (*it.GetEffectiveSettingsResult, error) {
	return this.settingsSvc.GetEffectiveSettings(ctx, query)
}
