package services

import (
	"fmt"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	basemodel "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	c "github.com/sky-as-code/nikki-erp/modules/settings/constants"
	"github.com/sky-as-code/nikki-erp/modules/settings/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/settings/interfaces/settings"
)

// SetSettings writes the changed items of one owner at one level.
//
// Everything here happens in one transaction: a half-applied settings save would leave a module
// configured with some of the values the user chose and some of the values they replaced, which is
// a state neither they nor the module can reason about.
//
// Writes are last-write-wins per row. There is no version check, and that is safe only because the
// payload carries exclusively the items the user changed and each item is its own row: an untouched
// setting is never written, so it can never be overwritten by someone who did not edit it.
func (this *SettingsDomainServiceImpl) SetSettings(
	ctx corectx.Context, level string, ownerType string, cmd it.SetSettingsCommand,
) (*it.SetSettingsResult, error) {
	if !isKnownLevel(level) {
		return nil, errors.Errorf("SetSettings: unknown level '%s'", level)
	}

	ownerId, err := ownerIdFor(ctx, ownerType)
	if err != nil {
		return nil, err
	}

	schemaRow, err := this.loadSchema(ctx, cmd.ModuleKey, level)
	if err != nil {
		return nil, err
	}
	if schemaRow == nil {
		return clientErrorResult("module_key", "settings.schema_not_registered",
			"this module has registered no settings at this level"), nil
	}

	declared, err := parseStoredSchema(schemaRow)
	if err != nil {
		return nil, err
	}

	// The stored rows are read before validating, not during: the policy that decides whether this
	// caller may write is the one already on the row, and reading it per item inside the loop
	// would issue one query per setting in the payload.
	stored, err := this.loadRecords(ctx, cmd.ModuleKey, level, ownerType, ownerId)
	if err != nil {
		return nil, err
	}

	// Validate the whole payload before writing any of it, so a rejected item cannot leave the
	// items before it already applied.
	validated, vErrs := validatePayload(declared, level, ownerType, stored, cmd.Items)
	if vErrs.Count() > 0 {
		return &it.SetSettingsResult{ClientErrors: vErrs}, nil
	}

	written := 0
	err = withSettingsTransaction(ctx, func(tranxCtx corectx.Context) error {
		count, err := this.applyItems(tranxCtx, applyParam{
			SchemaId:  schemaRow.GetId(),
			ModuleKey: cmd.ModuleKey,
			Level:     level,
			OwnerType: ownerType,
			OwnerId:   ownerId,
			Items:     validated,
		})
		written = count
		return err
	})
	if err != nil {
		return nil, err
	}

	return &it.SetSettingsResult{
		Data:    it.SetSettingsResultData{Updated: written},
		HasData: true,
	}, nil
}

type validatedItem struct {
	Name  string
	Value any

	// AllowOverride is the policy that applies to this item after the write: the caller's new
	// value when a tenant admin sent one, and the stored-or-declared policy otherwise. The
	// fan-out reads it to decide whether to enforce, so it must reflect the decision being made
	// now rather than the one that applied before this request.
	AllowOverride bool

	// AllowOverrideChanged records that the caller asked to change the policy, so the write knows
	// to persist the column. Distinct from AllowOverride being false, which is also the resolved
	// state of an untouched item whose module declared it non-overridable.
	AllowOverrideChanged bool
}

