package baserepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// existsStub answers Exists the way the tenant wrapper does: it merges its own key into the maps
// it receives, which is why KeyExist must hand it a copy.
type existsStub struct {
	dyn.BaseDynamicRepository
	found bool
}

func (this *existsStub) Exists(
	_ corectx.Context, keys []dmodel.DynamicFields,
) (*dyn.OpResult[dyn.RepoExistsResult], error) {
	keys[0]["tenant_id"] = "t1"
	result := dyn.RepoExistsResult{}
	if this.found {
		result.Existing = keys
	} else {
		result.NotExisting = keys
	}
	return &dyn.OpResult[dyn.RepoExistsResult]{Data: result, HasData: true}, nil
}

func TestKeyExist_ReportsAMatchingRow(t *testing.T) {
	result, err := KeyExist(nil, &existsStub{found: true}, dmodel.DynamicFields{"id": "r1", "org_id": "o1"})

	require.NoError(t, err)
	assert.True(t, result.Data)
}

func TestKeyExist_ReportsAMissingRow(t *testing.T) {
	result, err := KeyExist(nil, &existsStub{found: false}, dmodel.DynamicFields{"id": "r1", "org_id": "o1"})

	require.NoError(t, err)
	assert.True(t, result.HasData)
	assert.False(t, result.Data)
}

func TestKeyExist_LeavesTheCallersKeysUntouched(t *testing.T) {
	keys := dmodel.DynamicFields{"id": "r1", "org_id": "o1"}

	_, err := KeyExist(nil, &existsStub{found: true}, keys)

	require.NoError(t, err)
	assert.NotContains(t, keys, "tenant_id")
}
