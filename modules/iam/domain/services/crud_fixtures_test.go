package services

import (
	"context"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
)

// Fixtures shared by the composable-resource tests of this package: they exercise the overrides
// without a database.

type stubTranx struct {
	committed  bool
	rolledBack bool
}

func (this *stubTranx) Commit() error   { this.committed = true; return nil }
func (this *stubTranx) Rollback() error { this.rolledBack = true; return nil }

type stubCrudRepo struct {
	composable.CrudRepository
	schema *dmodel.ModelSchema
	tranx  *stubTranx

	// rows answers GetOne and KeyExist by id; updates records every Update.
	rows    map[string]dmodel.DynamicFields
	updates []dmodel.DynamicFields
}

func (this *stubCrudRepo) GetOne(_ corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[dmodel.DynamicFields], error) {
	row, ok := this.rows[composable.ParamString(param.Filter, "id")]
	return &dyn.OpResult[dmodel.DynamicFields]{Data: row, HasData: ok}, nil
}

func (this *stubCrudRepo) KeyExist(_ corectx.Context, keys dmodel.DynamicFields) (*dyn.OpResult[bool], error) {
	_, ok := this.rows[composable.ParamString(keys, "id")]
	return &dyn.OpResult[bool]{Data: ok, HasData: true}, nil
}

func (this *stubCrudRepo) Update(_ corectx.Context, data dmodel.DynamicFields) (*composable.MutateResult, error) {
	this.updates = append(this.updates, data)
	return &composable.MutateResult{HasData: true}, nil
}

func (this *stubCrudRepo) Schema() *dmodel.ModelSchema { return this.schema }

func (this *stubCrudRepo) BeginTransaction(corectx.Context) (database.DbTransaction, error) {
	this.tranx = &stubTranx{}
	return this.tranx, nil
}

// fakeCrudDomain is the composable default: it runs the hooks it is handed against the command
// and answers success, with a generated id on create.
type fakeCrudDomain struct {
	composable.CrudDomainService
	repo       *stubCrudRepo
	created    dmodel.DynamicFields
	rejectNext bool
	deleted    int
}

func newFakeCrudDomain() *fakeCrudDomain {
	return &fakeCrudDomain{repo: &stubCrudRepo{schema: dmodel.DefineModel("iam_test").Build()}}
}

func (this *fakeCrudDomain) Repository() composable.CrudRepository { return this.repo }

func (this *fakeCrudDomain) Create(
	ctx corectx.Context, cmd composable.CreateCommand, options ...composable.CreateOptions,
) (*composable.CreateResult, error) {
	entity := composable.NewDynamicEntityFrom(cmd)
	vErrs := ft.ClientErrors{}
	if len(options) > 0 && options[0].BeforeValidation != nil {
		var err error
		if entity, err = options[0].BeforeValidation(ctx, entity, &vErrs); err != nil {
			return nil, err
		}
	}
	if this.rejectNext {
		vErrs.Append(*ft.NewValidationError("name", "err_test", "rejected"))
	}
	if vErrs.Count() > 0 {
		return &composable.CreateResult{ClientErrors: vErrs}, nil
	}
	this.created = entity.GetFieldData()
	this.created["id"] = "01JTESTUSER000000000000000"
	return &composable.CreateResult{Data: this.created, HasData: true}, nil
}

// Delete runs the guard it is handed against the keys, the way the default service does after
// fetching the stored record.
func (this *fakeCrudDomain) Delete(
	ctx corectx.Context, cmd composable.DeleteCommand, options ...composable.DeleteOptions,
) (*composable.MutateResult, error) {
	vErrs := ft.ClientErrors{}
	if len(options) > 0 && options[0].ValidateExtra != nil {
		keys := composable.NewDynamicEntityFrom(cmd)
		if err := options[0].ValidateExtra(ctx, keys, keys, &vErrs); err != nil {
			return nil, err
		}
	}
	if vErrs.Count() > 0 {
		return &composable.MutateResult{ClientErrors: vErrs}, nil
	}
	this.deleted++
	return &composable.MutateResult{HasData: true}, nil
}

type fakePermRepo struct {
	itPerm.PermissionRepository
	rebuiltUsers  []model.Id
	rebuiltGroups []model.Id
	rebuiltRoles  []model.Id
}

func (this *fakePermRepo) RebuildUserPermission(_ corectx.Context, userId model.Id) error {
	this.rebuiltUsers = append(this.rebuiltUsers, userId)
	return nil
}

func (this *fakePermRepo) RebuildUserPermissionsForGroup(_ corectx.Context, groupId model.Id) error {
	this.rebuiltGroups = append(this.rebuiltGroups, groupId)
	return nil
}

func (this *fakePermRepo) RebuildUserPermissionsForRole(_ corectx.Context, roleId model.Id) error {
	this.rebuiltRoles = append(this.rebuiltRoles, roleId)
	return nil
}

func testContext() corectx.Context {
	return corectx.NewRequestContext(context.Background())
}
