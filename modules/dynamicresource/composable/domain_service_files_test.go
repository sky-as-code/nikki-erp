package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
)

type fakeStorageConsumer struct {
	asserted   []storage.AssertFilesUsableQuery
	marked     []storage.MarkFilesInUseCommand
	released   []storage.ReleaseFilesCommand
	refuse     bool
	markErr    error
	tranxOnUse []bool
}

func (this *fakeStorageConsumer) AssertFilesUsable(
	_ corectx.Context, query storage.AssertFilesUsableQuery,
) (ft.ClientErrors, error) {
	this.asserted = append(this.asserted, query)
	if !this.refuse {
		return nil, nil
	}
	cErrs := ft.ClientErrors{}
	for _, field := range query.Fields {
		cErrs.Append(*ft.NewBusinessViolation(field, ft.ErrorKey("file.err_not_usable"), "not usable"))
	}
	return cErrs, nil
}

func (this *fakeStorageConsumer) MarkFilesInUse(ctx corectx.Context, cmd storage.MarkFilesInUseCommand) error {
	this.marked = append(this.marked, cmd)
	this.tranxOnUse = append(this.tranxOnUse, ctx.GetDbTranx() != nil)
	return this.markErr
}

func (this *fakeStorageConsumer) ReleaseFiles(ctx corectx.Context, cmd storage.ReleaseFilesCommand) error {
	this.released = append(this.released, cmd)
	this.tranxOnUse = append(this.tranxOnUse, ctx.GetDbTranx() != nil)
	return nil
}

func (this *fakeStorageConsumer) DuplicateFile(
	_ corectx.Context, _ storage.DuplicateFileCommand,
) (*model.Id, error) {
	return nil, nil
}

func newFileSchema() *dmodel.ModelSchema {
	fileFlag := map[string]any{storage.MetadataFileField: true}
	return dmodel.DefineModel("cmp_file_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name(basemodel.FieldOrgId).DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name("logo_file_id").DataType(dmodel.FieldDataTypeString(0, 50)).
			Metadata(fileFlag)).
		Field(dmodel.DefineField().Name("banner_file_id").DataType(dmodel.FieldDataTypeString(0, 50)).
			Metadata(fileFlag)).
		Build()
}

func newTestLifecycle(consumer *fakeStorageConsumer) (*fileLifecycle, *stubRepository) {
	schema := newFileSchema()
	repo := &stubRepository{schema: schema}
	return newFileLifecycle(schema, repo, consumer), repo
}

// runCreate stands in for corecrud.Create: it runs the hook against the command, then saves.
func runCreate(input dmodel.DynamicFields) func(corectx.Context, CreateOptions) (*CreateResult, error) {
	return func(ctx corectx.Context, opts CreateOptions) (*CreateResult, error) {
		vErrs := ft.ClientErrors{}
		if err := opts.ValidateExtra(ctx, NewDynamicEntityFrom(input), nil, &vErrs); err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return &CreateResult{ClientErrors: vErrs}, nil
		}
		return &CreateResult{Data: input, HasData: true}, nil
	}
}

func runUpdate(input, stored dmodel.DynamicFields) func(corectx.Context, UpdateOptions) (*MutateResult, error) {
	return func(ctx corectx.Context, opts UpdateOptions) (*MutateResult, error) {
		vErrs := ft.ClientErrors{}
		err := opts.ValidateExtra(ctx, NewDynamicEntityFrom(input), NewDynamicEntityFrom(stored), &vErrs)
		if err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return &MutateResult{ClientErrors: vErrs}, nil
		}
		return &MutateResult{HasData: true}, nil
	}
}

func TestFileFieldsReadsTheMetadataFlag(t *testing.T) {
	assert.Equal(t, []string{"logo_file_id", "banner_file_id"}, fileFields(newFileSchema()))
	assert.Empty(t, fileFields(newPlainSchema()))
}

func TestNewFileLifecycleIsNilWithoutConsumerOrFileFields(t *testing.T) {
	assert.Nil(t, newFileLifecycle(newFileSchema(), nil, nil))
	assert.Nil(t, newFileLifecycle(newPlainSchema(), nil, &fakeStorageConsumer{}))
}

func TestSendsFilesSkipsWritesThatTouchNoFileField(t *testing.T) {
	files, _ := newTestLifecycle(&fakeStorageConsumer{})

	assert.False(t, files.sendsFiles(dmodel.DynamicFields{"name": "x"}, true))
	assert.False(t, files.sendsFiles(dmodel.DynamicFields{"logo_file_id": nil}, true), "an empty create")
	assert.True(t, files.sendsFiles(dmodel.DynamicFields{"logo_file_id": "f1"}, true))
	assert.True(t, files.sendsFiles(dmodel.DynamicFields{"logo_file_id": nil}, false), "a clear")
}

func TestCreateAssertsThenMarksInUseInsideTheTransaction(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, repo := newTestLifecycle(consumer)
	input := dmodel.DynamicFields{basemodel.FieldOrgId: "org1", "logo_file_id": "f1", "name": "x"}

	result, err := files.create(ownerContext(), CreateOptions{}, runCreate(input))

	require.NoError(t, err)
	assert.True(t, result.HasData)
	require.Len(t, consumer.asserted, 1)
	assert.Equal(t, map[model.Id]string{"f1": "logo_file_id"}, consumer.asserted[0].Fields)
	assert.Equal(t, model.Id("org1"), *consumer.asserted[0].OrgId)
	require.Len(t, consumer.marked, 1)
	assert.Equal(t, []model.Id{"f1"}, consumer.marked[0].FileIds)
	assert.Equal(t, []bool{true}, consumer.tranxOnUse)
	assert.True(t, repo.tranx.committed)
}

