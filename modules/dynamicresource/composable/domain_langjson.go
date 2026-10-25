package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// withLangJsonNormalization puts the LangJson normalization ahead of the caller's own hook, so a
// derived service's BeforeValidation already sees every enabled language filled in.
func (this *DefaultDomainServiceImpl) withLangJsonNormalization(callerHook BeforeValidationFn) BeforeValidationFn {
	return ChainBeforeValidation(newLangJsonNormalizer(this.schema), callerHook)
}

// newLangJsonNormalizer returns the hook that strips the languages the tenant has not enabled from
// every submitted LangJson field and fills the missing ones (model.LangJson.Normalize). Nil for a
// schema without a scalar LangJson field, so such a resource never reads the language catalogue.
func newLangJsonNormalizer(schema *dmodel.ModelSchema) BeforeValidationFn {
	fieldNames := langJsonFieldNames(schema)
	if len(fieldNames) == 0 {
		return nil
	}
	return func(ctx corectx.Context, entity *DynamicEntity, _ *ft.ClientErrors) (*DynamicEntity, error) {
		data := entity.GetFieldData()
		submitted := make([]string, 0, len(fieldNames))
		for _, name := range fieldNames {
			if data[name] != nil {
				submitted = append(submitted, name)
			}
		}
		if len(submitted) == 0 {
			return entity, nil
		}
		enabled, err := dyn.ResolveEnabledLanguages(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "normalize LangJson fields")
		}
		for _, name := range submitted {
			normalizeLangJsonField(data, name, enabled)
		}
		return entity, nil
	}
}

func langJsonFieldNames(schema *dmodel.ModelSchema) []string {
	if schema == nil {
		return nil
	}
	names := []string{}
	for name, field := range schema.Fields() {
		if field.ColumnType() == dmodel.FieldDataTypeNameLangJson && !field.IsArray() {
			names = append(names, name)
		}
	}
	return names
}

// normalizeLangJsonField leaves a value it cannot read as a whole LangJson to the schema validator:
// GetLangJson skips non-string entries, and normalizing what is left would hide the bad entry.
func normalizeLangJsonField(data dmodel.DynamicFields, name string, enabled []model.LanguageCode) {
	langObj := data.GetLangJson(name)
	if langObj == nil {
		return
	}
	if rawMap, isRawMap := data[name].(map[string]any); isRawMap && len(rawMap) != len(*langObj) {
		return
	}
	normalized := langObj.Normalize(enabled)
	data.SetLangJson(name, &normalized)
}
