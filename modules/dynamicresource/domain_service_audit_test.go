package dynamicresource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/audit"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type fakeRecorder struct {
	enabled bool
	fail    error
	entries []audit.Entry
}

func (this *fakeRecorder) Enabled() bool { return this.enabled }

func (this *fakeRecorder) Record(_ corectx.Context, entries []audit.Entry) error {
	if this.fail != nil {
		return this.fail
	}
	this.entries = append(this.entries, entries...)
	return nil
}

func changedOne() *MutateResult {
	return &MutateResult{HasData: true, Data: dyn.MutateResultData{AffectedCount: 1}}
}

// runUpdateWith plays corecrud's part: it hands the validated input and the stored row to the
// ValidateExtra chain, then reports the write.
func runUpdateWith(input, found dmodel.DynamicFields) func(corectx.Context, UpdateOptions) (*MutateResult, error) {
	return func(ctx corectx.Context, opts UpdateOptions) (*MutateResult, error) {
		vErrs := ft.ClientErrors{}
		if err := opts.ValidateExtra(ctx, NewDynamicEntityFrom(input), NewDynamicEntityFrom(found), &vErrs); err != nil {
			return nil, err
		}
		return changedOne(), nil
	}
}

// newTrackedSchema flags status with track_history, so its updates are recorded and its name
// changes are not.
func newTrackedSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("cmp_tracked_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name(basemodel.FieldOrgId).DataType(dmodel.FieldDataTypeString(0, 50))).
		Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 200))).
		Field(dmodel.DefineField().Name("status").DataType(dmodel.FieldDataTypeString(0, 50)).
			Metadata(map[string]any{audit.MetadataTrackHistory: true})).
		Field(dmodel.DefineField().Name(basemodel.FieldIsArchived).DataType(dmodel.FieldDataTypeBoolean())).
		Build()
}

func newTestTrail(recorder AuditRecorder) (*auditTrail, *stubRepository) {
	repo := &stubRepository{schema: newTrackedSchema()}
	return newAuditTrail(repo.schema, repo, recorder), repo
}

func TestAuditTrailUpdateRecordsOnlyTheTrackedFields(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	found := dmodel.DynamicFields{"id": "r1", basemodel.FieldOrgId: "org_1", "name": "old", "status": "draft"}

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "name": "new", "status": "confirmed"}, found))

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, audit.ActionUpdate, entry.Action)
	assert.Equal(t, "r1", *entry.ResourceId)
	assert.Equal(t, model.Id("org_1"), *entry.OrgId)
	assert.Equal(t, dmodel.DynamicFields{"status": "draft"}, entry.Before)
	assert.Equal(t, dmodel.DynamicFields{"status": "confirmed"}, entry.After)
	assert.True(t, repo.tranx.committed)
}

func TestAuditTrailUpdateOfAnUntrackedFieldWritesNoEntry(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, _ := newTestTrail(recorder)
	found := dmodel.DynamicFields{"id": "r1", "name": "old", "status": "draft"}

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "name": "new", "status": "draft"}, found))

	require.NoError(t, err)
	assert.Empty(t, recorder.entries)
}

func TestAuditTrailSkipsUpdatesOfASchemaWithoutTrackedFields(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail := newAuditTrail(newOrgSchema(), &stubRepository{schema: newOrgSchema()}, recorder)

	assert.True(t, trail.active(), "create, archive and delete are still recorded")
	assert.False(t, trail.activeForUpdate())
	assert.False(t, trail.tracksRowUpdate(dmodel.DynamicFields{"id": "r1", "name": "x"}))
}

func TestAuditTrailFailedRecordRollsTheWriteBack(t *testing.T) {
	recorder := &fakeRecorder{enabled: true, fail: errors.New("insert failed")}
	trail, repo := newTestTrail(recorder)

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "status": "confirmed"}, dmodel.DynamicFields{"id": "r1"}))

	require.Error(t, err)
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack)
}

func TestAuditTrailDeleteRecordsTheStoredRow(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	repo.record = dmodel.DynamicFields{"id": "r1", basemodel.FieldOrgId: "org_1", "name": "gone"}
	repo.found = true

	_, err := trail.delete(ownerContext(), dmodel.DynamicFields{"id": "r1"},
		func(corectx.Context) (*MutateResult, error) { return changedOne(), nil })

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	assert.Equal(t, audit.ActionDelete, recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"name": "gone"}, recorder.entries[0].Before)
	assert.Empty(t, recorder.entries[0].After)
}

