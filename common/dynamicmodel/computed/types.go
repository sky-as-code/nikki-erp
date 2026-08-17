package computed

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Type is the static type of a computed expression. It reuses the canonical FieldDataType name
// strings as its values — this is deliberately NOT a parallel type system: a computed field's
// inferred type is directly comparable with ModelField.DataType().String().
type Type string

const (
	// TypeUnknown means inference could not determine a type. It is always a validation error;
	// it never silently passes through.
	TypeUnknown Type = ""
	// TypeNull is the type of a literal nil. It unifies with any other type during inference
	// (the other type wins) and is never a field's final type.
	TypeNull Type = "null"

	TypeBoolean    Type = dmodel.FieldDataTypeNameBoolean
	TypeDecimal    Type = dmodel.FieldDataTypeNameDecimal
	TypeInt32      Type = dmodel.FieldDataTypeNameInt32
	TypeInt64      Type = dmodel.FieldDataTypeNameInt64
	TypeString     Type = dmodel.FieldDataTypeNameString
	TypeEnumString Type = dmodel.FieldDataTypeNameEnumString
	TypeDate       Type = dmodel.FieldDataTypeNameModelDate
	TypeDateTime   Type = dmodel.FieldDataTypeNameModelDateTime
	TypeTime       Type = dmodel.FieldDataTypeNameModelTime
)

// IsNumeric reports whether the type participates in arithmetic.
func (t Type) IsNumeric() bool {
	return t == TypeInt32 || t == TypeInt64 || t == TypeDecimal
}

// IsTexty reports whether the type behaves as text for string functions (concat, lower, ...).
// LangJson is deliberately excluded: it is a per-language map, not a string.
func (t Type) IsTexty() bool {
	switch t {
	case TypeString, TypeEnumString,
		Type(dmodel.FieldDataTypeNameEmail), Type(dmodel.FieldDataTypeNamePhone),
		Type(dmodel.FieldDataTypeNameUrl), Type(dmodel.FieldDataTypeNameSlug):
		return true
	}
	return false
}

// IsTemporal reports whether the type is a date/time kind.
func (t Type) IsTemporal() bool {
	return t == TypeDate || t == TypeDateTime || t == TypeTime
}

// ComparableWith reports whether two types may appear on the two sides of a comparison. Null is
// comparable with anything (the comparison then yields null at evaluation time).
func (t Type) ComparableWith(other Type) bool {
	if t == TypeNull || other == TypeNull {
		return true
	}
	return sameFamily(t, other)
}

func sameFamily(a Type, b Type) bool {
	switch {
	case a.IsNumeric():
		return b.IsNumeric()
	case a.IsTexty():
		return b.IsTexty()
	case a.IsTemporal():
		return b.IsTemporal()
	default:
		return a == b
	}
}

// widenNumeric gives the result type of an arithmetic operation over two numeric operands:
// int32 stays int32, mixing widths widens to int64, and any decimal operand makes the result
// decimal. Division is handled by the caller (always decimal).
func widenNumeric(a Type, b Type) Type {
	if a == TypeDecimal || b == TypeDecimal {
		return TypeDecimal
	}
	if a == TypeInt64 || b == TypeInt64 {
		return TypeInt64
	}
	return TypeInt32
}

// unify gives the common type of two branches (CASE arms, coalesce args). Null yields to the
// other side; numeric branches widen; distinct texty types generalize to string; anything else
// must match exactly.
func unify(a Type, b Type) (Type, bool) {
	if a == TypeNull {
		return b, true
	}
	if b == TypeNull {
		return a, true
	}
	if a == b {
		return a, true
	}
	if a.IsNumeric() && b.IsNumeric() {
		return widenNumeric(a, b), true
	}
	if a.IsTexty() && b.IsTexty() {
		return TypeString, true
	}
	return TypeUnknown, false
}
