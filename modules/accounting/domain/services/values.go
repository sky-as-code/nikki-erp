package services

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// Dynamic-model getters return pointers because a dynamic field may be absent. The resolver wants
// values: an absent rate is zero and an absent flag is false. Flattening here rather than at every
// call site keeps a forgotten nil check from panicking on a row with an empty optional column.

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefInt32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func derefBool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}

func derefDecimal(value *decimal.Decimal) decimal.Decimal {
	if value == nil {
		return decimal.Zero
	}
	return *value
}

// idString flattens an optional id; model.Id is a string alias, so this is derefString renamed.
func idString(value *model.Id) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// langText picks a display string out of a localized name. The snapshot needs a single frozen
// name, so English is preferred and any other locale accepted. Map iteration is unordered, so the
// fallback picks deterministically rather than arbitrarily.
func langText(tax *models.Tax) string {
	name := tax.GetName()
	if name == nil {
		return ""
	}
	if text, ok := (*name)[model.LanguageCodeEnUs]; ok && text != "" {
		return text
	}

	bestCode := ""
	bestText := ""
	for code, text := range *name {
		if text == "" {
			continue
		}
		if bestCode == "" || string(code) < bestCode {
			bestCode = string(code)
			bestText = text
		}
	}
	return bestText
}

// conditionValue flattens a stored condition operand into a scalar and, for a list operand, its
// members. The column is typed `any` because a condition compares against whatever its field holds.
// Determination compares everything as text, so the scalar form is what it wants; the list is
// returned separately because a comma inside a stored JSON array element is data, not a separator.
func conditionValue(raw any) (string, []string) {
	switch typed := raw.(type) {
	case nil:
		return "", nil
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, _ := conditionValue(item); text != "" {
				values = append(values, text)
			}
		}
		// A scalar operand is stored as a one-element array because the column's jsonmap type
		// accepts only an object or an array. Returning that element as the scalar too lets an "eq"
		// condition read the same column an "in" condition does.
		if len(values) == 1 {
			return values[0], values
		}
		return "", values
	case []string:
		return "", typed
	case string:
		return typed, nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	case decimal.Decimal:
		return typed.String(), nil
	case *decimal.Decimal:
		if typed == nil {
			return "", nil
		}
		return typed.String(), nil
	case float64:
		// Stored numbers arrive as float64 after a JSON round-trip. Formatting through decimal
		// avoids the exponent notation strconv produces for a large threshold, which would fail to
		// parse back on the comparison side.
		return decimal.NewFromFloat(typed).String(), nil
	case int:
		return decimal.NewFromInt(int64(typed)).String(), nil
	case int32:
		return decimal.NewFromInt(int64(typed)).String(), nil
	case int64:
		return decimal.NewFromInt(typed).String(), nil
	default:
		return "", nil
	}
}