func TestAuditTrailSetArchivedRecordsTheFlag(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	repo.record = dmodel.DynamicFields{"id": "r1", basemodel.FieldIsArchived: false}
	repo.found = true

	_, err := trail.setArchived(ownerContext(), dmodel.DynamicFields{"id": "r1", basemodel.FieldIsArchived: true},
		func(corectx.Context) (*MutateResult, error) { return changedOne(), nil })

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	assert.Equal(t, audit.ActionArchive, recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{basemodel.FieldIsArchived: true}, recorder.entries[0].After)
}

func TestAuditTrailIsInactiveWithoutASink(t *testing.T) {
	trail, _ := newTestTrail(&fakeRecorder{enabled: false})

	assert.False(t, trail.active())
	assert.False(t, newAuditTrail(newOrgSchema(), nil, nil).active())
}

func TestOnionOptOutMarksTheSchemaUnaudited(t *testing.T) {
	onion := &DynamicResourceEngineOnionImpl{SchemaName: "cmp_unaudited_resource", IsAudited: new(bool)}

	assert.Nil(t, onion.auditSink())
	assert.False(t, IsAuditedSchema("cmp_unaudited_resource"))
	assert.True(t, IsAuditedSchema("cmp_some_other_resource"))
	assert.Nil(t, onion.readAuditSink(), "reads are not audited unless opted in")
}

func TestManageM2mRecordsNoLinkEntry(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	previous := sharedAuditSink
	sharedAuditSink = recorder
	defer func() { sharedAuditSink = previous }()
	repo := &stubRepository{schema: newPlainSchema()}
	writer := &stubLinkWriter{result: changedOne()}

	_, err := manageM2mWith(ownerContext(), repo,
		dmodel.DynamicFields{basemodel.FieldId: "src_1", FieldM2mAdd: []any{"a"}, FieldM2mRemove: []any{"b"}},
		ManageM2mParam{Action: "test", DestSchemaName: "cmp_peer"}, writer.write)

	require.NoError(t, err)
	assert.Empty(t, recorder.entries)
}

func TestRepositoryTrailRecordsADirectInsert(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	row := dmodel.DynamicFields{"id": "l1", basemodel.FieldOrgId: "org_1", "name": "line"}

	_, err := trail.insert(ownerContext(), row,
		func(corectx.Context, dmodel.DynamicFields) (*dyn.OpResult[int], error) {
			return &dyn.OpResult[int]{Data: 1, HasData: true}, nil
		})

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	assert.Equal(t, audit.ActionCreate, recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"name": "line"}, recorder.entries[0].After)
	assert.True(t, repo.tranx.committed)
}

func TestRepositoryTrailNamesAnUpdateAfterTheCustomAction(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	repo.record = dmodel.DynamicFields{"id": "w1", "name": "w", "status": "active"}
	repo.found = true
	ctx := ownerContext()
	corectx.SetRequestAction(ctx, "suspend")
	data := dmodel.DynamicFields{"id": "w1", "name": "w2", "status": "suspended"}

	require.True(t, trail.tracksRowUpdate(data))
	_, err := trail.updateRow(ctx, data,
		func(corectx.Context, dmodel.DynamicFields) (*MutateResult, error) { return changedOne(), nil })

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	assert.Equal(t, "suspend", recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"status": "active"}, recorder.entries[0].Before)
	assert.Equal(t, dmodel.DynamicFields{"status": "suspended"}, recorder.entries[0].After)
}

func TestRepositoryTrailIgnoresADirectUpdateOfUntrackedFields(t *testing.T) {
	trail, _ := newTestTrail(&fakeRecorder{enabled: true})

	assert.False(t, trail.tracksRowUpdate(dmodel.DynamicFields{"id": "w1", "name": "total recalculated"}))
}

func TestRecordAuditRecordsOnlyATrackedChange(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	previous := sharedAuditSink
	sharedAuditSink = recorder
	defer func() { sharedAuditSink = previous }()
	row := dmodel.DynamicFields{"id": "o1", "name": "same", "status": "active"}

	require.NoError(t, RecordAudit(ownerContext(), newTrackedSchema(), "mark_deleted", row, row))
	require.NoError(t, RecordAudit(ownerContext(), newTrackedSchema(), "mark_deleted", row,
		dmodel.DynamicFields{"id": "o1", "name": "renamed"}))
	require.NoError(t, RecordAudit(ownerContext(), newTrackedSchema(), "mark_deleted", row,
		dmodel.DynamicFields{"id": "o1", "name": "gone", "status": "deleted"}))

	require.Len(t, recorder.entries, 1, "only the change of a tracked field is recorded")
	assert.Equal(t, "mark_deleted", recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"status": "deleted"}, recorder.entries[0].After)
}
