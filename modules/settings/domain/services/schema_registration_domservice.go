package services

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	basemodel "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// registerSchema records one module's setting definitions for one level.
//
// It is idempotent by [module_key, level] so that it can run unconditionally on every boot: a
// module does not know whether it has been deployed before, and making registration conditional
// would put that question in every module rather than here.
func (this *SettingsDomainServiceImpl) RegisterSchema(
	ctx corectx.Context, cmd it.RegisterSchemaCommand,
) (*it.RegisterSchemaResult, error) {
	vErrs := ft.ClientErrors{}
	if cmd.ModuleKey == "" {
		vErrs.Append(*ft.NewValidationError("module_key", "settings.module_key_required",
			"module_key is required"))
	}
	if !isKnownLevel(cmd.Level) {
		vErrs.Append(*ft.NewValidationError("level", "settings.level_invalid",
			"level must be one of tenant, org or user"))
	}
	if cmd.Schema == nil {
		vErrs.Append(*ft.NewValidationError("schema", "settings.schema_required",
			"schema is required"))
	} else if len(cmd.Schema.PrimaryKeys()) > 0 {
		// A settings schema describes values, not a table. PrimaryKeys is populated only by
		// populateDbMetadata, which runs only under ShouldBuildDb, so a non-empty one is exactly
		// the "this schema wants a table" signal. The only tables this module builds are its own.
		vErrs.Append(*ft.NewBusinessViolation("schema", "settings.schema_must_not_build_db",
			"a settings schema must be metadata-only and must not declare should_build_db"))
	}
	if vErrs.Count() > 0 {
		return &it.RegisterSchemaResult{ClientErrors: vErrs}, nil
	}

	if cErr := this.assertNamesUniqueAcrossLevels(ctx, cmd); cErr != nil {
		vErrs.Append(*cErr)
		return &it.RegisterSchemaResult{ClientErrors: vErrs}, nil
	}

	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsSchemaSchemaName)
	if !ok {
		return nil, errors.Errorf("RegisterSchema: the '%s' engine is not registered",
			models.SettingsSchemaSchemaName)
	}

	document, err := simplizedDocument(cmd.Schema)
	if err != nil {
		return nil, errors.Wrap(err, "RegisterSchema")
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.SettingsSchemaFieldModuleKey: cmd.ModuleKey,
			models.SettingsSchemaFieldLevel:     cmd.Level,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "RegisterSchema")
	}

	if found.HasData {
		existing := models.NewSettingsSchemaFrom(found.Data)
		return this.updateRegisteredSchema(ctx, existing, document)
	}
	return this.insertRegisteredSchema(ctx, cmd, document)
}

func (this *SettingsDomainServiceImpl) insertRegisteredSchema(
	ctx corectx.Context, cmd it.RegisterSchemaCommand, document any,
) (*it.RegisterSchemaResult, error) {
	engine, _ := dynamicresource.Registry().GetEngine(models.SettingsSchemaSchemaName)

	id, err := model.NewId()
	if err != nil {
		return nil, errors.Wrap(err, "RegisterSchema")
	}

	data := dmodel.DynamicFields{
		models.SettingsSchemaFieldId:        *id,
		models.SettingsSchemaFieldModuleKey: cmd.ModuleKey,
		models.SettingsSchemaFieldLevel:     cmd.Level,
		models.SettingsSchemaFieldSchema:    document,
	}

	// created_at is declared UseTypeDefault on auditable_model, but that default is applied by the
	// engine's create pipeline, which this write bypasses by going straight to the repository. The
	// column is NOT NULL, so it has to be filled here — from the field's own declared default, so
	// the schema stays the single definition of what "now" means for this column.
	if err := setDeclaredDefault(engine, data, basemodel.FieldCreatedAt); err != nil {
		return nil, err
	}
	// The insert result is inspected rather than discarded: the repository reports a rejected
	// write as a ClientError, not a Go error, so ignoring it makes every failed registration
	// report Created: true and leaves the table silently empty.
	insRes, err := engine.ResourceRepository().Insert(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "RegisterSchema")
	}
	if insRes.ClientErrors.Count() > 0 {
		return &it.RegisterSchemaResult{ClientErrors: insRes.ClientErrors}, nil
	}

	return &it.RegisterSchemaResult{
		Data:    it.RegisterSchemaResultData{Id: *id, Created: true},
		HasData: true,
	}, nil
}

