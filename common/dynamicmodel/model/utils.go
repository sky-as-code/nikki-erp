package model

import "github.com/sky-as-code/nikki-erp/common/array"

const JsonFieldTag = "json"

type DynamicModel interface {
	DynamicModelGetter
	DynamicModelSetter
}

type DynamicModelGetter interface {
	GetFieldData() DynamicFields
}

type DynamicModelSetter interface {
	SetFieldData(data DynamicFields)
}

type SchemaGetter interface {
	GetFieldData() DynamicFields
	GetSchema() *ModelSchema
}

func ExtractFieldsArr[TSrc DynamicModelGetter](arr []TSrc) []DynamicFields {
	return array.Map(arr, func(item TSrc) DynamicFields {
		return item.GetFieldData()
	})
}
