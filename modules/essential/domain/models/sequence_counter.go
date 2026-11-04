package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SequenceCounterSchemaName = "essential_sequence_counter"

	SeqCtrFieldId              = basemodel.FieldId
	SeqCtrFieldOrgId           = basemodel.FieldOrgId
	SeqCtrFieldDefinitionId    = "definition_id"
	SeqCtrFieldScopeObjectId   = "scope_object_id"
	SeqCtrFieldScopeKey        = "scope_key"
	SeqCtrFieldPeriodKey       = "period_key"
	SeqCtrFieldLastNumber      = "last_number"
	SeqCtrFieldAlgorithmParams = "algorithm_params"
	SeqCtrFieldCreatedAt       = basemodel.FieldCreatedAt
	SeqCtrFieldUpdatedAt       = basemodel.FieldUpdatedAt
	SeqCtrFieldCreatedBy       = "created_by"
)

//go:embed sequence_counter.json
var sequenceCounterSchemaJson string

func SequenceCounterSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(sequenceCounterSchemaJson)
}
