package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itExt "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/external"
)

type stubUomLookup struct {
	result *itExt.GetUomResult
}

func (this *stubUomLookup) GetUom(corectx.Context, itExt.GetUomQuery) (*itExt.GetUomResult, error) {
	return this.result, nil
}

func uomFound(archived bool) *stubUomLookup {
	return &stubUomLookup{result: &itExt.GetUomResult{
		Data: itExt.GetUomResultData{Id: "01UOM", IsArchived: archived}, HasData: true,
	}}
}

// The answer comes from Essential through the port. Reading Essential's table through Inventory's
// own resource hub always failed, and the fallback then passed every unit, archived or not.
func TestIsUomUsableAsksEssential(t *testing.T) {
	cases := []struct {
		name   string
		uoms   itExt.UomLookupExtService
		usable bool
	}{
		{"a live unit can be chosen", uomFound(false), true},
		{"an archived unit cannot", uomFound(true), false},
		{"a unit Essential does not know cannot", &stubUomLookup{result: &itExt.GetUomResult{}}, false},
		{"no UoM module bound validates nothing", nil, true},
	}
	for _, tc := range cases {
		usable, err := IsUomUsable(callerContext(), tc.uoms, "01UOM")
		require.NoError(t, err, tc.name)
		assert.Equal(t, tc.usable, usable, tc.name)
	}
}
