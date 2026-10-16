package settings

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// settingsRecordResource is the Settings module's settings_record schema name, the resource its
// application services assert on. It is spelled out because core cannot import that module.
const settingsRecordResource = "settings_record"

// UpdateSettingsCommand is what a module's own update endpoint receives. The module key is not in
// it: the endpoint belongs to one module, which names itself.
type UpdateSettingsCommand struct {
	Level string
	Items []SetSettingItem
}

// SettingsRule is a module's own check on a write, beyond what its schema can say -- typically one
// that reads the module's data, such as "this language is enabled". It runs only after the items
// passed the schema, so it may assume every name is declared and every value well-typed.
type SettingsRule func(ctx corectx.Context, level string, items []SetSettingItem) (ft.ClientErrors, error)

// ModuleSettingsWriter is the shared body of every module's settings update: authorize, refuse a
// level the module declares nothing at, validate the items against the module's schema, run the
// module's rules, then persist through SettingsService.
type ModuleSettingsWriter struct {
	moduleKey string
	schemas   map[string]*dmodel.ModelSchema
	rules     []SettingsRule
	service   SettingsService
}

// NewModuleSettingsWriter takes the schemas by level, the same ones the module registers with the
// Settings module, so the two validations can never disagree on what a setting is.
func NewModuleSettingsWriter(
	moduleKey string, schemas map[string]*dmodel.ModelSchema, service SettingsService, rules ...SettingsRule,
) *ModuleSettingsWriter {
	return &ModuleSettingsWriter{moduleKey: moduleKey, schemas: schemas, rules: rules, service: service}
}

func (this *ModuleSettingsWriter) Update(ctx corectx.Context, cmd UpdateSettingsCommand) (*SetSettingsResult, error) {
	if cErr := assertCanWrite(ctx, cmd.Level); cErr != nil {
		return clientErrorsResult(*cErr), nil
	}
	schema, ok := this.schemas[cmd.Level]
	if !ok || !IsKnownLevel(cmd.Level) {
		return clientErrorsResult(ft.ClientErrors{*ft.NewValidationError("level", "settings.schema_not_registered",
			"this module has registered no settings at this level")}), nil
	}
	if vErrs := ValidateItems(schema, cmd.Items); vErrs.Count() > 0 {
		return clientErrorsResult(vErrs), nil
	}
	for _, rule := range this.rules {
		vErrs, err := rule(ctx, cmd.Level, cmd.Items)
		if err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return clientErrorsResult(vErrs), nil
		}
	}
	return this.service.SetSettings(ctx, cmd.Level, SetSettingsCommand{ModuleKey: this.moduleKey, Items: cmd.Items})
}

func clientErrorsResult(vErrs ft.ClientErrors) *SetSettingsResult {
	return &SetSettingsResult{ClientErrors: vErrs}
}

// assertCanWrite runs the Settings module's own check before anything else, so a caller without
// the right learns that first rather than whether the module declares the level. A tenant or org
// write needs update on settings_record; a user write is the actor's own row. The Settings
// application service asserts again when the port reaches it; the repeat is a cache lookup.
func assertCanWrite(ctx corectx.Context, level string) *ft.ClientErrors {
	if level != LevelTenant && level != LevelOrg {
		return nil
	}
	return reguard.AssertPermission(ctx, reguard.Perm{
		ActionCode: "update", ResourceCode: settingsRecordResource, Scope: reguard.ResourceScopeTenant,
	})
}
