package services

import (
	"fmt"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	seqalgo "github.com/sky-as-code/nikki-erp/modules/essential/domain/services/sequence"
)

// seqMaxFormattedLength bounds the rendered code (BR §4.2); seqMaxDigits is the widest int64.
const (
	seqMaxFormattedLength = 512
	seqMaxDigits          = 19
	seqMaxPadding         = 32
)

// SequenceTimezoneDefaulter answers the time zone a new definition gets when it names none.
type SequenceTimezoneDefaulter func(ctx corectx.Context) string

// normalizeDefinitionCreate derives what the server owns on a new definition: the lowercased
// sub_scope and its scope_type, the ownership shape, the FF1 key and the time zone default.
func normalizeDefinitionCreate(defaultTimezone SequenceTimezoneDefaulter) dynamicresource.BeforeValidationFn {
	return func(
		ctx corectx.Context, entity *dynamicresource.DynamicEntity, vErrs *ft.ClientErrors,
	) (*dynamicresource.DynamicEntity, error) {
		fields := entity.GetFieldData()
		normalizeScope(fields)
		assertOwnershipShape(fields, vErrs)
		assertIncrementPositive(fields, vErrs)
		delete(fields, models.SeqDefFieldFirstIssuedAt)
		applyNumericDefaults(fields)
		if err := applyAlgorithmParams(fields); err != nil {
			return nil, err
		}
		if models.FieldString(fields, models.SeqDefFieldTimezone) == "" {
			fields[models.SeqDefFieldTimezone] = resolveDefaultTimezone(ctx, defaultTimezone)
		}
		return entity, nil
	}
}

// normalizeScope applies SEQ-DEF-01: sub_scope is lowercased; scope_type is the configuration scope,
// dotted with the sub_scope when there is one. Whatever scope_type the caller sent is replaced.
func normalizeScope(fields dmodel.DynamicFields) {
	scope := models.FieldString(fields, models.SeqDefFieldConfigurationScope)
	subScope := strings.ToLower(strings.TrimSpace(models.FieldString(fields, models.SeqDefFieldSubScope)))
	if subScope == "" {
		fields[models.SeqDefFieldSubScope] = nil
		fields[models.SeqDefFieldScopeType] = scope
		return
	}
	fields[models.SeqDefFieldSubScope] = subScope
	fields[models.SeqDefFieldScopeType] = scope + "." + subScope
}

// ScopeTypeOf is the scope_type normalizeScope stores, for a caller comparing a registration.
func ScopeTypeOf(configurationScope string, subScope string) string {
	fields := dmodel.DynamicFields{
		models.SeqDefFieldConfigurationScope: configurationScope,
		models.SeqDefFieldSubScope:           subScope,
	}
	normalizeScope(fields)
	return models.FieldString(fields, models.SeqDefFieldScopeType)
}

// assertOwnershipShape: a tenant definition belongs to no org, an org definition to exactly one.
func assertOwnershipShape(fields dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	orgId := models.FieldString(fields, models.SeqDefFieldOrgId)
	switch models.FieldString(fields, models.SeqDefFieldConfigurationScope) {
	case models.SeqConfigScopeTenant:
		if orgId != "" {
			vErrs.Append(*newSeqValidation(models.SeqDefFieldOrgId, ErrKeySeqInvalidInput,
				"a tenant-wide definition belongs to no organization; omit org_id"))
		}
	case models.SeqConfigScopeOrg:
		if orgId == "" {
			vErrs.Append(*dmodel.NewMissingFieldErr(models.SeqDefFieldOrgId))
		}
	}
}

// assertIncrementPositive refuses an explicit increment of 0 (BR AC07). The schema's default_value
// reads 0 as absent and would silently turn it into 1, so the range check cannot catch it.
func assertIncrementPositive(fields dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if increment, found := models.FieldInt64(fields, models.SeqDefFieldIncrement); found && increment == 0 {
		vErrs.Append(*ft.NewValidationError(models.SeqDefFieldIncrement, ft.ErrorKey("err_invalid_number_range"),
			"value must be between {{min}} and {{max}}",
			map[string]any{"min": 1, "max": int64(9000000000000000000)}))
	}
}

// applyNumericDefaults fills padding and initial_value here rather than in the schema: for both, 0 is a
// legitimate value, and a schema default_value would replace it (the schema reads 0 as absent).
func applyNumericDefaults(fields dmodel.DynamicFields) {
	if _, found := models.FieldInt64(fields, models.SeqDefFieldPadding); !found {
		fields[models.SeqDefFieldPadding] = models.SeqDefaultPadding
	}
	if _, found := models.FieldInt64(fields, models.SeqDefFieldInitialValue); !found {
		fields[models.SeqDefFieldInitialValue] = int64(1)
	}
}

