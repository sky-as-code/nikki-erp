package storagefield

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

type fakePort struct {
	usable  map[string]bool
	asked   []*model.Id
	deleted []string
}

func (this *fakePort) AssertFileUsable(_ corectx.Context, path string, orgId *model.Id) (*model.Id, error) {
	this.asked = append(this.asked, orgId)
	if !this.usable[path] {
		return nil, ft.NewNotFoundError("path")
	}
	id := model.Id("01FILE")
	return &id, nil
}

func (this *fakePort) DeleteFileByPath(_ corectx.Context, path string) error {
	this.deleted = append(this.deleted, path)
	return nil
}

func entity(fields dmodel.DynamicFields) *composable.DynamicEntity {
	return composable.NewDynamicEntityFrom(fields)
}

func TestCreateRefusesAPathTheOrgCannotUse(t *testing.T) {
	port := &fakePort{usable: map[string]bool{}}
	guard := NewGuard(port, ScopeOrg, "logo_file_path")
	vErrs := ft.NewClientErrors()

	err := guard.ValidateCreate()(corectx.NewRequestContext(context.Background()),
		entity(dmodel.DynamicFields{"logo_file_path": "/org/x.png", "org_id": "01ORG"}), nil, vErrs)

	require.NoError(t, err)
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "logo_file_path", (*vErrs)[0].Field)
	require.Len(t, port.asked, 1)
	assert.Equal(t, model.Id("01ORG"), *port.asked[0])
}

func TestTenantScopeAsksWithoutAnOrg(t *testing.T) {
	port := &fakePort{usable: map[string]bool{"/t/a.png": true}}
	guard := NewGuard(port, ScopeTenant, "avatar_file_path")
	vErrs := ft.NewClientErrors()

	err := guard.ValidateCreate()(corectx.NewRequestContext(context.Background()),
		entity(dmodel.DynamicFields{"avatar_file_path": "/t/a.png", "org_id": "01ORG"}), nil, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 0, vErrs.Count())
	assert.Nil(t, port.asked[0])
}

func TestUpdateDeletesTheReplacedAndClearedPathsAfterSuccess(t *testing.T) {
	port := &fakePort{usable: map[string]bool{"/o/new.png": true}}
	guard := NewGuard(port, ScopeOrg, "a_file_path", "b_file_path", "c_file_path")
	found := entity(dmodel.DynamicFields{
		"org_id": "01ORG", "a_file_path": "/o/old.png", "b_file_path": "/o/b.png", "c_file_path": "/o/c.png",
	})
	input := entity(dmodel.DynamicFields{"a_file_path": "/o/new.png", "b_file_path": nil, "c_file_path": "/o/c.png"})

	result, err := guard.Update(corectx.NewRequestContext(context.Background()), composable.UpdateOptions{},
		func(opts composable.UpdateOptions) (*composable.MutateResult, error) {
			vErrs := ft.NewClientErrors()
			require.NoError(t, opts.ValidateExtra(nil, input, found, vErrs))
			require.Equal(t, 0, vErrs.Count())
			return &composable.MutateResult{HasData: true, Data: dyn.MutateResultData{}}, nil
		})

	require.NoError(t, err)
	assert.True(t, result.HasData)
	assert.ElementsMatch(t, []string{"/o/old.png", "/o/b.png"}, port.deleted)
	require.Len(t, port.asked, 1, "an unchanged path is not re-checked")
}

func TestUpdateKeepsTheOldFileWhenTheWriteIsRefused(t *testing.T) {
	port := &fakePort{usable: map[string]bool{"/o/new.png": true}}
	guard := NewGuard(port, ScopeOrg, "a_file_path")
	found := entity(dmodel.DynamicFields{"org_id": "01ORG", "a_file_path": "/o/old.png"})
	input := entity(dmodel.DynamicFields{"a_file_path": "/o/new.png"})

	_, err := guard.Update(corectx.NewRequestContext(context.Background()), composable.UpdateOptions{},
		func(opts composable.UpdateOptions) (*composable.MutateResult, error) {
			require.NoError(t, opts.ValidateExtra(nil, input, found, ft.NewClientErrors()))
			return &composable.MutateResult{ClientErrors: ft.ClientErrors{*ft.NewEtagMismatchedError()}}, nil
		})

	require.NoError(t, err)
	assert.Empty(t, port.deleted)
}
