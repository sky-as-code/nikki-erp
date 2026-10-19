package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	domain "github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
)

func TestAsProvisioning_IgnoresCallerStatus(t *testing.T) {
	cmd := asProvisioning(dmodel.DynamicFields{
		domain.OrgFieldSlug:          "acme",
		domain.OrgFieldStatus:        string(domain.OrgStatusReady),
		domain.OrgFieldStatusMessage: "spoofed",
	})

	assert.Equal(t, string(domain.OrgStatusProvisioning), cmd[domain.OrgFieldStatus])
	assert.NotContains(t, cmd, domain.OrgFieldStatusMessage)
	assert.Equal(t, "acme", cmd[domain.OrgFieldSlug])
}

func TestDeleteOrg_RefusedUntilMarkedDeleted(t *testing.T) {
	base := newFakeCrudDomain()
	svc := NewOrgCrudDomainService(base, nil, orgDispatcher(t, &orgUsageBus{}))

	result, err := svc.Delete(corectx.NewRequestContext(context.Background()), dmodel.DynamicFields{
		basemodel.FieldId:     "01JORG0000000000000000000A",
		domain.OrgFieldStatus: string(domain.OrgStatusReady),
	})

	require.NoError(t, err)
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, ErrKeyOrgNotDeleted, result.ClientErrors[0].Key)
	assert.Equal(t, 0, base.deleted)
}

func TestAssertOnlyStatusIfDeleted(t *testing.T) {
	deleted := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		domain.OrgFieldStatus: string(domain.OrgStatusDeleted),
	})
	ready := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		domain.OrgFieldStatus: string(domain.OrgStatusReady),
	})
	rename := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		basemodel.FieldId: "ORG", basemodel.FieldEtag: "1", domain.OrgFieldSlug: "renamed",
	})
	statusOnly := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		basemodel.FieldId: "ORG", basemodel.FieldEtag: "1", domain.OrgFieldStatusMessage: "iam:x",
	})
	ctx := corectx.NewRequestContext(context.Background())

	cases := []struct {
		input    *composable.DynamicEntity
		found    *composable.DynamicEntity
		refusals int
	}{
		{rename, deleted, 1},
		{statusOnly, deleted, 0},
		{rename, ready, 0},
	}
	for _, tc := range cases {
		vErrs := ft.ClientErrors{}
		require.NoError(t, assertOnlyStatusIfDeleted(ctx, tc.input, tc.found, &vErrs))
		assert.Equal(t, tc.refusals, vErrs.Count())
	}
}
