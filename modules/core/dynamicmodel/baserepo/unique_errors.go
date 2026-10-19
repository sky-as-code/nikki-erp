package baserepo

import (
	"strings"

	"github.com/lib/pq"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// normalizeWriteViolation turns the constraint violations an INSERT or UPDATE can raise into client
// errors, and returns nil for anything else.
func normalizeWriteViolation(err error, schema *dmodel.ModelSchema) *ft.ClientErrors {
	if fkErrs := normalizeFkViolation(err, schema, false); fkErrs != nil {
		return fkErrs
	}
	return normalizeUniqueViolation(err, schema)
}

// pgUniqueViolation is SQLSTATE 23505.
const pgUniqueViolation = "23505"

// normalizeUniqueViolation converts a unique violation the pre-insert collision check could not
// see into the same client error that check reports, and returns nil for anything else. The check
// covers plain and composite uniques; a partial unique (a NULL branch) reaches the database.
func normalizeUniqueViolation(err error, schema *dmodel.ModelSchema) *ft.ClientErrors {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) || string(pgErr.Code) != pgUniqueViolation {
		return nil
	}
	clientErrs := ft.NewClientErrors()
	clientErrs.Append(*ft.NewAnonymousBusinessViolation(
		ft.ErrorKey("err_unique_constraint_violated"),
		"unique constraint violated {{uniques}}",
		map[string]any{"uniques": uniqueGroupsOfConstraint(pgErr.Constraint, schema)},
	))
	return clientErrs
}

// uniqueGroupsOfConstraint names the fields of the index the database refused on. Only groups
// declaring an index_name can be recognised; anything else reports no group rather than a guess.
func uniqueGroupsOfConstraint(constraint string, schema *dmodel.ModelSchema) [][]string {
	for _, group := range schema.PartialUniques() {
		stem := strings.ToLower(strings.TrimSpace(group.IndexName))
		if stem == "" {
			continue
		}
		withNullable := append(append([]string{}, group.NotNullFields...), group.NullableField)
		switch constraint {
		case stem + "_ukey_null":
			return [][]string{group.NotNullFields}
		case stem + "_ukey_notnull", stem + "_ukey":
			return [][]string{withNullable}
		}
	}
	for _, group := range schema.AllUniqueGroups() {
		stem := strings.ToLower(strings.TrimSpace(group.IndexName))
		if stem != "" && constraint == stem+"_ukey" {
			return [][]string{group.Fields}
		}
	}
	return [][]string{}
}
