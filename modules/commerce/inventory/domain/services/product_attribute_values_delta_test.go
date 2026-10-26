package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func valueName(id string, enName string) attributeValueName {
	return attributeValueName{id: model.Id(id), name: model.LangJson{"en-US": enName}}
}

func nameRow(enName string) dmodel.DynamicFields {
	return dmodel.DynamicFields{"name": map[string]any{"en-US": enName}}
}

func errorFields(cErrs ft.ClientErrors) []string {
	fields := make([]string, 0, len(cErrs))
	for _, item := range cErrs {
		fields = append(fields, item.Field)
	}
	return fields
}

func TestNormalizeAttributeValueNameTrimsFoldsAndComposes(t *testing.T) {
	decomposed := "Đỏ" // "Đỏ" with a combining hook above
	assert.Equal(t, NormalizeAttributeValueName("Đỏ"), NormalizeAttributeValueName("  "+decomposed+" "))
	assert.Equal(t, NormalizeAttributeValueName("RED"), NormalizeAttributeValueName("red"))
	assert.Equal(t, NormalizeAttributeValueName("Straße"), NormalizeAttributeValueName("STRASSE"))
	assert.NotEqual(t, NormalizeAttributeValueName("Red"), NormalizeAttributeValueName("Reds"))
}

func TestFindNameClashComparesPerLocaleAndSkipsSelf(t *testing.T) {
	siblings := []attributeValueName{
		{id: "v1", name: model.LangJson{"en-US": "Red", "vi-VN": "Đỏ"}},
	}

	assert.Nil(t, findNameClash(attributeValueName{id: "v1", name: model.LangJson{"en-US": "red"}}, siblings),
		"a value never clashes with itself")
	assert.Nil(t, findNameClash(attributeValueName{name: model.LangJson{"vi-VN": "Red"}}, siblings),
		"names are compared within one locale only")

	clash := findNameClash(attributeValueName{name: model.LangJson{"vi-VN": " đỏ "}}, siblings)
	require.NotNil(t, clash)
	assert.Equal(t, model.LanguageCode("vi-VN"), clash.locale)
	assert.Equal(t, "đỏ", clash.name)
}

func TestTrimLangJsonInPlaceKeepsTheShape(t *testing.T) {
	fromJson := dmodel.DynamicFields{"name": map[string]any{"en-US": "  Red "}}
	trimLangJsonInPlace(fromJson, "name")
	assert.Equal(t, "Red", fromJson["name"].(map[string]any)["en-US"])

	fromGo := dmodel.DynamicFields{"name": model.LangJson{"en-US": " Blue"}}
	trimLangJsonInPlace(fromGo, "name")
	assert.Equal(t, "Blue", fromGo["name"].(model.LangJson)["en-US"])
}

func TestLiftValuesDeltaRemovesTheFieldAndDecodes(t *testing.T) {
	cmd := dmodel.DynamicFields{"id": "a1", "values": map[string]any{
		"add":    []any{map[string]any{"code": "red", "name": map[string]any{"en-US": "Red"}}},
		"update": []any{map[string]any{"id": "v1", "etag": "e1", "sequence": 2}},
		"remove": []any{"v2"},
	}}

	delta, malformed := itProduct.LiftProductAttributeValuesDelta(cmd)

	require.False(t, malformed)
	require.NotNil(t, delta)
	_, stillThere := cmd["values"]
	assert.False(t, stillThere, "the attribute schema must never see values")
	assert.Len(t, delta.Add, 1)
	assert.Equal(t, "v1", delta.Update[0]["id"])
	assert.Equal(t, []model.Id{"v2"}, delta.Remove)
}

func TestLiftValuesDeltaReportsMalformed(t *testing.T) {
	_, malformed := itProduct.LiftProductAttributeValuesDelta(dmodel.DynamicFields{"values": []any{"x"}})
	assert.True(t, malformed)

	delta, malformed := itProduct.LiftProductAttributeValuesDelta(dmodel.DynamicFields{"code": "c"})
	assert.False(t, malformed)
	assert.Nil(t, delta)
}

func TestValuesDeltaShapeOnCreateAcceptsAddOnly(t *testing.T) {
	delta := itProduct.ProductAttributeValuesDelta{Remove: []model.Id{"v1"}}

	assert.Equal(t, []string{"values"}, errorFields(validateValuesDeltaShape(delta, true)))
	assert.Empty(t, validateValuesDeltaShape(delta, false))
}

func TestValuesDeltaShapeRefusesMissingAndRepeatedIds(t *testing.T) {
	delta := itProduct.ProductAttributeValuesDelta{
		Update: []dmodel.DynamicFields{{"etag": "e"}, {"id": "v1"}},
		Remove: []model.Id{"v1"},
	}

	assert.Equal(t, []string{"values.update[0].id", "values.remove[0]"},
		errorFields(validateValuesDeltaShape(delta, false)))
}

