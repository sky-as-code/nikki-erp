package computed

// AggregateFunction names one of the fixed aggregate functions an AggregateExpr may apply to its
// source collection. The set is closed: SQL compilation maps each name to exactly one SQL form,
// so an unknown name is a schema validation error, never a string that reaches a query.
type AggregateFunction string

const (
	AggCount         AggregateFunction = "count"
	AggCountDistinct AggregateFunction = "count_distinct"
	AggSum           AggregateFunction = "sum"
	AggAvg           AggregateFunction = "avg"
	AggMin           AggregateFunction = "min"
	AggMax           AggregateFunction = "max"
)

// IsValid reports whether the name is one of the supported aggregate functions.
func (this AggregateFunction) IsValid() bool {
	switch this {
	case AggCount, AggCountDistinct, AggSum, AggAvg, AggMin, AggMax:
		return true
	}
	return false
}
