package app

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	itProduct "github.com/sky-as-code/nikki-erp/modules/commerce/inventory/interfaces/product"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// Duplicate duplicates a template with its current configuration and its own clone of the preview
// image (CR01 §11). The duplicate is created through this service's own Create, so the caller also
// needs create, the duplicate lands in the source's org, its variants are planned exactly as for a
// template entered by hand, and the composable default marks the cloned image in use.
func (this *ProductTemplateApplicationServiceImpl) Duplicate(
	ctx corectx.Context, cmd itProduct.DuplicateTemplateCommand,
) (*itProduct.DuplicateTemplateResult, error) {
	orgId, cErrs := this.AssertAction(ctx, PermissionDuplicate, cmd)
	if cErrs != nil {
		return &itProduct.DuplicateTemplateResult{ClientErrors: *cErrs}, nil
	}
	if orgId != nil {
		if cErrs, err := this.AssertRecordInOrg(ctx, cmd, *orgId); err != nil || cErrs != nil {
			return duplicateFailure(cErrs, err)
		}
	}

	duplicator, ok := this.productSvc.(itProduct.TemplateDuplicator)
	if !ok {
		return nil, errors.New("Duplicate: the product service cannot duplicate templates")
	}
	built, err := duplicator.BuildTemplateDuplicate(ctx, itProduct.DuplicateTemplateQuery{
		TemplateId: readStringField(cmd, paramRecordId),
		Name:       cmd[itProduct.DuplicateTemplateFieldName],
	})
	if err != nil {
		return nil, errors.Wrap(err, "Duplicate")
	}
	if built.ClientErrors.Count() > 0 || !built.HasData {
		return &itProduct.DuplicateTemplateResult{ClientErrors: built.ClientErrors}, nil
	}

	payload := built.Data
	if orgId != nil {
		payload[basemodel.FieldOrgId] = string(*orgId)
	}
	return this.Create(ctx, payload)
}

func duplicateFailure(cErrs *ft.ClientErrors, err error) (*itProduct.DuplicateTemplateResult, error) {
	if err != nil {
		return nil, err
	}
	return &itProduct.DuplicateTemplateResult{ClientErrors: *cErrs}, nil
}
