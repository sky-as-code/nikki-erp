package cqrs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	salessettings "github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/settings"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

type stubOrgSettings struct {
	values map[string]any
	err    error
}

func (this stubOrgSettings) GetOrgSettingValues(corectx.Context, model.Id, string) (map[string]any, error) {
	return this.values, this.err
}

func sequenceRef(id, orgId string) usagecheck.ResourceRef {
	return usagecheck.ResourceRef{
		ResourceName: usagecheck.ResourceSequenceDefinition, Identifier: usagecheck.NewIdentifier(id, orgId),
	}
}

func TestSequenceUsedWhileASettingHoldsIt(t *testing.T) {
	reader := stubOrgSettings{values: map[string]any{salessettings.OrgSettingBillNumberSequenceId: "DEF1"}}
	ctx := corectx.NewRequestContext(context.Background())

	used, usedBy, err := sequenceUsedBySettings(ctx, reader, sequenceRef("DEF1", "ORG1"))
	require.NoError(t, err)
	require.True(t, used)
	require.Equal(t, "sales."+salessettings.OrgSettingBillNumberSequenceId, usedBy)

	used, _, err = sequenceUsedBySettings(ctx, reader, sequenceRef("DEF2", "ORG1"))
	require.NoError(t, err)
	require.False(t, used)
}

// A failed read must refuse the archive, never permit it.
func TestSequenceUsageFailsWhenSettingsAreUnreadable(t *testing.T) {
	reader := stubOrgSettings{err: errors.New("down")}

	_, _, err := sequenceUsedBySettings(corectx.NewRequestContext(context.Background()), reader, sequenceRef("DEF1", "ORG1"))
	require.Error(t, err)
}

// Sales registers org-scoped definitions only, so a tenant-scoped one is none of its business.
func TestSequenceWithoutOrgIsNotSales(t *testing.T) {
	reader := stubOrgSettings{err: errors.New("must not be read")}

	used, _, err := sequenceUsedBySettings(corectx.NewRequestContext(context.Background()), reader, sequenceRef("DEF1", ""))
	require.NoError(t, err)
	require.False(t, used)
}
