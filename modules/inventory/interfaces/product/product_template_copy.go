package product

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The copy action: `POST product_template/:id/copy` with {org_id, name?}.
type (
	CopyTemplateCommand = dmodel.DynamicFields
	CopyTemplateResult  = dyn.OpResult[dmodel.DynamicFields]
)

// CopyTemplateFieldName optionally names the copy, as a langjson; absent, the source name gets a
// copy suffix.
const CopyTemplateFieldName = "name"

// CopyTemplateQuery asks for the create payload of a template's copy.
type CopyTemplateQuery struct {
	TemplateId string
	Name       any
}

// TemplateCopier builds the create payload of a template's copy, configuration included, without
// writing; the application service creates it through the ordinary template create.
type TemplateCopier interface {
	BuildTemplateCopy(ctx corectx.Context, query CopyTemplateQuery) (*dyn.OpResult[dmodel.DynamicFields], error)
}
