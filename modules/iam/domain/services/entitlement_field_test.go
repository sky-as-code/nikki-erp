package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	c "github.com/sky-as-code/nikki-erp/modules/iam/constants"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func entitlementIn(scope c.ResourceScope, orgId *model.Id, fieldName *string) *domain.Entitlement {
	ent := domain.NewEntitlement()
	ent.SetScope(&scope)
	ent.SetOrgId(orgId)
	ent.GetFieldData().SetString(domain.EntitlementFieldFieldName, fieldName)
	return ent
}

func expressionOf(ent *domain.Entitlement, action *string, resource *string) string {
	ent.CalculateExpression(action, resource)
	applyFieldName(ent, action, resource)
	return *ent.GetFieldData().GetString(domain.EntitlementFieldExpression)
}

func TestEntitlementExpressionCarriesTheFieldName(t *testing.T) {
	ent := entitlementIn(c.ResourceScopeTenant, nil, util.ToPtr("display_name"))
	assert.Equal(t, "update:iam_user.display_name:tenant",
		expressionOf(ent, util.ToPtr("update"), util.ToPtr("iam_user")))

	orgId := model.Id("01JQZ0X0000000000000000001")
	ent = entitlementIn(c.ResourceScopeOrg, &orgId, util.ToPtr("email"))
	assert.Equal(t, "*:iam_user.email:org/01JQZ0X0000000000000000001", expressionOf(ent, nil, util.ToPtr("iam_user")))
}

func TestEntitlementExpressionWithoutFieldIsResourceLevel(t *testing.T) {
	ent := entitlementIn(c.ResourceScopeTenant, nil, nil)
	assert.Equal(t, "read:iam_user:tenant", expressionOf(ent, util.ToPtr("read"), util.ToPtr("iam_user")))
}
