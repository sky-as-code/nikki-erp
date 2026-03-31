package dynamicmodel

/*
	DO NOT DELETE. LEAVE HERE FOR LATER USE.
*/

// func registerCoreSearchPredicates(queryBuilder orm.QueryBuilder) error {
// 	return queryBuilder.RegisterPredefinedPredicate(basemodel.FieldIsArchived, isArchivedPredicateTreatment)
// }

// func isArchivedPredicateTreatment(
// 	op dmodel.Operator, value any,
// ) (orm.PredefinedPredicateResult, ft.ClientErrors) {
// 	var zero orm.PredefinedPredicateResult
// 	if op != dmodel.Equals {
// 		return zero, orm.ClientErrorsUnsupportedFilterOperator(basemodel.FieldIsArchived)
// 	}
// 	b, err := dmodel.CoerceFilterBool(value)
// 	if err != nil {
// 		return zero, invalidBoolFilterValue(basemodel.FieldIsArchived)
// 	}
// 	if b {
// 		return archivedAtSetResult(), nil
// 	}
// 	return archivedAtNotSetResult(), nil
// }

// func invalidBoolFilterValue(field string) ft.ClientErrors {
// 	return ft.ClientErrors{
// 		*ft.NewValidationError(field, ft.ErrorKey("err_invalid_filter_value"),
// 			"value must be a boolean"),
// 	}
// }

// func archivedAtSetResult() orm.PredefinedPredicateResult {
// 	return orm.PredefinedPredicateResult{
// 		NewFieldName: basemodel.FieldArchivedAt,
// 		NewOperator:  dmodel.IsSet,
// 	}
// }

// func archivedAtNotSetResult() orm.PredefinedPredicateResult {
// 	return orm.PredefinedPredicateResult{
// 		NewFieldName: basemodel.FieldArchivedAt,
// 		NewOperator:  dmodel.IsNotSet,
// 	}
// }
