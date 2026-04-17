package basemodel

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/model"
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

func (this DynamicModelBase) GetId() *model.Id {
	return this.GetFieldData().GetModelId(FieldId)
}

func (this *DynamicModelBase) SetId(v *model.Id) {
	this.GetFieldData().SetModelId(FieldId, v)
}

func (this DynamicModelBase) IsArchived() *bool {
	return this.GetFieldData().GetBool(FieldIsArchived)
}

func (this DynamicModelBase) MustIsArchived() bool {
	b := this.GetFieldData().GetBool(FieldIsArchived)
	if b == nil {
		panic(errors.New("is_archived is nil"))
	}
	return *b
}

func (this *DynamicModelBase) SetIsArchived(v *bool) {
	this.GetFieldData().SetBool(FieldIsArchived, v)
}

func (this DynamicModelBase) GetEtag() *model.Etag {
	return this.GetFieldData().GetEtag(FieldEtag)
}

func (this *DynamicModelBase) SetEtag(v *model.Etag) {
	this.GetFieldData().SetEtag(FieldEtag, v)
}

func (this DynamicModelBase) GetOrgId() *model.Id {
	return this.GetFieldData().GetModelId(FieldOrgId)
}

func (this *DynamicModelBase) SetOrgId(v *model.Id) {
	this.GetFieldData().SetModelId(FieldOrgId, v)
}

func (this DynamicModelBase) MarshalJSON() ([]byte, error) {
	if this.fields == nil {
		return []byte("{}"), nil
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
