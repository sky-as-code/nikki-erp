package services

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
)

// AssertOrgDeletable asks every module holding an org_id whether the organization is still in use.
// The identifier carries no org scope: the org is the scope, so a row of any org naming this id is
// exactly the reference being looked for.
func AssertOrgDeletable(
	ctx corectx.Context, dispatcher *usagecheck.Dispatcher, orgId string, vErrs *ft.ClientErrors,
) error {
	if orgId == "" {
		return nil
	}
	return usagecheck.AssertDeletable(ctx, dispatcher, c.IamModuleName, basemodel.FieldId,
		usagecheck.ResourceRef{
			ResourceName: usagecheck.ResourceOrg,
			Identifier:   usagecheck.NewIdentifier(orgId, ""),
		}, vErrs)
}

// NewOrgEngineDeleteGuard wraps the iam_org engine's resource service, so the generic
// DELETE /v1/iam/iam_org route runs the same usage check as DeleteOrg. Install with
// Engine.SetResourceService.
func NewOrgEngineDeleteGuard(
	base drif.DynamicResourceService, dispatcher *usagecheck.Dispatcher,
) drif.DynamicResourceService {
	return &OrgEngineDeleteGuard{DynamicResourceService: base, dispatcher: dispatcher}
}

type OrgEngineDeleteGuard struct {
	drif.DynamicResourceService
	dispatcher *usagecheck.Dispatcher
}

func (this *OrgEngineDeleteGuard) Delete(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	vErrs := ft.NewClientErrors()
	orgId, _ := params[basemodel.FieldId].(string)
	if err := AssertOrgDeletable(ctx, this.dispatcher, orgId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.DynamicResourceService.Delete(ctx, params)
}
