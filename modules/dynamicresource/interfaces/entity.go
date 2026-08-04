package interfaces

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// DynamicEntity is the schema-agnostic domain model of the dynamic resource engine.
// It carries nothing but the field map, which makes it usable as the TDomain type
// argument of the generic helpers in modules/core/dynamicmodel/crud and .../baserepo,
// no matter which schema the enclosing engine serves.
type DynamicEntity struct {
	fields dmodel.DynamicFields
}

func NewDynamicEntity() *DynamicEntity {
	return &DynamicEntity{fields: dmodel.DynamicFields{}}
}

func NewDynamicEntityFrom(fields dmodel.DynamicFields) *DynamicEntity {
	if fields == nil {
		fields = dmodel.DynamicFields{}
	}
	return &DynamicEntity{fields: fields}
}

// GetFieldData implements dmodel.DynamicModelGetter.
func (this *DynamicEntity) GetFieldData() dmodel.DynamicFields {
	if this.fields == nil {
		this.fields = dmodel.DynamicFields{}
	}
	return this.fields
}

// SetFieldData implements dmodel.DynamicModelSetter.
func (this *DynamicEntity) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}
