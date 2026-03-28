package repository

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
)

func registerIdentitySearchPredicates(queryBuilder orm.QueryBuilder) error {
	return queryBuilder.RegisterPredefinedPredicate(
		domain.UserFieldIsLocked, userIsLockedPredicateTreatment, domain.UserSchemaName)
}

func userIsLockedPredicateTreatment(
	op dmodel.Operator, value any,
) (orm.PredefinedPredicateResult, ft.ClientErrors) {
	var zero orm.PredefinedPredicateResult
	if op != dmodel.Equals && op != dmodel.NotEquals {
		return zero, orm.ClientErrorsUnsupportedFilterOperator(domain.UserFieldIsLocked)
	}
	b, err := dmodel.CoerceFilterBool(value)
	if err != nil {
		return zero, userLockedInvalidValue()
	}
	locked := string(domain.UserStatusLocked)
	if b {
		return userLockedEqualsStatus(locked), nil
	}
	return userLockedNotEqualsStatus(locked), nil
}

func userLockedInvalidValue() ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(domain.UserFieldIsLocked, ft.ErrorKey("err_invalid_filter_value"),
			"value must be a boolean"),
	}
}

func userLockedEqualsStatus(locked string) orm.PredefinedPredicateResult {
	return orm.PredefinedPredicateResult{
		NewFieldName: domain.UserFieldStatus,
		NewOperator:  dmodel.Equals,
		NewValue:     locked,
	}
}

func userLockedNotEqualsStatus(locked string) orm.PredefinedPredicateResult {
	return orm.PredefinedPredicateResult{
		NewFieldName: domain.UserFieldStatus,
		NewOperator:  dmodel.NotEquals,
		NewValue:     locked,
	}
}
