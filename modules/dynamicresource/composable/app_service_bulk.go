package composable

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// CreateBulk writes many records under one permission check and one transaction. It lives on
// the application service rather than the domain service on purpose: the loop must call the
// OUTERMOST domain service so a module's Create/Update overrides run per row, and only the
// application layer holds that reference.
func (this *DefaultApplicationServiceImpl) CreateBulk(
	ctx corectx.Context, cmd BulkCreateCommand,
) (*BulkCreateResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionBulkCreate); cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	orgId, cErrs := this.guardCreateLike(ctx, cmd)
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	rows, cErrs := paramsToBulkRows(cmd)
	if cErrs != nil {
		return &BulkCreateResult{ClientErrors: *cErrs}, nil
	}
	forceOrgId(rows, orgId)
	return RunBulkCreate(ctx, this.domSvc, rows, nil)
}

// guardCreateLike is the permission step Create takes, shared with the multi-row actions: an
// org-scoped resource whose caller named no org is left to schema validation (see Create), one
// that did name an org has it resolved and checked before anything is written.
func (this *DefaultApplicationServiceImpl) guardCreateLike(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	if this.orgIdIsSchemaRequired() && readString(params, basemodel.FieldOrgId) == "" {
		return nil, this.AssertPermission(ctx, PermissionCreate, nil)
	}
	return this.AssertAction(ctx, PermissionCreate, params)
}

// forceOrgId stamps the resolved org onto every row: the permission was checked against ONE
// org, so a row must not be allowed to name another.
func forceOrgId(rows []BulkRow, orgId *model.Id) {
	if orgId == nil {
		return
	}
	for _, row := range rows {
		row.Fields[basemodel.FieldOrgId] = string(*orgId)
	}
}
