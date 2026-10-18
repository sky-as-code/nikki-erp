package services

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
	itRecord "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsrecord"
	itSchema "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settingsschema"
)

func NewSettingsDomainServiceImpl(
	schemaRepo itSchema.SettingsSchemaRepository, recordRepo itRecord.SettingsRecordRepository,
) it.SettingsDomainService {
	return &SettingsDomainServiceImpl{schemaRepo: schemaRepo, recordRepo: recordRepo}
}

// SettingsDomainServiceImpl reads and writes through the repositories of the two settings onions:
// they already own the schema, the query builder and the tenant filter, and a second path to the
// same two tables would have to keep all three in step.
type SettingsDomainServiceImpl struct {
	schemaRepo itSchema.SettingsSchemaRepository
	recordRepo itRecord.SettingsRecordRepository
}

func isKnownLevel(level string) bool {
	switch level {
	case c.LevelTenant, c.LevelOrg, c.LevelUser:
		return true
	}
	return false
}

// loadSchema returns the registered schema row for a module at a level, or nil when none exists.
// A module that registered nothing for a level contributes no section to the settings page, so
// absence is an ordinary outcome rather than an error.
func (this *SettingsDomainServiceImpl) loadSchema(
	ctx corectx.Context, moduleKey string, level string,
) (*models.SettingsSchema, error) {
	found, err := this.schemaRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.SettingsSchemaFieldModuleKey: moduleKey,
			models.SettingsSchemaFieldLevel:     level,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadSchema")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewSettingsSchemaFrom(found.Data), nil
}

// fieldNameSet lists the setting names a built schema declares.
func fieldNameSet(schema *dmodel.ModelSchema) map[string]struct{} {
	names := map[string]struct{}{}
	if schema == nil {
		return names
	}
	for _, name := range schema.FieldNames() {
		names[name] = struct{}{}
	}
	return names
}

// fieldNameSetOfDocument lists the setting names a *stored* schema document declares. The stored
// form is model JSON, so its field names live in the "fields" array.
func fieldNameSetOfDocument(document map[string]any) map[string]struct{} {
	names := map[string]struct{}{}
	if document == nil {
		return names
	}
	// The stored document is model JSON, which lists fields as an array of objects each carrying
	// its own name. Reading the wrong shape here returns an empty set rather than failing, which
	// would make the cross-level uniqueness check pass vacuously and let two levels of one module
	// declare the same setting name — a collision on write, because the record unique key carries
	// no level.
	rawFields, ok := document["fields"].([]any)
	if !ok {
		return names
	}
	for _, rawField := range rawFields {
		field, ok := rawField.(map[string]any)
		if !ok {
			continue
		}
		if name, ok := field["name"].(string); ok {
			names[name] = struct{}{}
		}
	}
	return names
}

// sameSchemaDocument reports whether a stored document already equals the one being registered.
// Both sides are compared as encoded JSON: the stored side came back from jsonb with every number
// widened to float64 and every map key reordered, so a structural comparison of the Go values
// would report a difference on every boot and rewrite a row that did not change.
func sameSchemaDocument(stored map[string]any, incoming any) bool {
	if stored == nil {
		return false
	}
	storedJson, err := json.Marshal(stored)
	if err != nil {
		return false
	}
	incomingJson, err := json.Marshal(incoming)
	if err != nil {
		return false
	}

	var storedNormalized, incomingNormalized any
	if json.Unmarshal(storedJson, &storedNormalized) != nil {
		return false
	}
	if json.Unmarshal(incomingJson, &incomingNormalized) != nil {
		return false
	}

	renderedStored, err := json.Marshal(storedNormalized)
	if err != nil {
		return false
	}
	renderedIncoming, err := json.Marshal(incomingNormalized)
	if err != nil {
		return false
	}
	return string(renderedStored) == string(renderedIncoming)
}

// allowOverrideOf reads the per-setting override policy from the field's metadata. A setting that
// declares nothing is overridable: the restrictive reading would silently lock every setting a
// module forgot to annotate.
func allowOverrideOf(field *dmodel.ModelField) bool {
	if field == nil {
		return true
	}
	raw, ok := field.MetadataValue(c.MetadataKeyAllowOverride)
	if !ok {
		return true
	}
	allow, ok := raw.(bool)
	if !ok {
		return true
	}
	return allow
}

// schemaFromDocument rebuilds a model schema from its stored document.
//
// The stored form is model JSON — what ModelSchema.ToModelJson produced — so the round trip goes
// back through the one parser rather than a second reader that would have to be kept in step with
// it. Note this is NOT ToSimplized's output: that is the client-facing shape and the parser
// rejects it.
func schemaFromDocument(document map[string]any) (*dmodel.ModelSchema, error) {
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, errors.Wrap(err, "schemaFromDocument")
	}

	builder, cErrs := dmodel.ParseModelJsonSafe(string(encoded))
	if cErrs.Count() > 0 {
		return nil, errors.Wrap(cErrs.ToError(), "schemaFromDocument")
	}
	return builder.Build(), nil
}
