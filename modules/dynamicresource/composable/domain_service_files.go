package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
)

// fileLifecycle keeps a resource's `<name>_file_id` columns (metadata.file_field) in step with
// the storage module. A write that references a file asserts the file is usable before saving,
// then marks it in use in the same transaction as the save; a file the write replaced or cleared,
// or the files of a deleted record, are released back to pending so their objects expire. A file
// is never shared between records, so releasing one can never pull it from under another record.
//
// There is deliberately no foreign key to the file table: the reference crosses into a module
// whose rows a lifecycle rule removes, and the port is the only thing allowed to judge a file.
type fileLifecycle struct {
	consumer storage.StorageConsumer
	repo     CrudRepository
	fields   []string
}

// newFileLifecycle is nil when the schema has no file field or the binary registers no storage
// consumer, so a resource without files pays nothing.
func newFileLifecycle(
	schema *dmodel.ModelSchema, repo CrudRepository, consumer storage.StorageConsumer,
) *fileLifecycle {
	if consumer == nil || schema == nil {
		return nil
	}
	fields := fileFields(schema)
	if len(fields) == 0 {
		return nil
	}
	return &fileLifecycle{consumer: consumer, repo: repo, fields: fields}
}

// fileFields lists the stored columns flagged as file references.
func fileFields(schema *dmodel.ModelSchema) []string {
	fields := []string{}
	for _, field := range schema.Columns() {
		if flag, ok := field.MetadataValue(storage.MetadataFileField); ok && flag == true {
			fields = append(fields, field.Name())
		}
	}
	return fields
}

// fileChanges is what one write does to its file fields, gathered by the ValidateExtra hook.
type fileChanges struct {
	added   map[model.Id]string
	removed []model.Id
	orgId   *model.Id
}

func (this *fileChanges) addedIds() []model.Id {
	ids := make([]model.Id, 0, len(this.added))
	for id := range this.added {
		ids = append(ids, id)
	}
	return ids
}

// sendsFiles tells whether a command touches any file field. A create that leaves every file
// field empty, or an update that sends none, skips the lifecycle and its transaction.
func (this *fileLifecycle) sendsFiles(cmd dmodel.DynamicFields, isCreate bool) bool {
	for _, field := range this.fields {
		_, sent := cmd[field]
		if sent && (!isCreate || fileIdOf(cmd, field) != "") {
			return true
		}
	}
	return false
}

// collect is chained last into the write's ValidateExtra: the module's own rules refuse first,
// and the storage port is asked only about a write that would otherwise go through.
func (this *fileLifecycle) collect(changes *fileChanges) ValidateExtraFn {
	return func(ctx corectx.Context, input, found *DynamicEntity, vErrs *ft.ClientErrors) error {
		inputFields, foundFields := input.GetFieldData(), entityFields(found)
		changes.orgId = orgOf(inputFields, foundFields)
		for _, field := range this.fields {
			if _, sent := inputFields[field]; !sent {
				continue
			}
			newId, oldId := fileIdOf(inputFields, field), fileIdOf(foundFields, field)
			if newId == oldId {
				continue
			}
			if newId != "" {
				if _, taken := changes.added[newId]; taken {
					vErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("file.err_file_already_referenced"),
						"a file can be referenced by one field only"))
					continue
				}
				changes.added[newId] = field
			}
			if oldId != "" {
				changes.removed = append(changes.removed, oldId)
			}
		}
		if len(changes.added) == 0 || vErrs.Count() > 0 {
			return nil
		}
		cErrs, err := this.consumer.AssertFilesUsable(ctx, storage.AssertFilesUsableQuery{
			Fields: changes.added, OrgId: changes.orgId,
		})
		if err != nil {
			return errors.Wrap(err, "fileLifecycle.AssertFilesUsable")
		}
		vErrs.Concat(cErrs)
		return nil
	}
}

// commit runs inside the write's transaction once the row is saved: a failure here rolls the
// save back, so a record never commits pointing at a file that is still pending.
func (this *fileLifecycle) commit(ctx corectx.Context, changes *fileChanges) error {
	if len(changes.added) > 0 {
		err := this.consumer.MarkFilesInUse(ctx, storage.MarkFilesInUseCommand{
			FileIds: changes.addedIds(), OrgId: changes.orgId,
		})
		if err != nil {
			return errors.Wrap(err, "fileLifecycle.MarkFilesInUse")
		}
	}
	if len(changes.removed) > 0 {
		err := this.consumer.ReleaseFiles(ctx, storage.ReleaseFilesCommand{FileIds: changes.removed})
		return errors.Wrap(err, "fileLifecycle.ReleaseFiles")
	}
	return nil
}

