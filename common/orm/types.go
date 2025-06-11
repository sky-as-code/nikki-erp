package orm

import (
	"entgo.io/ent/dialect/sql"
)

type Predicate = func(*sql.Selector)
type EdgePredicate = func(Predicate) Predicate
type OrderOption = func(*sql.Selector)

type AnyOperator = func(string, any) Predicate
type CollectionOperator = func(string, ...any) Predicate
type NullOperator = func(string) Predicate
type StringOperator = func(string, string) Predicate
type Operator = string

const (
	// Basic operators
	Equals       Operator = "="
	NotEquals    Operator = "!="
	GreaterThan  Operator = ">"
	GreaterEqual Operator = ">="
	LessThan     Operator = "<"
	LessEqual    Operator = "<="

	// Text search operators
	Contains      Operator = "*"
	NotContains   Operator = "!*"
	StartsWith    Operator = "^"
	NotStartsWith Operator = "!^"
	EndsWith      Operator = "$"
	NotEndsWith   Operator = "!$"

	// Container operators
	In    Operator = "in"
	NotIn Operator = "not_in"

	// Null operators
	IsSet    Operator = "is_set"
	IsNotSet Operator = "not_set"
)

var AnyOperators = map[Operator]AnyOperator{
	Equals:       sql.FieldEQ,
	NotEquals:    sql.FieldNEQ,
	GreaterThan:  sql.FieldGT,
	GreaterEqual: sql.FieldGTE,
	LessThan:     sql.FieldLT,
	LessEqual:    sql.FieldLTE,
}

var CollectionOperators = map[Operator]CollectionOperator{
	In:    sql.FieldIn[any],
	NotIn: sql.FieldNotIn[any],
}

var NullOperators = map[Operator]NullOperator{
	IsNotSet: sql.FieldIsNull,
	IsSet:    sql.FieldNotNull,
}

var StringOperators = map[Operator]StringOperator{
	Contains: sql.FieldContainsFold,
	NotContains: func(field string, value string) Predicate {
		return sql.NotPredicates(sql.FieldContainsFold(field, value))
	},
	StartsWith: sql.FieldHasPrefixFold,
	NotStartsWith: func(field string, value string) Predicate {
		return sql.NotPredicates(sql.FieldHasPrefixFold(field, value))
	},
	EndsWith: sql.FieldHasSuffixFold,
	NotEndsWith: func(field string, value string) Predicate {
		return sql.NotPredicates(sql.FieldHasSuffixFold(field, value))
	},
}

var NoopPredicate = func(*sql.Selector) {}

type GenericPredicate[TParam ~Predicate, TReturn ~Predicate] = func(...TParam) TReturn

func ToEdgePredicate[
	TParam ~Predicate,
	TReturn ~Predicate,
](sourceFn GenericPredicate[TParam, TReturn]) EdgePredicate {
	return func(pred Predicate) Predicate {
		return sourceFn(pred)
	}
}
