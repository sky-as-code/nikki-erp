package orm

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// PredefinedPredicateAllSchemas matches any schema when registering or resolving predicates.
const PredefinedPredicateAllSchemas = "*"

// PredefinedPredicateResult carries optional rewrites after a predefined predicate runs.
type PredefinedPredicateResult struct {
	NewFieldName string
	NewOperator  dmodel.Operator
	NewValue     any
	NewValues    []any
}

// PredefinedPredicateTreatment maps a filter field to SQL condition parts for a schema.
type PredefinedPredicateTreatment func(
	operator dmodel.Operator, fieldValue any,
) (PredefinedPredicateResult, ft.ClientErrors)

type errClientSqlErrors struct {
	errors ft.ClientErrors
}

func (e *errClientSqlErrors) Error() string {
	return "sql client errors"
}

func wrapClientSqlErrors(c ft.ClientErrors) error {
	if len(c) == 0 {
		return nil
	}
	return &errClientSqlErrors{errors: c}
}

// ClientErrorsUnsupportedFilterOperator is the standard validation payload for an unsupported condition operator on a field.
func ClientErrorsUnsupportedFilterOperator(field string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_unsupported_filter_operator"),
			"unsupported operator for this field"),
	}
}

func clientErrorsNestedFieldNotSupported(field string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_nested_field_not_supported"),
			"nested field paths are not supported on this schema"),
	}
}

func (this *PgQueryBuilder) RegisterPredefinedPredicate(
	fieldName string,
	treatment PredefinedPredicateTreatment,
	schemaName ...string,
) error {
	if fieldName == "" {
		return errors.New("field name is required")
	}
	if treatment == nil {
		return errors.New("treatment is required")
	}
	for _, s := range schemaName {
		if s == "" {
			return errors.New("schema name must not be empty")
		}
	}
	schemas := schemaName
	if len(schemas) == 0 {
		schemas = []string{PredefinedPredicateAllSchemas}
	}
	return this.storePredefinedPredicates(fieldName, treatment, schemas)
}

func (this *PgQueryBuilder) storePredefinedPredicates(
	fieldName string,
	treatment PredefinedPredicateTreatment,
	schemas []string,
) error {
	for _, schema := range schemas {
		if this.predefinedPredicates[schema] == nil {
			this.predefinedPredicates[schema] = make(map[string]PredefinedPredicateTreatment)
		}
		if _, exists := this.predefinedPredicates[schema][fieldName]; exists {
			return errors.New("predicate for this field is already registered for this schema")
		}
		this.predefinedPredicates[schema][fieldName] = treatment
	}
	return nil
}

func (this *PgQueryBuilder) GetPredefinedPredicate(
	fieldName string, schemaName string,
) PredefinedPredicateTreatment {
	if this.predefinedPredicates == nil {
		return nil
	}
	if fn := lookupPredefinedPredicate(this.predefinedPredicates, schemaName, fieldName); fn != nil {
		return fn
	}
	return lookupPredefinedPredicate(this.predefinedPredicates, PredefinedPredicateAllSchemas, fieldName)
}

func lookupPredefinedPredicate(
	reg map[string]map[string]PredefinedPredicateTreatment,
	schema, field string,
) PredefinedPredicateTreatment {
	byField, ok := reg[schema]
	if !ok {
		return nil
	}
	return byField[field]
}
