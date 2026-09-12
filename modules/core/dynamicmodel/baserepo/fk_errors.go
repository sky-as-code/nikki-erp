package baserepo

import (
	"strings"

	"github.com/lib/pq"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// Foreign key violations are ordinary business refusals, not server faults, and this is where
// they stop being driver errors.
//
// Two things reach the database as the same SQLSTATE and mean opposite things. Writing a row
// that names a parent which does not exist is the writer's mistake — a bad id in the request.
// Deleting a parent something still references is a refusal to destroy a reference — the row
// is fine, the delete is not. Left unhandled both escape as an unwrapped driver error and the
// client gets a 500 for what it could have fixed.
const (
	// ErrorKeyForeignKeyViolation reports a reference to something that does not exist, raised
	// on INSERT and UPDATE.
	ErrorKeyForeignKeyViolation = "foreign_key_violation"

	// ErrorKeyResourceInUse reports a parent that may not be deleted while referenced. It is
	// deliberately the same key usagecheck.ErrorKeyResourceInUse reports before the statement
	// runs: a client meets one condition, whether the business check caught it or the database
	// did, and the two must not look like different problems.
	ErrorKeyResourceInUse = "resource_in_use"

	// pgForeignKeyViolation is SQLSTATE 23503, the only code either case arrives as.
	pgForeignKeyViolation = "23503"
)

// normalizeFkViolation converts a foreign key violation into client errors, and returns nil for
// anything else so the caller can pass other errors through untouched.
//
// The driver is lib/pq, so the error carries the constraint name the database refused on. That
// name is generated as "{table}_{column}_fkey", which is what lets the refusal point at a field
// rather than at the statement. It is not always recoverable — PostgreSQL truncates identifiers
// at 63 bytes, and a long table and column together do get cut — so the field is best-effort and
// the message carries the detail either way.
func normalizeFkViolation(err error, schema *dmodel.ModelSchema, isDelete bool) *ft.ClientErrors {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) || string(pgErr.Code) != pgForeignKeyViolation {
		return nil
	}

	clientErrs := ft.NewClientErrors()
	if isDelete {
		// On delete the constraint belongs to the CHILD table, so its name identifies who still
		// references this row, not a column of it. The violation is reported against the row as
		// a whole for that reason.
		clientErrs.Append(*ft.NewBusinessViolation(
			"", ErrorKeyResourceInUse,
			"this record is still referenced by "+referencingTableOf(pgErr, schema)+
				" and cannot be deleted",
			map[string]any{"referenced_by": referencingTableOf(pgErr, schema)},
		))
		return clientErrs
	}

	field := fkFieldFromConstraint(pgErr, schema)
	clientErrs.Append(*ft.NewBusinessViolation(
		field, ErrorKeyForeignKeyViolation,
		"the referenced record does not exist",
		map[string]any{"constraint": pgErr.Constraint},
	))
	return clientErrs
}

// fkFieldFromConstraint recovers the offending column from the constraint name, which the
// generator builds as "{table}_{column}_fkey" (orm.PgQueryBuilder.appendCompositeForeignKey).
// It answers "" when the name does not follow that shape or was truncated past recognition,
// which makes the violation record-level rather than wrong about which field is at fault.
func fkFieldFromConstraint(pgErr *pq.Error, schema *dmodel.ModelSchema) string {
	if pgErr.Column != "" {
		return pgErr.Column
	}

	name := strings.TrimSuffix(pgErr.Constraint, "_fkey")
	if name == pgErr.Constraint {
		return ""
	}
	if schema != nil {
		if trimmed := strings.TrimPrefix(name, schema.TableName()+"_"); trimmed != name {
			if _, ok := schema.Field(trimmed); ok {
				return trimmed
			}
		}
	}
	return ""
}

// referencingTableOf names what still points at the row being deleted. The constraint lives on
// the referencing table and is named after it, so the prefix is the answer; pq also reports the
// table directly on some errors, which is preferred when present.
func referencingTableOf(pgErr *pq.Error, schema *dmodel.ModelSchema) string {
	if pgErr.Table != "" {
		return pgErr.Table
	}
	name := strings.TrimSuffix(pgErr.Constraint, "_fkey")
	if idx := strings.LastIndex(name, "_"); idx > 0 {
		return name[:idx]
	}
	if pgErr.Constraint != "" {
		return pgErr.Constraint
	}
	return "another record"
}
