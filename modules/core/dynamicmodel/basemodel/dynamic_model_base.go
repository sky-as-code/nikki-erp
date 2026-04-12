package basemodel

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/json"
)

func NewDynamicModel(fields ...dmodel.DynamicFields) DynamicModelBase {
	var f dmodel.DynamicFields
	if len(fields) == 0 {
		f = make(dmodel.DynamicFields)
	} else {
		f = fields[0]
	}
	return DynamicModelBase{fields: f}
}

// Embed this struct to your model entity to make it serializable to JSON.
type DynamicModelBase struct {
	fields dmodel.DynamicFields
}

func (this DynamicModelBase) GetFieldData() dmodel.DynamicFields {
	return this.fields
}

func (this *DynamicModelBase) SetFieldData(data dmodel.DynamicFields) {
	this.fields = data
}

func (this DynamicModelBase) MarshalJSON() ([]byte, error) {
	if this.fields == nil {
		return nil, nil
	}
	return json.Marshal(this.fields)
}

func (this *DynamicModelBase) UnmarshalJSON(data []byte) error {
	var raw dmodel.DynamicFields
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil {
		raw = make(dmodel.DynamicFields)
	}
	this.fields = raw
	return nil
}
