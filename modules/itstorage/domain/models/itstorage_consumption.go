package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	c "github.com/sky-as-code/nikki-erp/modules/itstorage/constants"
)

const (
	ConsumptionSchemaName = c.ResourceItStorageConsumption

	ConsumptionFieldId            = basemodel.FieldId
	ConsumptionFieldEtag          = basemodel.FieldEtag
	ConsumptionFieldCreatedAt     = basemodel.FieldCreatedAt
	ConsumptionFieldOwnerType     = "owner_type"
	ConsumptionFieldOwnerId       = "owner_id"
	ConsumptionFieldUsedBytes     = "used_bytes"
	ConsumptionFieldReservedBytes = "reserved_bytes"
)

//go:embed itstorage_consumption.json
var consumptionSchemaJson string

func ConsumptionSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(consumptionSchemaJson)
}

// ConsumptionOwner names one row of itstorage_consumption. InitialUsedBytes seeds used_bytes
// when the row does not exist yet, so an owner whose files predate the table starts from what
// its space already holds instead of from zero.
type ConsumptionOwner struct {
	Type             string
	Id               model.Id
	InitialUsedBytes int64
}
