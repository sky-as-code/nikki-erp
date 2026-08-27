package services

import (
	"github.com/shopspring/decimal"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

// Dynamic-model getters return pointers, because a dynamic field genuinely may be absent. The
// resolver wants values: an absent rate is zero, an absent flag is false, and neither is a
// condition the caller can do anything about. These helpers do that flattening in one place rather
// than at forty call sites, where a forgotten nil check would be a panic on a row that merely left
// an optional column empty.

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

// idString flattens an optional id. model.Id is a string alias, so this is derefString with a name
// that says what it holds.
func idString(value *model.Id) string {
	if value == nil {
		return ""
	}
	return string(*value)
}

// langText picks a display string out of a localized name.
//
// The snapshot needs a human-readable tax name frozen at calculation time, but it is a single
// string while the stored name is per-locale. English is preferred and any other locale accepted,
// because a name in the wrong language is still better on an audit trail than an empty one. Map
// iteration is unordered, so the fallback picks deterministically rather than arbitrarily.
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

// conditionValue flattens a stored condition operand into a scalar and, when the operand is a
// list, its members.
//
// The column is typed `any` because a condition compares against whatever its field holds — a
// string, a number, a boolean, or a list for the in/not_in operators. Determination compares
// everything as text (it parses numerically where the operator calls for it), so the scalar form
// is what it wants; the list is returned separately because a comma inside a stored JSON array
// element is data, not a separator.
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
		// A scalar operand is stored as a one-element array, because the column's jsonmap type
		// accepts only an object or an array — never a bare string. Returning that element as the
		// scalar too is what lets an "eq" condition read the same column an "in" condition does.
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
		// avoids the exponent notation strconv would produce for a large threshold, which would
		// then fail to parse back on the comparison side.
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
