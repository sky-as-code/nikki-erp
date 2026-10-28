package composable

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

func newTestTrail(recorder *fakeRecorder) (*auditTrail, *stubRepository) {
	repo := &stubRepository{schema: newOrgSchema()}
	return newAuditTrail(repo.schema, repo, recorder), repo
}

func TestAuditTrailUpdateRecordsOnlyTheChangedFields(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, repo := newTestTrail(recorder)
	found := dmodel.DynamicFields{"id": "r1", basemodel.FieldOrgId: "org_1", "name": "old"}

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "name": "new"}, found))

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	entry := recorder.entries[0]
	assert.Equal(t, audit.ActionUpdate, entry.Action)
	assert.Equal(t, "r1", *entry.ResourceId)
	assert.Equal(t, model.Id("org_1"), *entry.OrgId)
	assert.Equal(t, dmodel.DynamicFields{"name": "old"}, entry.Before)
	assert.Equal(t, dmodel.DynamicFields{"name": "new"}, entry.After)
	assert.True(t, repo.tranx.committed)
}

func TestAuditTrailUpdateWithoutChangeWritesNoEntry(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	trail, _ := newTestTrail(recorder)
	found := dmodel.DynamicFields{"id": "r1", "name": "same"}

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "name": "same"}, found))

	require.NoError(t, err)
	assert.Empty(t, recorder.entries)
}

func TestAuditTrailFailedRecordRollsTheWriteBack(t *testing.T) {
	recorder := &fakeRecorder{enabled: true, fail: errors.New("insert failed")}
	trail, repo := newTestTrail(recorder)

	_, err := trail.update(ownerContext(), UpdateOptions{},
		runUpdateWith(dmodel.DynamicFields{"id": "r1", "name": "new"}, dmodel.DynamicFields{"id": "r1"}))

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

func TestManageM2mRecordsLinkAndUnlink(t *testing.T) {
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
	require.Len(t, recorder.entries, 2)
	assert.Equal(t, audit.ActionLink, recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"cmp_peer": []string{"a"}}, recorder.entries[0].After)
	assert.Equal(t, audit.ActionUnlink, recorder.entries[1].Action)
	assert.Equal(t, dmodel.DynamicFields{"cmp_peer": []string{"b"}}, recorder.entries[1].Before)
}

func TestPeerLinkEntriesMirrorTheChangeOnEachPeer(t *testing.T) {
	orgId := model.Id("org_1")
	change := M2mChange{SrcId: "src_1", Added: []model.Id{"a", "b"}, Removed: []model.Id{"c"}}

	entries := peerLinkEntries(newPlainSchema(), newOrgSchema(), change, &orgId)

	require.Len(t, entries, 3)
	assert.Equal(t, "cmp_org_resource", entries[0].ResourceCode)
	assert.Equal(t, "a", *entries[0].ResourceId)
	assert.Equal(t, &orgId, entries[0].OrgId, "a peer with an org column takes the source's org")
	assert.Equal(t, dmodel.DynamicFields{"cmp_plain_resource": []string{"src_1"}}, entries[1].After)
	assert.Equal(t, audit.ActionUnlink, entries[2].Action)
	assert.Equal(t, dmodel.DynamicFields{"cmp_plain_resource": []string{"src_1"}}, entries[2].Before)
}

func TestPeerLinkEntriesSkipAnUnknownOrUnauditedPeer(t *testing.T) {
	change := M2mChange{SrcId: "src_1", Added: []model.Id{"a"}}
	markUnaudited("cmp_org_resource")
	defer unauditedSchemas.Delete("cmp_org_resource")

	assert.Empty(t, peerLinkEntries(newPlainSchema(), nil, change, nil))
	assert.Empty(t, peerLinkEntries(newPlainSchema(), newOrgSchema(), change, nil))
	assert.Nil(t, peerLinkEntries(newOrgSchema(), newPlainSchema(), change, nil)[0].OrgId,
		"a peer without an org column gets no org")
}

func TestRecordLinkAuditWritesLinkAndUnlink(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	previous := sharedAuditSink
	sharedAuditSink = recorder
	defer func() { sharedAuditSink = previous }()

	require.NoError(t, RecordLinkAudit(ownerContext(), newPlainSchema(), "r1", nil, "entitlements", []string{"x"}, nil))
	require.NoError(t, RecordLinkAudit(ownerContext(), newPlainSchema(), "r1", nil, "entitlements", nil, nil))

	require.Len(t, recorder.entries, 1, "an empty change records nothing")
	assert.Equal(t, dmodel.DynamicFields{"entitlements": []string{"x"}}, recorder.entries[0].After)
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
	repo.record = dmodel.DynamicFields{"id": "w1", "name": "active"}
	repo.found = true
	ctx := ownerContext()
	corectx.SetRequestAction(ctx, "suspend")

	_, err := trail.updateRow(ctx, dmodel.DynamicFields{"id": "w1", "name": "suspended"},
		func(corectx.Context, dmodel.DynamicFields) (*MutateResult, error) { return changedOne(), nil })

	require.NoError(t, err)
	require.Len(t, recorder.entries, 1)
	assert.Equal(t, "suspend", recorder.entries[0].Action)
	assert.Equal(t, dmodel.DynamicFields{"name": "active"}, recorder.entries[0].Before)
}

func TestRecordAuditSkipsAnUnchangedUpdate(t *testing.T) {
	recorder := &fakeRecorder{enabled: true}
	previous := sharedAuditSink
	sharedAuditSink = recorder
	defer func() { sharedAuditSink = previous }()
	row := dmodel.DynamicFields{"id": "o1", "name": "same"}

	require.NoError(t, RecordAudit(ownerContext(), newOrgSchema(), "mark_deleted", row, row))
	require.NoError(t, RecordAudit(ownerContext(), newOrgSchema(), "mark_deleted", row,
		dmodel.DynamicFields{"id": "o1", "name": "gone"}))

	require.Len(t, recorder.entries, 1)
	assert.Equal(t, "mark_deleted", recorder.entries[0].Action)
}
