package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

func TestAssertRoleNotSystemManaged(t *testing.T) {
	cases := map[string]struct {
		managedBy any
		refused   bool
	}{
		"system-managed is refused": {"system", true},
		"module-managed is allowed": {"module", false},
		"unmanaged is allowed":      {nil, false},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			found := composable.NewDynamicEntityFrom(dmodel.DynamicFields{
				basemodel.FieldManagedByType: testCase.managedBy,
			})
			vErrs := ft.ClientErrors{}

			require.NoError(t, assertRoleNotSystemManaged(nil, nil, found, &vErrs))
			assert.Equal(t, testCase.refused, vErrs.Count() > 0)
		})
	}
}

func TestAssertRoleCreateNotSystemManaged(t *testing.T) {
	for managedBy, refused := range map[string]bool{"system": true, "module": false, "user": false} {
		input := composable.NewDynamicEntityFrom(dmodel.DynamicFields{basemodel.FieldManagedByType: managedBy})
		vErrs := ft.ClientErrors{}

		require.NoError(t, assertRoleCreateNotSystemManaged(nil, input, nil, &vErrs))
		assert.Equal(t, refused, vErrs.Count() > 0, managedBy)
	}
}
