package cqrs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
	modconstants "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	pinvsettings "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/settings"
	itExt "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/external"
)

type heldSettings map[string]any

func (this heldSettings) GetOrgSettingValues(corectx.Context, model.Id, string) (map[string]any, error) {
	return this, nil
}

func checkerHolding(held string) usagecheck.ResourceUsageChecker {
	return SequenceUsageChecker(func() (itExt.OrgSettingsReaderExtService, error) {
		return heldSettings{pinvsettings.OrgSettingInvoiceNumberSequenceId: held}, nil
	})
}

func sequenceRef(id, orgId string) usagecheck.ResourceRef {
	return usagecheck.ResourceRef{
		ResourceName: usagecheck.ResourceSequenceDefinition,
		Identifier:   map[string]string{usagecheck.IdentifierKeyId: id, usagecheck.IdentifierKeyOrgId: orgId},
	}
}

func TestTheHeldInvoiceSeriesIsInUse(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())

	used, usedBy, err := checkerHolding("01DEF").IsUsed(ctx, sequenceRef("01DEF", "01ORG"))

	require.NoError(t, err)
	assert.True(t, used)
	assert.Equal(t, modconstants.PaymentInvoiceModuleName+"."+pinvsettings.OrgSettingInvoiceNumberSequenceId, usedBy)
}

func TestAnotherOrTenantDefinitionIsNotInUse(t *testing.T) {
	ctx := corectx.NewRequestContext(t.Context())

	used, _, err := checkerHolding("01DEF").IsUsed(ctx, sequenceRef("01OTHER", "01ORG"))
	require.NoError(t, err)
	assert.False(t, used)

	used, _, err = checkerHolding("01DEF").IsUsed(ctx, sequenceRef("01DEF", ""))
	require.NoError(t, err)
	assert.False(t, used)
}

// Essential asks every dependant about each resource it guards; a missing checker blocks deletes.
func TestTheRegistryAnswersEveryEssentialResource(t *testing.T) {
	registry, err := newUsageCheckRegistry()
	require.NoError(t, err)

	for _, resource := range []string{usagecheck.ResourceUom, usagecheck.ResourceSequenceDefinition} {
		_, found := registry.Checker("essential", resource)
		assert.Truef(t, found, "no checker for %s", resource)
	}
}
