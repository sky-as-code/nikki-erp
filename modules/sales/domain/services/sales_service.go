// Package services holds the Sales business rules.
//
// The rules live here rather than in dynamicengines because an engine action is transport: it reads
// params and hands them on. A rule written in the action is unreachable from CQRS, from another
// module's port, and from a test that does not build an engine. See docs/wiki/07 §6.7.
package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// engineFor resolves a resource engine from the shared registry.
//
// It is a var rather than a plain function so a test can substitute the registry without building
// one, which is how the lifecycle rules are tested without a database.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the sibling packages that need another resource's
// engine — an action reaching the service of the resource it is not itself attached to.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// withTransaction runs body inside one database transaction on the named schema's repository.
//
// The rollback is deferred unconditionally: a commit that already ran makes it a no-op, and a body
// that returned an error must not leave a half-applied change behind. Note what this means for a
// refusal — a rule that rejects the request sets its result and returns nil, so the transaction
// commits harmlessly. Returning an error would roll back AND answer 500, which is wrong for a
// violation the caller could fix.
func withTransaction(
	ctx corectx.Context, schemaName string, body func(tranxCtx corectx.Context) error,
) error {
	engine, err := engineFor(schemaName)
	if err != nil {
		return err
	}

	tranx, err := engine.ResourceRepository().BeginTransaction(ctx)
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
	engine, err := engineFor(schemaName)
	if err != nil {
		return nil, err
	}
	found, err := engine.ResourceRepository().FindByKeys(ctx, dmodel.DynamicFields{idField: id})
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
// declared no_update: the client-facing rule is that a status cannot be edited through a plain
// update, and these operations are the only sanctioned way to move one.
//
// The etag of the row as read is carried into the update, so a concurrent writer that moved the
// record between the read and the write loses rather than silently overwriting.
func writeChanges(
	ctx corectx.Context, schemaName string, record dmodel.DynamicFields, changes dmodel.DynamicFields,
) error {
	engine, err := engineFor(schemaName)
	if err != nil {
		return err
	}

	update := make(dmodel.DynamicFields, len(changes)+2)
	for key, value := range changes {
		update[key] = value
	}
	update[basemodel.FieldId] = stringOf(record, basemodel.FieldId)
	update[basemodel.FieldEtag] = stringOf(record, basemodel.FieldEtag)

	_, err = engine.ResourceRepository().Update(ctx, update)
	return errors.Wrap(err, "writeChanges")
}

// violationResult refuses an operation with a business violation the caller can act on.
//
// The Field carries the schema name rather than a payload field: an operation-level refusal has no
// single offending input, and naming one would point a form at the wrong box.
func violationResult(schemaName, key, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation(schemaName, key, message))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

// notFoundResult is the answer when an operation names a record that does not exist. A client
// error rather than a server one: the id came from the caller.
func notFoundResult(schemaName, id string) *dyn.OpResult[dyn.MutateResultData] {
	return violationResult(schemaName, schemaName+".not_found",
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

// NormalizeChannelCode lowercases and trims an integration code before it is stored.
//
// The change request permits uppercase input and requires it be normalised BEFORE create, never
// after: once persisted the code is immutable, so a later normalisation would silently repoint
// every integration that had already resolved the old spelling.
func NormalizeChannelCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

// IsValidChannelCode reports whether a code matches the canonical [a-z0-9]+ format.
//
// Alphanumeric only: no whitespace, no underscore, hyphen or slash. The code travels in URLs and
// configuration files of other modules, so the narrow alphabet keeps it from needing escaping
// anywhere it appears.
func IsValidChannelCode(code string) bool {
	if code == "" {
		return false
	}
	for _, char := range code {
		isDigit := char >= '0' && char <= '9'
		isLower := char >= 'a' && char <= 'z'
		if !isDigit && !isLower {
			return false
		}
	}
	return true
}
