package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

const templateValueTestOrg = model.Id("01ORG00000000000000000001")

// templateValueFakeBase is the composable default reduced to what Update touches. The embedded nil
// interface panics on anything else, which is the point: Update must not reach the default Update.
type templateValueFakeBase struct {
	composable.CrudApplicationService
	domSvc       *templateValueFakeDomain
	recordChecks int
}

func (this *templateValueFakeBase) ResolveOrgScope(
	corectx.Context, dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	orgId := templateValueTestOrg
	return &orgId, nil
}

func (this *templateValueFakeBase) AssertRecordInOrg(
	corectx.Context, dmodel.DynamicFields, model.Id,
) (*ft.ClientErrors, error) {
	this.recordChecks++
	return nil, nil
}

func (this *templateValueFakeBase) DomainService() composable.CrudDomainService { return this.domSvc }

type templateValueFakeDomain struct {
	composable.CrudDomainService
	updates int
}

func (this *templateValueFakeDomain) Update(
	corectx.Context, composable.UpdateCommand, ...composable.UpdateOptions,
) (*composable.MutateResult, error) {
	this.updates++
	return &composable.MutateResult{}, nil
}

func templateValueCaller(entitlement string) corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		Principal:    corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("01USER0000000000000000001")},
		UserOrgIds:   ds.NewSetFrom(templateValueTestOrg),
		Entitlements: ds.NewSetFrom(entitlement),
	})
	return ctx
}

func orgExpression(action, resource string) string {
	orgId := templateValueTestOrg
	return requestguard.BuildExpression(action, resource, requestguard.ResourceScopeOrg, &orgId)
}

func TestTemplateValueUpdateAssertsUpdateValueOnTheAttribute(t *testing.T) {
	assert.Equal(t, "update_value", templateValueUpdatePermission.ActionCode)
	assert.Equal(t, "inventory_product_attribute", templateValueUpdatePermission.ResourceCode)
}

func TestTemplateValueUpdatePassesWithUpdateValue(t *testing.T) {
	base := &templateValueFakeBase{domSvc: &templateValueFakeDomain{}}
	svc := NewProductTemplateAttributeValueApplicationService(base)

	res, err := svc.Update(templateValueCaller(orgExpression("update_value", "inventory_product_attribute")),
		composable.UpdateCommand{"id": "01VALUE", "org_id": string(templateValueTestOrg)})

	require.NoError(t, err)
	assert.Zero(t, res.ClientErrors.Count())
	assert.Equal(t, 1, base.recordChecks, "the record must still be confined to the org")
	assert.Equal(t, 1, base.domSvc.updates)
}

// The default permission of the serving resource no longer suffices.
func TestTemplateValueUpdateRefusesTheResourceOwnUpdate(t *testing.T) {
	base := &templateValueFakeBase{domSvc: &templateValueFakeDomain{}}
	svc := NewProductTemplateAttributeValueApplicationService(base)

	res, err := svc.Update(
		templateValueCaller(orgExpression("update", "inventory_product_template_attribute_value")),
		composable.UpdateCommand{"id": "01VALUE", "org_id": string(templateValueTestOrg)})

	require.NoError(t, err)
	assert.NotZero(t, res.ClientErrors.Count())
	assert.Zero(t, base.domSvc.updates)
}
