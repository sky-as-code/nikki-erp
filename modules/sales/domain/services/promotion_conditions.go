package services

import (
	"sort"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services/pricing"
)

// Loading a program's conditions out of the three tables SALES-018 stores them in, into the shape
// the evaluator takes.
//
// The split is the whole point of D-13: the evaluator in pricing/eligibility.go is pure and knows
// nothing about repositories, so it can be tested exhaustively without a database. This file is the
// only place that turns rows into its inputs, and it is deliberately thin — every decision about
// what the conditions MEAN lives in the evaluator, not here.
//
// Three levels, because a set-valued condition ("any of these five variants") stores its members in
// their own table rather than a column (D-07):
//
//	sales_promotion_condition_groups   -> ORed with each other
//	sales_promotion_conditions         -> ANDed within a group
//	sales_promotion_condition_targets  -> the set an `in` / `not_in` reads

// loadConditionGroups reads one program's conditions, ready for pricing.IsEligible.
//
// A program with no groups comes back empty, which the evaluator reads as unconditionally eligible.
// That is correct rather than a special case: a program with no conditions is one the operator wants
// applied to everything, and "10% off everything" would otherwise be inexpressible.
func loadConditionGroups(
	ctx corectx.Context, programId string,
) ([]pricing.ConditionGroup, error) {
	groupRecords, err := searchBy(ctx,
		models.SalesPromotionConditionGroupSchemaName,
		models.SalesPromotionConditionGroupFieldProgramId, programId)
	if err != nil {
		return nil, err
	}
	if len(groupRecords) == 0 {
		return nil, nil
	}

	groups := make([]pricing.ConditionGroup, 0, len(groupRecords))
	for _, groupRecord := range groupRecords {
		groupId := stringOf(groupRecord, models.SalesPromotionConditionGroupFieldId)

		conditionRecords, err := searchBy(ctx,
			models.SalesPromotionConditionSchemaName,
			models.SalesPromotionConditionFieldGroupId, groupId)
		if err != nil {
			return nil, err
		}

		conditions := make([]pricing.Condition, 0, len(conditionRecords))
		for _, conditionRecord := range conditionRecords {
			condition, err := conditionFrom(ctx, conditionRecord)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, condition)
		}

		groups = append(groups, pricing.ConditionGroup{
			Sequence:   int32Of(groupRecord, models.SalesPromotionConditionGroupFieldSequence),
			Conditions: conditions,
		})
	}

	// Sorted by sequence so that evaluation is reproducible whatever order the rows came back in —
	// the same property BR 29 demands of conflict resolution. The evaluator ORs the groups, so the
	// order cannot change the ANSWER; it can change which group short-circuits first, and a
	// determinism requirement that holds only for the answer is one nobody can test.
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Sequence < groups[j].Sequence
	})
	return groups, nil
}

// conditionFrom turns one condition row, plus its targets, into an evaluator condition.
func conditionFrom(
	ctx corectx.Context, record dmodel.DynamicFields,
) (pricing.Condition, error) {
	condition := pricing.Condition{
		Type:         stringOf(record, models.SalesPromotionConditionFieldConditionType),
		Operator:     stringOf(record, models.SalesPromotionConditionFieldOperator),
		ValueText:    stringOf(record, models.SalesPromotionConditionFieldValueText),
		ValueDecimal: decimalOf(record, models.SalesPromotionConditionFieldValueDecimal),
		ValueFrom:    decimalOf(record, models.SalesPromotionConditionFieldValueFrom),
		ValueTo:      decimalOf(record, models.SalesPromotionConditionFieldValueTo),
	}

	conditionId := stringOf(record, models.SalesPromotionConditionFieldId)
	targetRecords, err := searchBy(ctx,
		models.SalesPromotionConditionTargetSchemaName,
		models.SalesPromotionConditionTargetFieldConditionId, conditionId)
	if err != nil {
		return condition, err
	}
	for _, targetRecord := range targetRecords {
		condition.TargetIds = append(condition.TargetIds,
			stringOf(targetRecord, models.SalesPromotionConditionTargetFieldTargetId))
	}
	return condition, nil
}

// searchBy reads every row of a schema whose field holds the given value.
//
// One page of MODEL_RULE_PAGE_MAX. A program with more conditions than that page holds is a
// misconfiguration rather than a case to paginate for — the sum-of-products shape is meant to
// express a campaign an operator can reason about, not an arbitrary boolean formula.
func searchBy(
	ctx corectx.Context, schemaName, field, value string,
) ([]dmodel.DynamicFields, error) {
	engine, err := engineFor(schemaName)
	if err != nil {
		return nil, err
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(field, dmodel.Equals, value))

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph,
		Page:  0,
		Size:  model.MODEL_RULE_PAGE_MAX_SIZE,
	})
	if err != nil {
		return nil, err
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data.Items, nil
}

// int32Of reads an integer field, treating an absent one as zero.
//
// Every numeric shape is accepted because a value that has been through a jsonb column and back
// arrives as whatever the JSON decoder chose — a whole number is a float64, not an int.
func int32Of(record dmodel.DynamicFields, field string) int32 {
	value, present := record[field]
	if !present || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int32:
		return typed
	case int64:
		return int32(typed)
	case int:
		return int32(typed)
	case float64:
		return int32(typed)
	case float32:
		return int32(typed)
	case *int32:
		if typed != nil {
			return *typed
		}
	}
	return 0
}

// decimalOf reads a decimal field, treating an absent one as zero.
//
// A decimal crosses JSON as a string precisely so it does not lose precision; a malformed one reads
// as zero rather than propagating, because a condition threshold that will not parse must not become
// a threshold of NaN.
func decimalOf(record dmodel.DynamicFields, field string) decimal.Decimal {
	value, present := record[field]
	if !present || value == nil {
		return decimal.Zero
	}
	switch typed := value.(type) {
	case decimal.Decimal:
		return typed
	case *decimal.Decimal:
		if typed != nil {
			return *typed
		}
	case string:
		if parsed, err := decimal.NewFromString(typed); err == nil {
			return parsed
		}
	case float64:
		return decimal.NewFromFloat(typed)
	}
	return decimal.Zero
}

// dateTimeOf reads a datetime field, or nil when it is absent.
func dateTimeOf(record dmodel.DynamicFields, field string) *model.ModelDateTime {
	value, present := record[field]
	if !present || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case model.ModelDateTime:
		return &typed
	case *model.ModelDateTime:
		return typed
	}
	return nil
}
