package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
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