func (this *fileLifecycle) create(
	ctx corectx.Context, opts CreateOptions,
	run func(ctx corectx.Context, opts CreateOptions) (*CreateResult, error),
) (*CreateResult, error) {
	changes := &fileChanges{added: map[model.Id]string{}}
	opts.ValidateExtra = ChainValidateExtra(opts.ValidateExtra, this.collect(changes))
	var result *CreateResult
	refused, err := InRefusableTransaction(ctx, this.repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			result, err = run(tranxCtx, opts)
			if err != nil || result == nil || result.ClientErrors.Count() > 0 {
				return clientErrorsOf(result), err
			}
			return nil, this.commit(tranxCtx, changes)
		})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return &CreateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

func (this *fileLifecycle) update(
	ctx corectx.Context, opts UpdateOptions,
	run func(ctx corectx.Context, opts UpdateOptions) (*MutateResult, error),
) (*MutateResult, error) {
	changes := &fileChanges{added: map[model.Id]string{}}
	opts.ValidateExtra = ChainValidateExtra(opts.ValidateExtra, this.collect(changes))
	var result *MutateResult
	refused, err := InRefusableTransaction(ctx, this.repo,
		func(tranxCtx corectx.Context) (*ft.ClientErrors, error) {
			var err error
			result, err = run(tranxCtx, opts)
			if err != nil || result == nil || result.ClientErrors.Count() > 0 {
				return mutateClientErrorsOf(result), err
			}
			if !result.HasData {
				return nil, nil
			}
			return nil, this.commit(tranxCtx, changes)
		})
	if err != nil {
		return nil, err
	}
	if refused != nil {
		return &MutateResult{ClientErrors: *refused}, nil
	}
	return result, nil
}

// delete reads the record's files before removing it and releases them in the same
// transaction, so a delete that rolls back leaves its files in use.
func (this *fileLifecycle) delete(
	ctx corectx.Context, cmd DeleteCommand,
	run func(ctx corectx.Context) (*MutateResult, error),
) (*MutateResult, error) {
	var result *MutateResult
	err := InTransaction(ctx, this.repo, func(tranxCtx corectx.Context) error {
		fileIds, err := this.storedFileIds(tranxCtx, cmd)
		if err != nil {
			return err
		}
		result, err = run(tranxCtx)
		if err != nil || result == nil || result.ClientErrors.Count() > 0 || !result.HasData {
			return err
		}
		if len(fileIds) == 0 {
			return nil
		}
		err = this.consumer.ReleaseFiles(tranxCtx, storage.ReleaseFilesCommand{FileIds: fileIds})
		return errors.Wrap(err, "fileLifecycle.ReleaseFiles")
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (this *fileLifecycle) storedFileIds(ctx corectx.Context, cmd DeleteCommand) ([]model.Id, error) {
	id := readId(cmd, basemodel.FieldId)
	if id == "" {
		return nil, nil
	}
	found, err := this.repo.FindByKeys(ctx, dmodel.DynamicFields{basemodel.FieldId: id})
	if err != nil {
		return nil, errors.Wrap(err, "fileLifecycle.storedFileIds")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	fileIds := []model.Id{}
	for _, field := range this.fields {
		if fileId := fileIdOf(found.Data, field); fileId != "" {
			fileIds = append(fileIds, fileId)
		}
	}
	return fileIds, nil
}

// orgOf prefers the stored org, since an update need not repeat org_id and a record never moves
// org. A schema without org_id, or a global record, resolves to nil: the tenant's space.
func orgOf(input, found dmodel.DynamicFields) *model.Id {
	for _, fields := range []dmodel.DynamicFields{found, input} {
		if orgId := fileIdOf(fields, basemodel.FieldOrgId); orgId != "" {
			return &orgId
		}
	}
	return nil
}

// fileIdOf reads an id-typed value as a field map carries it (model.Id aliases string).
func fileIdOf(fields dmodel.DynamicFields, field string) model.Id {
	if fields == nil {
		return ""
	}
	switch value := fields[field].(type) {
	case string:
		return value
	case *string:
		if value != nil {
			return *value
		}
	}
	return ""
}

func entityFields(entity *DynamicEntity) dmodel.DynamicFields {
	if entity == nil {
		return nil
	}
	return entity.GetFieldData()
}

func clientErrorsOf(result *CreateResult) *ft.ClientErrors {
	if result == nil {
		return nil
	}
	return &result.ClientErrors
}

func mutateClientErrorsOf(result *MutateResult) *ft.ClientErrors {
	if result == nil {
		return nil
	}
	return &result.ClientErrors
}