// updateRegisteredSchema rewrites the stored document when the module's declaration has changed.
// An identical document is left alone so that a boot which changed nothing writes nothing.
func (this *SettingsDomainServiceImpl) updateRegisteredSchema(
	ctx corectx.Context, existing *models.SettingsSchema, document any,
) (*it.RegisterSchemaResult, error) {
	id := existing.GetId()
	if id == nil {
		return nil, errors.New("updateRegisteredSchema: the stored schema row has no id")
	}

	if sameSchemaDocument(existing.GetSchema(), document) {
		return &it.RegisterSchemaResult{
			Data:    it.RegisterSchemaResultData{Id: *id, Created: false},
			HasData: true,
		}, nil
	}

	engine, _ := dynamicresource.Registry().GetEngine(models.SettingsSchemaSchemaName)
	data := dmodel.DynamicFields{
		models.SettingsSchemaFieldId:     *id,
		models.SettingsSchemaFieldSchema: document,
	}
	updRes, err := engine.ResourceRepository().Update(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "updateRegisteredSchema")
	}
	if updRes.ClientErrors.Count() > 0 {
		return &it.RegisterSchemaResult{ClientErrors: updRes.ClientErrors}, nil
	}

	return &it.RegisterSchemaResult{
		Data:    it.RegisterSchemaResultData{Id: *id, Created: false},
		HasData: true,
	}, nil
}

// assertNamesUniqueAcrossLevels rejects a schema declaring a setting name another level of the
// same module already declares. The unique key is [tenant, module, name, owner], with no level in
// it, so two levels sharing a name would collide on write for an owner holding both.
func (this *SettingsDomainServiceImpl) assertNamesUniqueAcrossLevels(
	ctx corectx.Context, cmd it.RegisterSchemaCommand,
) *ft.ClientErrorItem {
	incoming := fieldNameSet(cmd.Schema)

	for _, level := range []string{c.LevelTenant, c.LevelOrg, c.LevelUser} {
		if level == cmd.Level {
			continue
		}
		other, err := this.loadSchema(ctx, cmd.ModuleKey, level)
		if err != nil || other == nil {
			continue
		}
		for name := range fieldNameSetOfDocument(other.GetSchema()) {
			if _, clashes := incoming[name]; clashes {
				return ft.NewBusinessViolation("schema", "settings.name_used_at_another_level",
					"setting name '"+name+"' is already declared at the '"+level+"' level of this module")
			}
		}
	}
	return nil
}

// simplizedDocument renders a built schema as the JSON object the schema column stores.
//
// It emits MODEL JSON, not ToSimplized's output. The two are opposite directions and are not
// interchangeable: ToSimplized is the client-facing output shape (fields keyed by name, is_*
// booleans, "enumString"), while the parser reads the authoring shape (fields as an array,
// "required_for_create", "enum_string") and refuses properties it does not declare. The stored
// document is read back through that parser on every settings read, so it must be written in the
// shape the parser accepts.
//
// The round trip is pinned in schema_document_test.go: a document that stores but cannot be re-read
// breaks every read of that module's settings while registration still reports success, which is
// exactly how this went unnoticed once.
func simplizedDocument(schema *dmodel.ModelSchema) (map[string]any, error) {
	document, err := schema.ToModelJson()
	if err != nil {
		return nil, errors.Wrap(err, "simplizedDocument")
	}

	// Round-tripped through JSON so the stored value is a plain map of JSON-native types. The
	// column is jsonmap, whose type check rejects Go structs such as the LangJson label — and
	// reports the rejection as a ClientError rather than an error, which is easy to swallow.
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, errors.Wrap(err, "simplizedDocument")
	}

	var normalized map[string]any
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, errors.Wrap(err, "simplizedDocument")
	}
	return normalized, nil
}

// setDeclaredDefault fills one field from the default its schema declares.
//
// Registration writes through the repository rather than the engine's create pipeline, because it
// runs at start-up with no caller to authorize and no request to validate. The pipeline is what
// normally applies a field's declared default, so a direct write has to do it here — reading the
// default off the schema rather than restating it, so the two cannot drift.
func setDeclaredDefault(engine drif.DynamicResourceEngine, data dmodel.DynamicFields, fieldName string) error {
	field, ok := engine.Schema().Field(fieldName)
	if !ok {
		return errors.Errorf("setDeclaredDefault: '%s' declares no '%s' field",
			models.SettingsSchemaSchemaName, fieldName)
	}

	defaultValue := field.DataType().DefaultValue()
	if defaultValue.Get() == nil {
		return errors.Errorf("setDeclaredDefault: '%s' has no type default to apply", fieldName)
	}
	data[fieldName] = *defaultValue.Get()
	return nil
}