func TestCreateRefusedByStorageMarksNothingAndRollsBack(t *testing.T) {
	consumer := &fakeStorageConsumer{refuse: true}
	files, repo := newTestLifecycle(consumer)

	result, err := files.create(ownerContext(), CreateOptions{},
		runCreate(dmodel.DynamicFields{"logo_file_id": "f1"}))

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, "logo_file_id", result.ClientErrors[0].Field)
	assert.Empty(t, consumer.marked)
	assert.False(t, repo.tranx.committed)
}

func TestCreateRunsModuleRulesBeforeAskingStorage(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, _ := newTestLifecycle(consumer)

	result, err := files.create(ownerContext(), CreateOptions{ValidateExtra: violation("rule")},
		runCreate(dmodel.DynamicFields{"logo_file_id": "f1"}))

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, consumer.asserted)
}

func TestCreateRefusesOneFileInTwoFields(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, _ := newTestLifecycle(consumer)

	result, err := files.create(ownerContext(), CreateOptions{},
		runCreate(dmodel.DynamicFields{"logo_file_id": "f1", "banner_file_id": "f1"}))

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Empty(t, consumer.marked)
}

func TestCreateFailingToMarkRollsBack(t *testing.T) {
	consumer := &fakeStorageConsumer{markErr: errors.New("tag failed")}
	files, repo := newTestLifecycle(consumer)

	_, err := files.create(ownerContext(), CreateOptions{}, runCreate(dmodel.DynamicFields{"logo_file_id": "f1"}))

	require.Error(t, err)
	assert.True(t, repo.tranx.rolledBack)
	assert.False(t, repo.tranx.committed)
}

func TestCreateJoinsTheCallersTransaction(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, repo := newTestLifecycle(consumer)
	ctx := ownerContext()
	outer := &stubTransaction{}
	ctx.SetDbTranx(outer)

	_, err := files.create(ctx, CreateOptions{}, runCreate(dmodel.DynamicFields{"logo_file_id": "f1"}))

	require.NoError(t, err)
	assert.Nil(t, repo.tranx, "no nested transaction is opened")
	assert.Len(t, consumer.marked, 1)
	assert.False(t, outer.committed, "committing is the caller's decision")
}

func TestUpdateMarksTheNewFileAndReleasesTheReplacedOne(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, _ := newTestLifecycle(consumer)
	stored := dmodel.DynamicFields{basemodel.FieldOrgId: "org1", "logo_file_id": "old", "banner_file_id": "keep"}
	input := dmodel.DynamicFields{"logo_file_id": "new", "banner_file_id": "keep"}

	result, err := files.update(ownerContext(), UpdateOptions{}, runUpdate(input, stored))

	require.NoError(t, err)
	assert.True(t, result.HasData)
	require.Len(t, consumer.asserted, 1)
	assert.Equal(t, map[model.Id]string{"new": "logo_file_id"}, consumer.asserted[0].Fields)
	assert.Equal(t, model.Id("org1"), *consumer.asserted[0].OrgId, "the stored org is used")
	assert.Equal(t, []model.Id{"new"}, consumer.marked[0].FileIds)
	require.Len(t, consumer.released, 1)
	assert.Equal(t, []model.Id{"old"}, consumer.released[0].FileIds)
}

func TestUpdateClearingAFileOnlyReleasesIt(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, _ := newTestLifecycle(consumer)
	stored := dmodel.DynamicFields{"logo_file_id": "old"}

	_, err := files.update(ownerContext(), UpdateOptions{},
		runUpdate(dmodel.DynamicFields{"logo_file_id": nil}, stored))

	require.NoError(t, err)
	assert.Empty(t, consumer.asserted)
	assert.Empty(t, consumer.marked)
	assert.Equal(t, []model.Id{"old"}, consumer.released[0].FileIds)
}

func TestUpdateWithTheSameFileDoesNothing(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, _ := newTestLifecycle(consumer)
	stored := dmodel.DynamicFields{"logo_file_id": "same"}

	_, err := files.update(ownerContext(), UpdateOptions{},
		runUpdate(dmodel.DynamicFields{"logo_file_id": "same"}, stored))

	require.NoError(t, err)
	assert.Empty(t, consumer.asserted)
	assert.Empty(t, consumer.marked)
	assert.Empty(t, consumer.released)
}

func TestDeleteReleasesTheStoredFiles(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, repo := newTestLifecycle(consumer)
	repo.record = dmodel.DynamicFields{"id": "r1", "logo_file_id": "f1", "banner_file_id": nil}
	repo.found = true

	result, err := files.delete(ownerContext(), dmodel.DynamicFields{"id": "r1"},
		func(corectx.Context) (*MutateResult, error) {
			return &MutateResult{HasData: true, Data: dyn.MutateResultData{AffectedCount: 1}}, nil
		})

	require.NoError(t, err)
	assert.True(t, result.HasData)
	require.Len(t, consumer.released, 1)
	assert.Equal(t, []model.Id{"f1"}, consumer.released[0].FileIds)
	assert.Equal(t, []bool{true}, consumer.tranxOnUse)
}

func TestDeleteRefusedReleasesNothing(t *testing.T) {
	consumer := &fakeStorageConsumer{}
	files, repo := newTestLifecycle(consumer)
	repo.record = dmodel.DynamicFields{"id": "r1", "logo_file_id": "f1"}
	repo.found = true

	_, err := files.delete(ownerContext(), dmodel.DynamicFields{"id": "r1"},
		func(corectx.Context) (*MutateResult, error) {
			return &MutateResult{ClientErrors: ft.ClientErrors{*ft.NewAnonymousNotFoundError()}}, nil
		})

	require.NoError(t, err)
	assert.Empty(t, consumer.released)
}
