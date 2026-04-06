package crud

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

func Validate[T dmodel.SchemaGetter](cmdQuery T) (*T, ft.ClientErrors) {
	schema := cmdQuery.GetSchema()
	sanitized, cErrs := schema.ValidateStruct(cmdQuery)
	if cErrs.Count() > 0 {
		return nil, cErrs
	}

	return sanitized.(*T), nil
}
