package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// The documented wire form is a string, which keeps the precision BR-UOM-ESS-018 needs; a JSON
// number is still accepted, as the old binder did.
func TestReadQuantityAcceptsStringAndNumber(t *testing.T) {
	quantity, cErrs := readQuantity(dmodel.DynamicFields{fieldQuantity: "0.9144"})
	require.Nil(t, cErrs)
	assert.Equal(t, "0.9144", quantity.String())

	quantity, cErrs = readQuantity(dmodel.DynamicFields{fieldQuantity: float64(2)})
	require.Nil(t, cErrs)
	assert.Equal(t, "2", quantity.String())
}

func TestReadQuantityRefusesGarbage(t *testing.T) {
	_, cErrs := readQuantity(dmodel.DynamicFields{fieldQuantity: "two"})
	require.NotNil(t, cErrs)
	assert.Equal(t, fieldQuantity, (*cErrs)[0].Field)
}
