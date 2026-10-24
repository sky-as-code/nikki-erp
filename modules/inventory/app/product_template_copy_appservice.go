package app

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// Copy duplicates a template with its current configuration (CR01 §11). The copy is created
// through this service's own Create, so the caller also needs create, the copy lands in the
// source's org, and its variants are planned exactly as for a template entered by hand.
func (this *ProductTemplateApplicationServiceImpl) Copy(
	ctx corectx.Context, cmd itProduct.CopyTemplateCommand,
) (*itProduct.CopyTemplateResult, error) {
	orgId, cErrs := this.AssertAction(ctx, PermissionCopy, cmd)
	if cErrs != nil {
		return &itProduct.CopyTemplateResult{ClientErrors: *cErrs}, nil
	}
	if orgId != nil {
		if cErrs, err := this.AssertRecordInOrg(ctx, cmd, *orgId); err != nil || cErrs != nil {
			return copyFailure(cErrs, err)
		}
	}

	copier, ok := this.productSvc.(itProduct.TemplateCopier)
	if !ok {
		return nil, errors.New("Copy: the product service cannot copy templates")
	}
	built, err := copier.BuildTemplateCopy(ctx, itProduct.CopyTemplateQuery{
		TemplateId: readStringField(cmd, paramRecordId),
		Name:       cmd[itProduct.CopyTemplateFieldName],
	})
	if err != nil {
		return nil, errors.Wrap(err, "Copy")
	}
	if built.ClientErrors.Count() > 0 || !built.HasData {
		return &itProduct.CopyTemplateResult{ClientErrors: built.ClientErrors}, nil
	}

	payload := built.Data
	if orgId != nil {
		payload[basemodel.FieldOrgId] = string(*orgId)
	}
	return this.Create(ctx, payload)
}

func copyFailure(cErrs *ft.ClientErrors, err error) (*itProduct.CopyTemplateResult, error) {
	if err != nil {
		return nil, err
	}
	return &itProduct.CopyTemplateResult{ClientErrors: *cErrs}, nil
}