// validatePayload checks every submitted item against its declaration and this actor's rights.
//
// Everything it rejects is something the caller can fix, so each failure is a ClientErrors entry
// rather than a Go error, which would surface as a 500 telling them nothing.
func validatePayload(
	declared *dmodel.ModelSchema, level string, ownerType string,
	stored map[string]*models.SettingsRecord, items []it.SetSettingItem,
) ([]validatedItem, ft.ClientErrors) {
	vErrs := ft.ClientErrors{}
	validated := make([]validatedItem, 0, len(items))

	for _, item := range items {
		field, ok := declared.Field(item.Name)
		if !ok {
			// AC-10: a name no schema declares is the caller's error, not a server fault.
			vErrs.Append(*ft.NewValidationError(item.Name, "settings.unknown_setting",
				"is not a setting of this module at this level"))
			continue
		}
		// The policy in force *before* this write decides whether the caller may write at all.
		// Sourcing it from the stored row matters: a tenant admin who locked this setting earlier
		// must keep an org from editing it now, and that decision lives on the row, not in the
		// module's declaration.
		allowOverride := allowOverrideFor(field, stored[item.Name])
		if !isEditable(level, ownerType, allowOverride) {
			vErrs.Append(*ft.NewBusinessViolation(item.Name, "settings.setting_is_managed",
				"is managed by your Tenant Administrator"))
			continue
		}
		// Only a tenant owner may change the policy. For anyone else the field is not merely
		// ignored but refused: silently dropping it would let an org believe it had unlocked a
		// setting the tenant had locked.
		allowOverrideChanged := false
		if item.AllowOverride != nil {
			if ownerType != c.OwnerTypeTenant {
				vErrs.Append(*ft.NewBusinessViolation(item.Name, "settings.override_not_yours",
					"only a Tenant Administrator may change whether this setting can be overridden"))
				continue
			}
			allowOverride = *item.AllowOverride
			allowOverrideChanged = true
		}
		// A nil value clears the setting back to its declared default, which every field accepts.
		if item.Value != nil {
			if _, vErr := field.Validate(item.Value, true); vErr != nil {
				vErrs.Append(*vErr)
				continue
			}
			if vErr := assertWithinDeclaredRange(field, item.Value); vErr != nil {
				vErrs.Append(*vErr)
				continue
			}
		}
		validated = append(validated, validatedItem{
			Name:                 item.Name,
			Value:                item.Value,
			AllowOverride:        allowOverride,
			AllowOverrideChanged: allowOverrideChanged,
		})
	}
	return validated, vErrs
}

type applyParam struct {
	SchemaId  *model.Id
	ModuleKey string
	Level     string
	OwnerType string
	OwnerId   string
	Items     []validatedItem
}

// applyItems writes each item's own row, then fans an enforced tenant value out onto its children.
func (this *SettingsDomainServiceImpl) applyItems(
	ctx corectx.Context, param applyParam,
) (int, error) {
	written := 0
	for _, item := range param.Items {
		count, err := this.upsertRecord(ctx, param, item)
		if err != nil {
			return written, err
		}
		written += count

		// The fan-out is the whole of the override rule: enforcement is a physical write, so a
		// child's stored value becomes the tenant's rather than being shadowed at read time.
		// Only a tenant owner can enforce, and only onto a level below its own.
		if param.OwnerType == c.OwnerTypeTenant && !item.AllowOverride && param.Level != c.LevelTenant {
			count, err := this.fanOutToChildren(ctx, param, item)
			if err != nil {
				return written, err
			}
			written += count
		}
	}
	return written, nil
}

// upsertRecord writes one owner's row for one setting name, inserting it when the owner has none.
//
// A row can legitimately be missing: rows are seeded when an owner is created, so a setting added
// to a schema afterwards has none until someone saves it.
func (this *SettingsDomainServiceImpl) upsertRecord(
	ctx corectx.Context, param applyParam, item validatedItem,
) (int, error) {
	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsRecordSchemaName)
	if !ok {
		return 0, errors.Errorf("upsertRecord: the '%s' engine is not registered",
			models.SettingsRecordSchemaName)
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			models.SettingsRecordFieldModuleKey: param.ModuleKey,
			models.SettingsRecordFieldName:      item.Name,
			models.SettingsRecordFieldOwnerId:   param.OwnerId,
		},
	})
	if err != nil {
		return 0, errors.Wrap(err, "upsertRecord")
	}

	envelope := map[string]any{models.ValueEnvelopeKey: item.Value}

	if found.HasData {
		existing := models.NewSettingsRecordFrom(found.Data)
		id := existing.GetId()
		if id == nil {
			return 0, errors.New("upsertRecord: the stored record row has no id")
		}
		updated := dmodel.DynamicFields{
			models.SettingsRecordFieldId:    *id,
			models.SettingsRecordFieldValue: envelope,
		}
		// Written only when the caller asked to change it. Sending the resolved policy on every
		// update would stamp a module's declared default onto rows a tenant admin never ruled on,
		// turning a fallback into a decision.
		if item.AllowOverrideChanged {
			updated[models.SettingsRecordFieldAllowOverride] = item.AllowOverride
		}
		updRes, err := engine.ResourceRepository().Update(ctx, updated)
		if err != nil {
			return 0, errors.Wrap(err, "upsertRecord")
		}
		// A rejected write arrives as a ClientError rather than a Go error, so discarding the
		// result reports a save that never happened.
		if updRes.ClientErrors.Count() > 0 {
			return 0, errors.Wrap(updRes.ClientErrors.ToError(), "upsertRecord")
		}
		return 1, nil
	}

	id, err := model.NewId()
	if err != nil {
		return 0, errors.Wrap(err, "upsertRecord")
	}
	data := dmodel.DynamicFields{
		models.SettingsRecordFieldId:        *id,
		models.SettingsRecordFieldModuleKey: param.ModuleKey,
		models.SettingsRecordFieldLevel:     param.Level,
		models.SettingsRecordFieldOwnerType: param.OwnerType,
		models.SettingsRecordFieldOwnerId:   param.OwnerId,
		models.SettingsRecordFieldName:      item.Name,
		models.SettingsRecordFieldValue:     envelope,
	}
	if item.AllowOverrideChanged {
		data[models.SettingsRecordFieldAllowOverride] = item.AllowOverride
	}
	if param.SchemaId != nil {
		data[models.SettingsRecordFieldSchemaId] = *param.SchemaId
	}
	// created_at is declared UseTypeDefault, but that default is applied by the engine's create
	// pipeline, which this write bypasses by going straight to the repository. The column is NOT
	// NULL, so it is filled here from the field's own declared default.
	if err := setDeclaredDefault(engine, data, basemodel.FieldCreatedAt); err != nil {
		return 0, err
	}

	insRes, err := engine.ResourceRepository().Insert(ctx, data)
	if err != nil {
		return 0, errors.Wrap(err, "upsertRecord")
	}
	if insRes.ClientErrors.Count() > 0 {
		return 0, errors.Wrap(insRes.ClientErrors.ToError(), "upsertRecord")
	}
	return 1, nil
}

