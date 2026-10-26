package product

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// The duplicate action: `POST product_template/:id/duplicate` with {org_id, name?}.
type (
	DuplicateTemplateCommand = dmodel.DynamicFields
	DuplicateTemplateResult  = dyn.OpResult[dmodel.DynamicFields]
)

// DuplicateTemplateFieldName optionally names the duplicate, as a langjson; absent, the source name
// gets a copy suffix.
const DuplicateTemplateFieldName = "name"

// DuplicateTemplateQuery asks for the create payload of a template's duplicate.
type DuplicateTemplateQuery struct {
	TemplateId string
	Name       any
}

// TemplateDuplicator builds the create payload of a template's duplicate, configuration included;
// the application service creates it through the ordinary template create. The only thing it
// writes is the clone of the preview image, a pending file that expires if the create fails.
type TemplateDuplicator interface {
	BuildTemplateDuplicate(ctx corectx.Context, query DuplicateTemplateQuery) (*dyn.OpResult[dmodel.DynamicFields], error)
}
