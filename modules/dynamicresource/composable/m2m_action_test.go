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
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// stubLinkWriter answers a canned result and remembers the change it was asked to write.
type stubLinkWriter struct {
	result  *MutateResult
	written *M2mChange
}

func (this *stubLinkWriter) write(
	_ corectx.Context, _ CrudRepository, change M2mChange, _ ManageM2mParam,
) (*MutateResult, error) {
	this.written = &change
	return this.result, nil
}

func TestParseM2mChangeReadsIdAndLists(t *testing.T) {
	change, cErrs := ParseM2mChange(dmodel.DynamicFields{
		basemodel.FieldId: "src_1", FieldM2mAdd: []any{"a", "b"}, FieldM2mRemove: []any{"c"},
	})

	require.Nil(t, cErrs)
	assert.Equal(t, model.Id("src_1"), change.SrcId)
	assert.ElementsMatch(t, []model.Id{"a", "b"}, change.Added)
	assert.Equal(t, []model.Id{"c"}, change.Removed)
	assert.ElementsMatch(t, []model.Id{"a", "b", "c"}, change.AllPeerIds())
}

func TestParseM2mChangeRejectsANonIdList(t *testing.T) {
	_, cErrs := ParseM2mChange(dmodel.DynamicFields{basemodel.FieldId: "src_1", FieldM2mAdd: []any{5}})

	require.NotNil(t, cErrs)
	assert.Equal(t, 1, cErrs.Count())
}

func TestManageM2mCommitsAndRunsAfterChange(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	writer := &stubLinkWriter{result: &MutateResult{HasData: true}}
	var seen *M2mChange

	result, err := manageM2mWith(ownerContext(), repo,
		dmodel.DynamicFields{basemodel.FieldId: "src_1", FieldM2mAdd: []any{"a"}},
		ManageM2mParam{Action: "test", AfterChange: func(_ corectx.Context, change M2mChange) error {
			seen = &change
			return nil
		}}, writer.write)

	require.NoError(t, err)
	assert.True(t, result.HasData)
	require.NotNil(t, seen, "a change that linked something must reach the after-change hook")
	assert.Equal(t, []model.Id{"a"}, seen.Added)
	assert.True(t, repo.tranx.committed)
}

func TestManageM2mValidateRefusalWritesNothing(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	writer := &stubLinkWriter{result: &MutateResult{HasData: true}}

	result, err := manageM2mWith(ownerContext(), repo, dmodel.DynamicFields{basemodel.FieldId: "src_1"},
		ManageM2mParam{Validate: func(_ corectx.Context, _ M2mChange, vErrs *ft.ClientErrors) error {
			vErrs.Append(*ft.NewBusinessViolation("id", "err_test", "refused"))
			return nil
		}}, writer.write)

	require.NoError(t, err)
	assert.Equal(t, 1, result.ClientErrors.Count())
	assert.Nil(t, writer.written)
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack)
}

func TestManageM2mAfterChangeFailureRollsBack(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	writer := &stubLinkWriter{result: &MutateResult{HasData: true}}

	_, err := manageM2mWith(ownerContext(), repo, dmodel.DynamicFields{basemodel.FieldId: "src_1"},
		ManageM2mParam{Action: "test", AfterChange: func(corectx.Context, M2mChange) error {
			return errors.New("rebuild failed")
		}}, writer.write)

	require.Error(t, err)
	assert.False(t, repo.tranx.committed)
	assert.True(t, repo.tranx.rolledBack)
}

func TestManageM2mNoChangeSkipsAfterChange(t *testing.T) {
	repo := &stubRepository{schema: newPlainSchema()}
	writer := &stubLinkWriter{result: &MutateResult{HasData: false}}
	called := false

	result, err := manageM2mWith(ownerContext(), repo, dmodel.DynamicFields{basemodel.FieldId: "src_1"},
		ManageM2mParam{AfterChange: func(corectx.Context, M2mChange) error {
			called = true
			return nil
		}}, writer.write)

	require.NoError(t, err)
	assert.False(t, result.HasData, "an unchanged link set answers not-found, like the old manage routes")
	assert.False(t, called)
}

func TestAssertRecordActionRefusesWithoutThePermission(t *testing.T) {
	_, appSvc := newAppService(newPlainSchema())

	cErrs, err := AssertRecordAction(deniedContext(), appSvc, dmodel.DynamicFields{basemodel.FieldId: "x"},
		ActionPermission{ActionCode: "manage_users", ResourceCode: "iam_user"})

	require.NoError(t, err)
	require.NotNil(t, cErrs)
	assert.Positive(t, cErrs.Count())
}

func TestAssertRecordActionConfinesTheRecordToTheOrg(t *testing.T) {
	domSvc, appSvc := newAppService(newOrgSchema())
	repo := domSvc.repo.(*stubRepository)

	cErrs, err := AssertRecordAction(memberContext(), appSvc,
		dmodel.DynamicFields{basemodel.FieldId: "x", basemodel.FieldOrgId: "org_mine"},
		ActionPermission{ActionCode: "manage_users"})

	require.NoError(t, err)
	require.NotNil(t, cErrs, "a record missing from the caller's org answers not-found")
	assert.Equal(t, "org_mine", repo.existKeys[basemodel.FieldOrgId])
}

func TestSearchLinkedAddsTheLinkConditionAndDefaultFields(t *testing.T) {
	peer := newFakeDomainService(newPlainSchema())

	_, err := SearchLinked(ownerContext(), peer, dmodel.DynamicFields{basemodel.FieldId: "src_1"},
		"assigned_users", "src_1", []string{"id", "name"})

	require.NoError(t, err)
	assert.Equal(t, []string{"id", "name"}, peer.seen["fields"])
	_, hasId := peer.seen[basemodel.FieldId]
	assert.False(t, hasId, "the source id must not filter the peer by its own id")
	graph, ok := peer.seen[fieldNameGraph].(*dmodel.SearchGraph)
	require.True(t, ok)
	assert.False(t, graph.IsEmpty())
}
