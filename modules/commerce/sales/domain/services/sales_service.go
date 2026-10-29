// Package services holds the Sales business rules.
//
// The rules live here rather than in dynamicengines because an engine action is transport: a rule
// written in the action is unreachable from CQRS, from another module's port, and from a test that
// does not build an engine.
package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// withTransaction runs body inside one database transaction on the named schema's repository.
//
// The rollback is deferred unconditionally: a commit that already ran makes it a no-op. Note what
// this means for a refusal: a rule that rejects the request sets its result and returns nil, so the
// transaction commits harmlessly. Returning an error would roll back and answer 500, which is wrong
// for a violation the caller could fix.
func withTransaction(
	ctx corectx.Context, schemaName string, body func(tranxCtx corectx.Context) error,
) error {
	repo, err := repoFor(schemaName)
	if err != nil {
		return err
	}

	tranx, err := repo.BeginTransaction(ctx)
	if err != nil {
		return errors.Wrap(err, "withTransaction")
	}
	defer tranx.Rollback()

	tranxCtx := corectx.CloneRequestContext(ctx)
	tranxCtx.SetDbTranx(tranx)

	if err := body(tranxCtx); err != nil {
		return err
	}
	return errors.Wrap(tranx.Commit(), "withTransaction")
}

// loadRecord fetches one row of a schema by id, or nil when there is none.
func loadRecord(
	ctx corectx.Context, schemaName string, idField string, id string,
) (dmodel.DynamicFields, error) {
	repo, err := repoFor(schemaName)
	if err != nil {
		return nil, err
	}
	found, err := repo.FindByKeys(ctx, dmodel.DynamicFields{idField: id})
	if err != nil {
		return nil, errors.Wrap(err, "loadRecord")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return found.Data, nil
}

// writeChanges applies a change set to an existing row.
//
// It goes through the repository rather than the resource service because the lifecycle fields are
// declared no_update: a status cannot be edited through a plain update, and these operations are
// the only sanctioned way to move one. The etag of the row as read is carried into the update, so a
// concurrent writer that moved the record between read and write loses rather than overwriting.
func writeChanges(
	ctx corectx.Context, schemaName string, record dmodel.DynamicFields, changes dmodel.DynamicFields,
) error {
	repo, err := repoFor(schemaName)
	if err != nil {
		return err
	}

	update := make(dmodel.DynamicFields, len(changes)+2)
	for key, value := range changes {
		update[key] = value
	}
	update[basemodel.FieldId] = stringOf(record, basemodel.FieldId)
	if _, versioned := record[basemodel.FieldEtag]; versioned {
		update[basemodel.FieldEtag] = stringOf(record, basemodel.FieldEtag)
	}

	_, err = repo.Update(ctx, update)
	return errors.Wrap(err, "writeChanges")
}

// violationResult refuses an operation with a business violation the caller can act on. field names
// the offending input; an operation-level refusal passes "" so no form box is blamed for it.
func violationResult(field, key, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey(key, "sales"), message))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// notFoundResult is the answer when an operation names a record that does not exist. A client
// error rather than a server one: the id came from the caller.
func notFoundResult(schemaName, id string) *dyn.OpResult[dyn.MutateResultData] {
	return violationResult(basemodel.FieldId, schemaName+".not_found",
		"no "+schemaName+" with id '"+id+"'")
}

// mutateOk reports one affected record.
func mutateOk() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{
		Data:    dyn.MutateResultData{AffectedCount: 1},
		HasData: true,
	}
}

func stringOf(record dmodel.DynamicFields, field string) string {
	if record == nil {
		return ""
	}
	value, ok := record[field]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(string); ok {
		return typed
	}
	if typed, ok := value.(*string); ok && typed != nil {
		return *typed
	}
	return ""
}

func boolOf(record dmodel.DynamicFields, field string) bool {
	if record == nil {
		return false
	}
	value, ok := record[field]
	if !ok || value == nil {
		return false
	}
	if typed, ok := value.(bool); ok {
		return typed
	}
	if typed, ok := value.(*bool); ok && typed != nil {
		return *typed
	}
	return false
}

// NormalizeChannelCode puts an integration code in the strict resource-code form before it is stored
// or looked up, so a kiosk still configured with the legacy "vdmc" resolves the "VDMC" channel.
//
// Normalisation must happen BEFORE create, never after: once persisted the code is immutable, so a
// later normalisation would silently repoint every integration that resolved the old spelling.
func NormalizeChannelCode(code string) string {
	return dmodel.NormalizeResourceCode(code)
}

// IsValidChannelCode reports whether a code matches the resource-code format [A-Z0-9-]+. The narrow
// alphabet keeps a code that travels in URLs and other modules' configuration free of escaping.
func IsValidChannelCode(code string) bool {
	return dmodel.IsValidResourceCode(code)
}
