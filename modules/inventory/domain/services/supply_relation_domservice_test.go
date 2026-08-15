package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// A relation declares who may restock whom. Both endpoints are required and they must differ:
// a warehouse supplying itself describes no route at all.
func TestSupplyRelationEndpointRules(t *testing.T) {
	service := &SupplyRelationDomainServiceImpl{}

	missing, err := service.assertRelationValid(nil, dmodel.DynamicFields{
		models.WarehouseSupplyRelationFieldSourceWarehouseId: "wh-1",
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, missing.Count(), "a relation with no destination is incomplete")

	selfSupply, err := service.assertRelationValid(nil, dmodel.DynamicFields{
		models.WarehouseSupplyRelationFieldSourceWarehouseId:      "wh-1",
		models.WarehouseSupplyRelationFieldDestinationWarehouseId: "wh-1",
	}, "")
	assert.NoError(t, err)
	assert.Equal(t, 1, selfSupply.Count(), "a warehouse cannot supply itself")
}

func TestIsArchivedWarehouse(t *testing.T) {
	archived := models.NewWarehouseFrom(dmodel.DynamicFields{"is_archived": true})
	assert.True(t, isArchivedWarehouse(*archived))

	live := models.NewWarehouseFrom(dmodel.DynamicFields{"is_archived": false})
	assert.False(t, isArchivedWarehouse(*live))

	unset := models.NewWarehouseFrom(dmodel.DynamicFields{})
	assert.False(t, isArchivedWarehouse(*unset), "an absent flag is not archived")
}

// The assertion that only Warehouse and Inventory Location carry a status lives in the models
// package, where the base mixins a schema build needs are already registered. See
// warehouse_schema_test.go.