func clientErrorResult(field string, key string, message string) *it.SetSettingsResult {
	vErrs := ft.ClientErrors{}
	vErrs.Append(*ft.NewBusinessViolation(field, key, message))
	return &it.SetSettingsResult{ClientErrors: vErrs}
}

// withSettingsTransaction runs body inside one transaction on a scoped copy of the context.
//
// The transaction goes on a clone, never on ctx itself: setting it on the caller's context would
// leave a committed transaction visible to whatever runs next.
func withSettingsTransaction(ctx corectx.Context, body func(tranxCtx corectx.Context) error) error {
	engine, ok := dynamicresource.Registry().GetEngine(models.SettingsRecordSchemaName)
	if !ok {
		return errors.Errorf("withSettingsTransaction: the '%s' engine is not registered",
			models.SettingsRecordSchemaName)
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withSettingsTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withSettingsTransaction")
}

// assertWithinDeclaredRange rejects a numeric value outside its field's declared bounds.
//
// This duplicates a check ModelField.Validate is supposed to perform, and does so deliberately.
// Validate treats a numeric ZERO as an absent value platform-wide — isNilOrEmpty returns true for
// it — so it returns before the range check and a submitted 0 is stored as a real value. For an
// ordinary field that is harmless, but a settings value arrives straight from an API caller and is
// read back as policy: a session timeout of 0 means every session in the tenant expires
// immediately, locking everyone out through the interface that set it.
//
// Scoped to numbers, because zero is the only value the platform helper misreads. Fixing this in
// isNilOrEmpty would change validation for every numeric field in the product, which is a larger
// decision than this module should make on its own.
func assertWithinDeclaredRange(field *dmodel.ModelField, raw any) *ft.ClientErrorItem {
	number, isNumber := asFloat(raw)
	if !isNumber {
		return nil
	}

	limits, ok := declaredNumericRange(field)
	if !ok {
		return nil
	}
	if number < limits[0] || number > limits[1] {
		return ft.NewValidationError(field.Name(), "settings.value_out_of_range",
			fmt.Sprintf("must be between %v and %v", limits[0], limits[1]))
	}
	return nil
}

// asFloat reports whether a decoded JSON value is a number, in whatever concrete type it arrived
// as. A JSON body decodes numbers as float64, but a value round-tripped through the model layer
// can be an int32, so both are accepted.
func asFloat(raw any) (float64, bool) {
	switch number := raw.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int32:
		return float64(number), true
	case int64:
		return float64(number), true
	}
	return 0, false
}

// declaredNumericRange reads a field's declared bounds, whatever slice type the data type
// constructor stored them as.
func declaredNumericRange(field *dmodel.ModelField) ([2]float64, bool) {
	raw, ok := field.DataType().Options()[dmodel.FieldDataTypeOptRange]
	if !ok {
		return [2]float64{}, false
	}

	switch limits := raw.(type) {
	case []int32:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	case []int64:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	case []int:
		if len(limits) == 2 {
			return [2]float64{float64(limits[0]), float64(limits[1])}, true
		}
	}
	return [2]float64{}, false
}
