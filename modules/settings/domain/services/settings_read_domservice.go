package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// GetSettings returns the items of one module at one level for one owner.
//
// Reads are plain reads. There is no precedence resolution here: when a tenant admin enforces a
// setting, the value is physically written onto every child row, so what this returns is already
// the effective value.
func (this *SettingsDomainServiceImpl) GetSettings(
	ctx corectx.Context, level string, ownerType string, query it.GetSettingsQuery,
) (*it.GetSettingsResult, error) {
	if !isKnownLevel(level) {
		return nil, errors.Errorf("GetSettings: unknown level '%s'", level)
	}

	ownerId, err := ownerIdFor(ctx, ownerType)
	if err != nil {
		return nil, err
	}

	schemaRow, err := this.loadSchema(ctx, query.ModuleKey, level)
	if err != nil {
		return nil, err
	}
	// A module that registered no schema for this level contributes no section, which is an
	// ordinary outcome rather than a missing resource.
	if schemaRow == nil {
		return &it.GetSettingsResult{
			Data:    it.GetSettingsResultData{ModuleKey: query.ModuleKey, OwnerType: ownerType, Items: []it.SettingItem{}},
			HasData: true,
		}, nil
	}

	declared, err := parseStoredSchema(schemaRow)
	if err != nil {
		return nil, err
	}

	stored, err := this.loadRecords(ctx, query.ModuleKey, level, ownerType, ownerId)
	if err != nil {
		return nil, err
	}

	items := make([]it.SettingItem, 0, len(declared.FieldNames()))
	for _, name := range declared.FieldNames() {
		field, ok := declared.Field(name)
		if !ok {
			continue
		}
		items = append(items, buildItem(field, name, level, ownerType, stored[name]))
	}

	return &it.GetSettingsResult{
		Data:    it.GetSettingsResultData{ModuleKey: query.ModuleKey, OwnerType: ownerType, Items: items},
		HasData: true,
	}, nil
}

// buildItem merges one declared field with its stored row, if any.
//
// A name with no row still renders, using the schema's declared default: rows are seeded when an
// owner is created, so a setting added to a schema after that has no row anywhere until someone
// saves it.
func buildItem(
	field *dmodel.ModelField, name string, level string, ownerType string, record *models.SettingsRecord,
) it.SettingItem {
	allowOverride := allowOverrideFor(field, record)

	item := it.SettingItem{
		Name:          name,
		Level:         level,
		AllowOverride: allowOverride,
		Editable:      isEditable(level, ownerType, allowOverride),
		Field:         field,
	}

	if record != nil {
		if val, ok := record.GetValue(); ok {
			item.Value = val
			item.HasValue = true
			return item
		}
	}
	if defaultValue := field.Default(); defaultValue != nil {
		if val := defaultValue.Get(); val != nil {
			item.Value = *val
		}
	}
	return item
}

// allowOverrideFor resolves the override policy for one item, preferring the stored row.
//
// The flag lives on the record so a tenant admin can decide it per tenant, but a row exists only
// once someone has saved the setting, and a module's declaration is still the starting point. So
// the row wins when it has ruled, and the field's metadata answers otherwise -- which keeps every
// setting behaving exactly as it did before any tenant admin touched it.
func allowOverrideFor(field *dmodel.ModelField, record *models.SettingsRecord) bool {
	if record != nil {
		if stored := record.GetAllowOverride(); stored != nil {
			return *stored
		}
	}
	return allowOverrideOf(field)
}

// isEditable reports whether this actor may change the item.
//
// A tenant owner may always edit, at every level — that is what makes enforcement possible. Below
// the tenant, an item whose schema says allow_override=false is managed by the tenant admin and is
// shown read-only rather than hidden, so the owner can see the value that applies to them.
func isEditable(level string, ownerType string, allowOverride bool) bool {
	if ownerType == c.OwnerTypeTenant {
		return true
	}
	if level != ownerType {
		return false
	}
	return allowOverride
}

// loadRecords returns this owner's stored rows for a module and level, keyed by setting name.
func (this *SettingsDomainServiceImpl) loadRecords(
	ctx corectx.Context, moduleKey string, level string, ownerType string, ownerId string,
) (map[string]*models.SettingsRecord, error) {
	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsRecordSchemaName)
	if !ok {
		return nil, errors.Errorf("loadRecords: the '%s' engine is not registered",
			models.SettingsRecordSchemaName)
	}

	// The conditions go in Graph, NOT in RepoSearchParam.Filter: Search builds its WHERE clause
	// from the graph alone and ignores Filter entirely. Passing them as a filter is silently
	// accepted and returns every row of the tenant, so one owner reads another owner's values —
	// which reads as a stale value rather than as the cross-owner leak it is.
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(
			models.SettingsRecordFieldModuleKey, dmodel.Equals, moduleKey),
		*dmodel.NewSearchNode().NewCondition(
			models.SettingsRecordFieldLevel, dmodel.Equals, level),
		*dmodel.NewSearchNode().NewCondition(
			models.SettingsRecordFieldOwnerType, dmodel.Equals, ownerType),
		*dmodel.NewSearchNode().NewCondition(
			models.SettingsRecordFieldOwnerId, dmodel.Equals, ownerId),
	)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Size:  maxSettingsPerModule,
	})
	if err != nil {
		return nil, errors.Wrap(err, "loadRecords")
	}

	byName := map[string]*models.SettingsRecord{}
	if !found.HasData {
		return byName, nil
	}
	for _, row := range found.Data.Items {
		record := models.NewSettingsRecordFrom(row)
		if name := record.GetName(); name != nil {
			byName[*name] = record
		}
	}
	return byName, nil
}

// maxSettingsPerModule bounds a module's settings page in one query. A module declaring more than
// this has outgrown a single settings pane, which is a design problem rather than a paging one.
const maxSettingsPerModule = 500

// parseStoredSchema rebuilds the model schema from its stored document, so that field types,
// defaults and allow_override metadata are read from the same declaration the values were written
// against.
func parseStoredSchema(row *models.SettingsSchema) (*dmodel.ModelSchema, error) {
	document := row.GetSchema()
	if document == nil {
		return nil, errors.New("parseStoredSchema: the stored schema document is empty")
	}
	return schemaFromDocument(document)
}