// applyAlgorithmParams: FF1 gets its own random AES-256 key and no padding (the ciphertext has its
// own width); increment keeps no params. Any params the caller sent are discarded.
func applyAlgorithmParams(fields dmodel.DynamicFields) error {
	if models.FieldString(fields, models.SeqDefFieldAlgorithm) != models.SeqAlgorithmFf1Fpe {
		fields[models.SeqDefFieldAlgorithmParams] = nil
		return nil
	}
	key, err := seqalgo.NewKeyHex()
	if err != nil {
		return errors.Wrap(err, "generate the FF1 key")
	}
	fields[models.SeqDefFieldPadding] = int32(0)
	fields[models.SeqDefFieldAlgorithmParams] = map[string]any{models.SeqAlgorithmParamKey: key}
	return nil
}

func resolveDefaultTimezone(ctx corectx.Context, defaultTimezone SequenceTimezoneDefaulter) string {
	if defaultTimezone == nil {
		return models.SeqDefaultTimezone
	}
	if zone := defaultTimezone(ctx); zone != "" {
		if _, err := time.LoadLocation(zone); err == nil {
			return zone
		}
	}
	return models.SeqDefaultTimezone
}

// validateDefinitionCreate checks the configuration of a new definition.
func validateDefinitionCreate() dynamicresource.ValidateExtraFn {
	return func(
		_ corectx.Context, input *dynamicresource.DynamicEntity, _ *dynamicresource.DynamicEntity,
		vErrs *ft.ClientErrors,
	) error {
		ValidateDefinitionConfig(models.NewSequenceDefinitionFrom(input.GetFieldData()), vErrs)
		return nil
	}
}

// validateDefinitionUpdate refuses a configuration change once a number was issued (SEQ-CFG-01),
// then checks the record the update produces.
func validateDefinitionUpdate(schema *dmodel.ModelSchema) dynamicresource.ValidateExtraFn {
	return func(
		_ corectx.Context, input *dynamicresource.DynamicEntity, found *dynamicresource.DynamicEntity,
		vErrs *ft.ClientErrors,
	) error {
		if found == nil {
			return nil
		}
		stored := found.GetFieldData()
		if models.NewSequenceDefinitionFrom(stored).IsFirstIssued() {
			AssertConfigurationUnlocked(schema, input.GetFieldData(), stored, vErrs)
			return nil
		}
		ValidateDefinitionConfig(models.NewSequenceDefinitionFrom(mergeFields(stored, input.GetFieldData())), vErrs)
		return nil
	}
}

// seqAlwaysWritable are the fields an issued definition still accepts, plus the bookkeeping fields
// the update path itself carries.
var seqAlwaysWritable = map[string]bool{
	models.SeqDefFieldName: true, models.SeqDefFieldDescription: true,
	basemodel.FieldId: true, basemodel.FieldEtag: true, basemodel.FieldUpdatedAt: true,
	basemodel.FieldCreatedAt: true, basemodel.FieldCreatedBy: true, basemodel.FieldIsArchived: true,
}

// AssertConfigurationUnlocked reports every submitted field that would change a locked definition.
// A field sent with its stored value is not a change, so a client may send the whole form back.
func AssertConfigurationUnlocked(
	schema *dmodel.ModelSchema, input dmodel.DynamicFields, stored dmodel.DynamicFields, vErrs *ft.ClientErrors,
) {
	for key, value := range input {
		if seqAlwaysWritable[key] || (schema != nil && key == schema.TenantKey()) {
			continue
		}
		if sameValue(value, stored[key]) {
			continue
		}
		vErrs.Append(*newSeqViolation(key, ErrKeySeqConfigurationLocked,
			"the definition has issued numbers; create a new definition to change its rules"))
	}
}

func sameValue(left any, right any) bool {
	return fmt.Sprint(derefValue(left)) == fmt.Sprint(derefValue(right))
}

func derefValue(value any) any {
	switch typed := value.(type) {
	case *string:
		if typed != nil {
			return *typed
		}
		return nil
	case *int32:
		if typed != nil {
			return *typed
		}
		return nil
	case *int64:
		if typed != nil {
			return *typed
		}
		return nil
	case *bool:
		if typed != nil {
			return *typed
		}
		return nil
	}
	if number, ok := models.AnyInt64(value); ok {
		return number
	}
	return value
}

func mergeFields(stored dmodel.DynamicFields, submitted dmodel.DynamicFields) dmodel.DynamicFields {
	merged := make(dmodel.DynamicFields, len(stored)+len(submitted))
	for key, value := range stored {
		merged[key] = value
	}
	for key, value := range submitted {
		merged[key] = value
	}
	return merged
}