func TestValuesDeltaFinalStateFlagsDuplicateAdds(t *testing.T) {
	delta := itProduct.ProductAttributeValuesDelta{
		Add: []dmodel.DynamicFields{nameRow("Red"), nameRow("Blue"), nameRow(" red ")},
	}

	cErrs := validateValuesDeltaFinalState(nil, delta)

	assert.Equal(t, []string{"values.add[0].name", "values.add[2].name"}, errorFields(cErrs))
}

func TestValuesDeltaFinalStateComparesWithStoredValuesIncludingArchived(t *testing.T) {
	existing := []attributeValueName{valueName("v1", "Red"), valueName("v2", "Blue")}
	delta := itProduct.ProductAttributeValuesDelta{Add: []dmodel.DynamicFields{nameRow("RED")}}

	assert.Equal(t, []string{"values.add[0].name"}, errorFields(validateValuesDeltaFinalState(existing, delta)))
}

func TestValuesDeltaFinalStateAllowsReusingAFreedName(t *testing.T) {
	existing := []attributeValueName{valueName("v1", "Red"), valueName("v2", "Blue")}
	delta := itProduct.ProductAttributeValuesDelta{
		Remove: []model.Id{"v1"},
		Update: []dmodel.DynamicFields{{"id": "v2", "name": map[string]any{"en-US": "Green"}}},
		Add:    []dmodel.DynamicFields{nameRow("Red"), nameRow("Blue")},
	}

	assert.Empty(t, validateValuesDeltaFinalState(existing, delta))
}

func TestValuesDeltaFinalStateFlagsAnUpdateOntoASibling(t *testing.T) {
	existing := []attributeValueName{valueName("v1", "Red"), valueName("v2", "Blue")}
	delta := itProduct.ProductAttributeValuesDelta{
		Update: []dmodel.DynamicFields{{"id": "v2", "name": map[string]any{"en-US": "red"}}},
	}

	assert.Equal(t, []string{"values.update[0].name"}, errorFields(validateValuesDeltaFinalState(existing, delta)))
}

func TestValuesDeltaFinalStateRefusesForeignIds(t *testing.T) {
	existing := []attributeValueName{valueName("v1", "Red")}
	delta := itProduct.ProductAttributeValuesDelta{
		Update: []dmodel.DynamicFields{{"id": "other", "sequence": 1}},
		Remove: []model.Id{"gone"},
	}

	assert.ElementsMatch(t, []string{"values.update[0].id", "values.remove[0]"},
		errorFields(validateValuesDeltaFinalState(existing, delta)))
}

func TestPrefixRowErrorsPointsAtTheRow(t *testing.T) {
	cErrs := ft.ClientErrors{
		*ft.NewValidationError("name", "k", "m"),
		*ft.NewBusinessViolation("", "k", "m"),
	}

	assert.Equal(t, []string{"values.add[1].name", "values.add[1]"},
		errorFields(prefixRowErrors(cErrs, "values.add[1]")))
}

func TestPickDeltaRowDropsParentFields(t *testing.T) {
	row := dmodel.DynamicFields{"id": "v1", "etag": "e", "code": "c", "attribute_id": "x", "org_id": "o"}

	assert.Equal(t, dmodel.DynamicFields{"code": "c"}, pickDeltaRow(row))
	assert.Equal(t, dmodel.DynamicFields{"code": "c", "id": "v1", "etag": "e"}, pickDeltaRow(row, "id", "etag"))
}

func TestDisplayTypeMultiCheckboxRequiresNever(t *testing.T) {
	vErrs := ft.NewClientErrors()
	assert.True(t, assertDisplayTypeFitsMode("multi_checkbox", "dynamic", vErrs))
	assert.True(t, assertDisplayTypeFitsMode("multi_checkbox", "", vErrs), "the default mode is dynamic")
	assert.False(t, assertDisplayTypeFitsMode("multi_checkbox", "never", vErrs))
	assert.False(t, assertDisplayTypeFitsMode("color", "instant", vErrs))
	assert.Equal(t, 2, vErrs.Count())
}

func TestMergedStringPrefersSubmitted(t *testing.T) {
	stored := dmodel.DynamicFields{"display_type": "pills", "variant_creation_mode": "instant"}
	submitted := dmodel.DynamicFields{"variant_creation_mode": "never"}

	assert.Equal(t, "never", mergedString(submitted, stored, "variant_creation_mode"))
	assert.Equal(t, "pills", mergedString(submitted, stored, "display_type"))
}
