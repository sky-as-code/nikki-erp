package composable

import (
	"sync"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// AuditRecorder is the audit port as the engine holds it: audit.LazyAuditSink in production, a
// fake in tests. Enabled is false in a binary without the auditlog module, and the engine then
// skips the before-row reads and the transactions it would only need for the entry.
type AuditRecorder interface {
	audit.AuditSink
	Enabled() bool
}

// unauditedSchemas holds the resources built with IsAudited false, for the m2m actions, which
// run on a repository and never see the onion's flag.
var unauditedSchemas sync.Map

func markUnaudited(schemaName string) {
	unauditedSchemas.Store(schemaName, true)
}

// IsAuditedSchema reports whether the resource records its mutations. A custom mutation that
// calls the port itself checks it too.
func IsAuditedSchema(schemaName string) bool {
	_, optedOut := unauditedSchemas.Load(schemaName)
	return !optedOut
}

// auditTrail records the mutations of the default domain service. Each write and its entry run in
// one transaction, joined when the caller opened one: an entry that fails to record rolls the
// write back, and a write that rolls back leaves no entry.
type auditTrail struct {
	sink   AuditRecorder
	schema *dmodel.ModelSchema
	repo   CrudRepository
}

// newAuditTrail is nil for a resource that opted out.
func newAuditTrail(schema *dmodel.ModelSchema, repo CrudRepository, sink AuditRecorder) *auditTrail {
	if sink == nil || schema == nil {
		return nil
	}
	return &auditTrail{sink: sink, schema: schema, repo: repo}
}

func (this *auditTrail) active() bool {
	return this != nil && this.sink.Enabled()
}

func (this *auditTrail) create(
	ctx corectx.Context, opts CreateOptions,
	run func(ctx corectx.Context, opts CreateOptions) (*CreateResult, error),
) (*CreateResult, error) {
	var result *CreateResult
	refused, err := InRefusableTransaction(ctx, this.repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			result, err = run(tranxCtx, opts)
			if err != nil || result == nil || result.ClientErrors.Count() > 0 || !result.HasData {
				return clientErrorsOf(result), err
			}
			return nil, this.record(tranxCtx, audit.ActionCreate, nil, result.Data)
		})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return &CreateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// updateCapture keeps what corecrud already read and validated, so the diff costs no extra read:
// found is the stored row from the existence and etag check, input the validated update, whose
// field map later hooks may still replace (the pointer stays the same).
type updateCapture struct {
	input *DynamicEntity
	found *DynamicEntity
}

func (this *updateCapture) hook(_ corectx.Context, input, found *DynamicEntity, _ *ft.ClientErrors) error {
	this.input, this.found = input, found
	return nil
}

func (this *auditTrail) update(
	ctx corectx.Context, opts UpdateOptions,
	run func(ctx corectx.Context, opts UpdateOptions) (*MutateResult, error),
) (*MutateResult, error) {
	capture := &updateCapture{}
	opts.ValidateExtra = ChainValidateExtra(opts.ValidateExtra, capture.hook)
	var result *MutateResult
	refused, err := InRefusableTransaction(ctx, this.repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			result, err = run(tranxCtx, opts)
			if err != nil || result == nil || result.ClientErrors.Count() > 0 || !result.HasData {
				return mutateClientErrorsOf(result), err
			}
			if capture.input == nil {
				return nil, nil
			}
			return nil, this.record(tranxCtx, updateActionOf(tranxCtx),
				entityFieldsOrEmpty(capture.found), capture.input.GetFieldData())
		})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return &MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// delete reads the row before removing it: a delete entry is the record as it was.
func (this *auditTrail) delete(
	ctx corectx.Context, cmd DeleteCommand, run func(ctx corectx.Context) (*MutateResult, error),
) (*MutateResult, error) {
	var result *MutateResult
	err := InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		before, err := this.storedRow(tranxCtx, cmd)
		if err != nil {
			return err
		}
		result, err = run(tranxCtx)
		if err != nil || !removedOrChanged(result) || before == nil {
			return err
		}
		return this.record(tranxCtx, audit.ActionDelete, before, nil)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// setArchived records archive or unarchive with the flag's old and new value. Setting the flag it
// already has records nothing.
func (this *auditTrail) setArchived(
	ctx corectx.Context, cmd SetArchivedCommand, run func(ctx corectx.Context) (*MutateResult, error),
) (*MutateResult, error) {
	archived, hasFlag := readBool(cmd, basemodel.FieldIsArchived)
	action := audit.ActionUnarchive
	if archived {
		action = audit.ActionArchive
	}
	var result *MutateResult
	err := InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		before, err := this.storedRow(tranxCtx, cmd)
		if err != nil {
			return err
		}
		result, err = run(tranxCtx)
		if err != nil || !removedOrChanged(result) || before == nil || !hasFlag {
			return err
		}
		after := dmodel.DynamicFields{basemodel.FieldId: before[basemodel.FieldId], basemodel.FieldIsArchived: archived}
		return this.record(tranxCtx, action, before, after)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// updateActionOf names an update after the custom action it was made under ("suspend"), which
// the request guard labelled ctx with when that action's permission check passed. A create or a
// delete keeps its own verb: "confirm" creating a fulfillment still created it.
func updateActionOf(ctx corectx.Context) string {
	if code := corectx.RequestActionOf(ctx); code != "" {
		return code
	}
	return audit.ActionUpdate
}

func removedOrChanged(result *MutateResult) bool {
	return result != nil && result.ClientErrors.Count() == 0 && result.HasData && result.Data.AffectedCount > 0
}

// storedRow is nil for a missing record, which the write then reports itself.
func (this *auditTrail) storedRow(ctx corectx.Context, cmd dmodel.DynamicFields) (dmodel.DynamicFields, error) {
	id := readString(cmd, basemodel.FieldId)
	if id == "" {
		return nil, nil
	}
	found, err := this.repo.FindByKeys(ctx, dmodel.DynamicFields{basemodel.FieldId: id})
	if err != nil {
		return nil, errors.Wrap(err, "auditTrail.storedRow")
	}
	if found == nil || !found.HasData || found.ClientErrors.Count() > 0 {
		return nil, nil
	}
	return found.Data, nil
}

// record writes one entry. A nil before is a create, a nil after a delete; anything else that
// changed no audited field writes nothing.
func (this *auditTrail) record(
	ctx corectx.Context, action string, before, after dmodel.DynamicFields,
) error {
	entry := newAuditEntry(this.schema, action, before, after)
	if before != nil && after != nil && entry.IsEmpty() {
		return nil
	}
	err := this.sink.Record(ctx, []audit.Entry{entry})
	return errors.Wrapf(err, "record %s %s", action, this.schema.Name())
}

// newAuditEntry diffs the two rows and addresses the entry by the record's id and org.
func newAuditEntry(schema *dmodel.ModelSchema, action string, before, after dmodel.DynamicFields) audit.Entry {
	oldValues, newValues := audit.Diff(schema, before, after)
	entry := audit.Entry{
		Action:       action,
		ResourceCode: schema.Name(),
		OrgId:        auditOrgOf(before, after),
		Before:       oldValues,
		After:        newValues,
	}
	if id := firstId(after, before); id != "" {
		entry.ResourceId = &id
	}
	return entry
}

func firstId(rows ...dmodel.DynamicFields) string {
	for _, row := range rows {
		if id := readString(row, basemodel.FieldId); id != "" {
			return id
		}
	}
	return ""
}

// auditOrgOf prefers the stored org, since an update need not repeat org_id.
func auditOrgOf(rows ...dmodel.DynamicFields) *model.Id {
	for _, row := range rows {
		if orgId := readString(row, basemodel.FieldOrgId); orgId != "" {
			id := model.Id(orgId)
			return &id
		}
	}
	return nil
}

func entityFieldsOrEmpty(entity *DynamicEntity) dmodel.DynamicFields {
	if fields := entityFields(entity); fields != nil {
		return fields
	}
	return dmodel.DynamicFields{}
}
