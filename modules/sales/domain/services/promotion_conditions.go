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

// Loading a program's conditions out of the three tables that store them, into the shape the
// evaluator takes. The evaluator in pricing/eligibility.go is pure and knows nothing about
// repositories; this file is the only place that turns rows into its inputs, and every decision about
// what the conditions mean lives there, not here.
//
//	sales_promotion_condition_groups   -> ORed with each other
//	sales_promotion_conditions         -> ANDed within a group
//	sales_promotion_condition_targets  -> the set an `in` / `not_in` reads

// loadConditionGroups reads one program's conditions, ready for pricing.IsEligible. A program with no
// groups comes back empty, which the evaluator reads as unconditionally eligible — otherwise "10% off
// everything" would be inexpressible.
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

	// Sorted by sequence so evaluation is reproducible whatever order the rows came back in. The
	// evaluator ORs the groups, so order cannot change the answer, only which group short-circuits
	// first — but determinism that holds only for the answer is untestable.
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Sequence < groups[j].Sequence
	})
	return groups, nil
}

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

// searchBy reads every row of a schema whose field holds the given value, in one page of
// MODEL_RULE_PAGE_MAX. A program with more conditions than that is a misconfiguration rather than a
// case to paginate for.
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

// int32Of reads an integer field, treating an absent one as zero. Every numeric shape is accepted
// because a value that went through a jsonb column arrives as whatever the JSON decoder chose — a
// whole number comes back a float64.
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

// decimalOf reads a decimal field, treating an absent one as zero. A decimal crosses JSON as a string
// so it does not lose precision, and a malformed one reads as zero rather than becoming a NaN
// threshold.
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
