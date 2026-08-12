package services

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// engineFor resolves another resource's engine from the registry.
//
// It is a variable rather than a plain function so that a test can supply its own engines: the
// registry is a package singleton populated during Init, which a unit test has no way to build.
var engineFor = func(schemaName string) (drif.DynamicResourceEngine, error) {
	engine, ok := dynamicresource.Registry().GetEngine(schemaName)
	if !ok {
		return nil, errors.Errorf("no resource engine for '%s'", schemaName)
	}
	return engine, nil
}

// EngineFor exposes the registry lookup to the dynamicengines package, whose action callbacks
// receive only their own engine but need another resource's repository to apply a cross-schema
// rule. It delegates to engineFor so a test's substitution is honoured here too.
func EngineFor(schemaName string) (drif.DynamicResourceEngine, error) {
	return engineFor(schemaName)
}

// maxBatchedTemplates bounds the batched template read. A page larger than this is past what a
// synchronous request should carry, and the bound keeps one oversized page from becoming an
// unbounded IN-list.
const maxBatchedTemplates = 1000

// NewProductVariantDomainService derives the variant service from the engine's default one.
//
// base is the Product Variant engine's own resource service, which this type embeds: every
// built-in action keeps running through the default implementation, and the template_* handling
// below is layered on top. The result is installed with Engine.SetResourceService.
func NewProductVariantDomainService(base drif.DynamicResourceService) *ProductVariantDomainServiceImpl {
	return &ProductVariantDomainServiceImpl{DynamicResourceService: base}
}

// ProductVariantDomainServiceImpl fills a variant's template_* virtual fields.
//
// Those fields have no database column: the values live on the owning template. This service is
// what makes them appear on a read, and it is deliberately the only place that knows a
// template_{x} maps to template.{x} -- the generic engine cannot know which edge a virtual field
// derives from.
type ProductVariantDomainServiceImpl struct {
	drif.DynamicResourceService
}

var _ drif.DynamicResourceService = (*ProductVariantDomainServiceImpl)(nil)

// SetArchived archives the variant, stamps why it was archived, and brings its template in step.
//
// Both follow-ups must happen after the variant row is written, which is why they live here rather
// than in the engine's AfterValidationSuccess hook: that hook runs before MainProcess, so the
// "are any variants left?" count would still see this variant unarchived and never archive the
// template. See BR §8.9, BR-PROD-VAR-006 and BR-PROD-VAR-007.
func (this *ProductVariantDomainServiceImpl) SetArchived(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	result, err := this.DynamicResourceService.SetArchived(ctx, params)
	if err != nil || result == nil || result.ClientErrors.Count() > 0 {
		return result, err
	}

	variant := models.NewProductVariantFrom(params)
	archived := variant.IsArchived()
	variantId := derefString(variant.GetId())
	if archived == nil || variantId == "" {
		return result, nil
	}

	// The stamp is a second write, so it supersedes the etag the archive produced. Reporting the
	// archive's stale one would have the caller's next request rejected as a concurrent
	// modification by a change this same call made.
	stampEtag, err := this.stampArchiveSource(ctx, variantId, *archived)
	if err != nil {
		return nil, err
	}
	if stampEtag != "" {
		result.Data.Etag = stampEtag
	}

	templateId, err := this.templateIdOf(ctx, variantId)
	if err != nil || templateId == "" {
		return result, err
	}
	if err := this.syncTemplateAvailability(ctx, templateId, *archived); err != nil {
		return nil, err
	}
	return result, nil
}

// stampArchiveSource records that this archive was the user's own doing, so that a later template
// unarchive restores only the variants its cascade took down and leaves this one archived.
// Unarchiving clears the stamp again.
// It returns the etag the stamp produced, which becomes the row's current one.
func (this *ProductVariantDomainServiceImpl) stampArchiveSource(
	ctx corectx.Context, variantId string, archived bool,
) (string, error) {
	update := dmodel.DynamicFields{models.ProductVariantFieldId: variantId}
	if archived {
		update[models.ProductVariantFieldArchiveSource] = models.ArchiveSourceUser.String()
	} else {
		update[models.ProductVariantFieldArchiveSource] = nil
	}

	engine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}
	result, err := engine.ResourceRepository().Update(ctx, update)
	if err != nil {
		return "", errors.Wrap(err, "stampArchiveSource")
	}
	if result == nil || !result.HasData {
		return "", nil
	}
	return string(result.Data.Etag), nil
}

// syncTemplateAvailability archives a template once its last selectable variant is gone, and
// brings it back when a variant returns. A template with nothing transactable left must not keep
// advertising itself as available.
func (this *ProductVariantDomainServiceImpl) syncTemplateAvailability(
	ctx corectx.Context, templateId string, archived bool,
) error {
	variantEngine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return err
	}

	if archived {
		remaining, err := models.FindActiveTemplateVariants(
			ctx, variantEngine.ResourceRepository(), templateId, 1)
		if err != nil {
			return errors.Wrap(err, "syncTemplateAvailability")
		}
		if len(remaining) > 0 {
			return nil
		}
	}

	templateEngine, err := engineFor(models.ProductTemplateSchemaName)
	if err != nil {
		return err
	}
	// Written through the repository rather than the template's own set_archived action: this is
	// the consequence of a cascade, and re-entering that action would run the template's cascade
	// back over the variants that triggered it.
	_, err = templateEngine.ResourceRepository().Update(ctx, dmodel.DynamicFields{
		models.ProductTemplateFieldId: templateId,
		basemodel.FieldIsArchived:     archived,
	})
	return errors.Wrap(err, "syncTemplateAvailability")
}

// templateIdOf reads the owning template from the stored row: a set_archived payload carries only
// the id and the flag.
func (this *ProductVariantDomainServiceImpl) templateIdOf(
	ctx corectx.Context, variantId string,
) (string, error) {
	engine, err := engineFor(models.ProductVariantSchemaName)
	if err != nil {
		return "", err
	}

	found, err := engine.ResourceRepository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.ProductVariantFieldId: variantId},
		Fields: []string{models.ProductVariantFieldId, models.ProductVariantFieldProductTemplateId},
	})
	if err != nil {
		return "", errors.Wrap(err, "templateIdOf")
	}
	if !found.HasData {
		return "", nil
	}
	return derefString(models.NewProductVariantFrom(found.Data).GetProductTemplateId()), nil
}

// GetById fetches one variant and fills its template_* fields from the hydrated template edge.
//
// One record means one extra query for the edge, which is what the edge hydration already costs
// and is acceptable here. Search deliberately does not use this path -- see Search below.
func (this *ProductVariantDomainServiceImpl) GetById(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneWithTemplate(ctx, params, this.DynamicResourceService.GetById, "GetById")
}

// GetOne fetches one variant by any unique key, filling template_* the same way GetById does.
func (this *ProductVariantDomainServiceImpl) GetOne(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	return this.getOneWithTemplate(ctx, params, this.DynamicResourceService.GetOne, "GetOne")
}

type getOneFn func(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error)

func (this *ProductVariantDomainServiceImpl) getOneWithTemplate(
	ctx corectx.Context, params dmodel.DynamicFields, delegate getOneFn, action string,
) (*dyn.OpResult[dyn.SingleResultData[dmodel.DynamicFields]], error) {
	requested := readRequestedFields(params)
	wantsTemplate := len(requested) == 0 || namesAnyTemplateField(requested)

	if wantsTemplate && len(requested) > 0 {
		// The edge carries the values; product_template_id is what links a row back to it.
		params[paramFieldNames] = appendMissing(requested,
			models.ProductVariantEdgeTemplate, models.ProductVariantFieldProductTemplateId)
	}

	result, err := delegate(ctx, params)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	if !wantsTemplate {
		return result, nil
	}

	// With no projection the delegate returns every column but no edge — asking for the edge
	// would have turned the whole-record read into a two-field one. The template is read
	// separately instead, which is the same one extra query the edge would have cost.
	if len(requested) == 0 {
		err := this.fillTemplateFields(ctx, []dmodel.DynamicFields{result.Data.Item}, models.TemplateVirtualFields)
		return result, errors.Wrap(err, "ProductVariantDomainService."+action)
	}

	if err := fillFromHydratedEdge(result.Data.Item); err != nil {
		return nil, errors.Wrap(err, "ProductVariantDomainService."+action)
	}
	return result, nil
}

// Search fills template_* across a page with ONE batched template read, not one per row.
//
// The repository can hydrate the template edge, but it does so per row: a 50-variant page would
// cost 51 queries. Reading the distinct templates in a single follow-up keeps a page at two
// queries however large it is, which is the whole reason this override exists.
func (this *ProductVariantDomainServiceImpl) Search(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dyn.PagedResultData[dmodel.DynamicFields]], error) {
	requested := readRequestedFields(params)
	wantedVirtuals := templateFieldsIn(requested)

	// Only when the caller named fields: an empty selection means "the resource's default field
	// set", and narrowing it to the join key alone would drop every real column from the page.
	if len(wantedVirtuals) > 0 && len(requested) > 0 {
		// product_template_id is the join key the fill needs; without it the rows cannot be
		// matched back to the templates they came from.
		params[paramFieldNames] = appendMissing(requested, models.ProductVariantFieldProductTemplateId)
	}

	result, err := this.DynamicResourceService.Search(ctx, params)
	if err != nil || result == nil || !result.HasData {
		return result, err
	}
	// No template_* was asked for, so a search costs exactly what it did before.
	if len(wantedVirtuals) == 0 || len(result.Data.Items) == 0 {
		return result, nil
	}

	if err := this.fillTemplateFields(ctx, result.Data.Items, wantedVirtuals); err != nil {
		return nil, errors.Wrap(err, "ProductVariantDomainService.Search")
	}
	return result, nil
}

// fillTemplateFields reads every template the page refers to in one query, then copies the
// wanted fields onto each variant row.
func (this *ProductVariantDomainServiceImpl) fillTemplateFields(
	ctx corectx.Context, rows []dmodel.DynamicFields, wantedVirtuals []string,
) error {
	templateIds := distinctTemplateIds(rows)
	if len(templateIds) == 0 {
		return nil
	}

	templates, err := this.readTemplates(ctx, templateIds, wantedVirtuals)
	if err != nil {
		return err
	}

	for _, row := range rows {
		templateId := readStringField(row, models.ProductVariantFieldProductTemplateId)
		template, ok := templates[templateId]
		if !ok {
			// A variant whose template is missing reads as "unknown" rather than as a product
			// with an empty name, so its template_* fields are left absent.
			continue
		}
		variant := models.NewProductVariantFrom(row)
		variant.FillFromTemplate(template)
	}
	return nil
}

// readTemplates fetches the named templates in a single query, selecting only the columns the
// requested virtual fields actually need.
func (this *ProductVariantDomainServiceImpl) readTemplates(
	ctx corectx.Context, templateIds []string, wantedVirtuals []string,
) (map[string]*models.ProductTemplate, error) {
	engine, err := engineFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}

	graph := dmodel.NewSearchGraph()
	graph.NewCondition(models.ProductTemplateFieldId, dmodel.In, toAnySlice(templateIds)...)

	found, err := engine.ResourceRepository().Search(ctx, dyn.RepoSearchParam{
		Fields: templateSourceColumns(wantedVirtuals),
		Graph:  graph,
		Page:   0,
		Size:   len(templateIds),
	})
	if err != nil {
		return nil, errors.Wrap(err, "readTemplates")
	}
	if found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "readTemplates")
	}
	if !found.HasData {
		return map[string]*models.ProductTemplate{}, nil
	}

	byId := make(map[string]*models.ProductTemplate, len(found.Data.Items))
	for _, item := range found.Data.Items {
		template := models.NewProductTemplateFrom(item)
		if id := template.GetId(); id != nil {
			byId[string(*id)] = template
		}
	}
	return byId, nil
}

// fillFromHydratedEdge copies the template edge the repository already fetched onto the row's
// template_* fields. Used by the single-record reads, where the edge is cheap.
func fillFromHydratedEdge(row dmodel.DynamicFields) error {
	raw, ok := row[models.ProductVariantEdgeTemplate]
	if !ok || raw == nil {
		return nil
	}
	edge, ok := raw.(dmodel.DynamicFields)
	if !ok {
		return errors.Errorf(
			"fillFromHydratedEdge: edge %q has unexpected type %T",
			models.ProductVariantEdgeTemplate, raw)
	}

	variant := models.NewProductVariantFrom(row)
	variant.FillFromTemplate(models.NewProductTemplateFrom(edge))
	return nil
}

const paramFieldNames = "fields"

// RewriteTemplateFieldPath maps a variant's virtual field to the edge path the SQL layer can
// resolve: template_name -> template.name. A caller filtering or sorting on a virtual field goes
// through here, because the field itself has no column to point at.
func RewriteTemplateFieldPath(field string) (string, bool) {
	source, ok := models.TemplateSourceField[field]
	if !ok {
		return field, false
	}
	return models.ProductVariantEdgeTemplate + "." + source, true
}

func readRequestedFields(params dmodel.DynamicFields) []string {
	raw, ok := params[paramFieldNames]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return typed
	case []any:
		fields := make([]string, 0, len(typed))
		for _, item := range typed {
			if str, ok := item.(string); ok {
				fields = append(fields, str)
			}
		}
		return fields
	default:
		return nil
	}
}

// templateFieldsIn returns the requested virtual fields. An empty request means "everything",
// which includes them.
func templateFieldsIn(requested []string) []string {
	if len(requested) == 0 {
		return models.TemplateVirtualFields
	}
	wanted := make([]string, 0, len(requested))
	for _, field := range requested {
		if _, ok := models.TemplateSourceField[field]; ok {
			wanted = append(wanted, field)
		}
	}
	return wanted
}

func namesAnyTemplateField(requested []string) bool {
	return len(templateFieldsIn(requested)) > 0
}

// templateSourceColumns is the template-side column list needed to fill the wanted virtual
// fields, always including the id the rows are keyed by.
func templateSourceColumns(wantedVirtuals []string) []string {
	columns := []string{models.ProductTemplateFieldId}
	for _, virtual := range wantedVirtuals {
		if source, ok := models.TemplateSourceField[virtual]; ok {
			columns = appendMissing(columns, source)
		}
	}
	return columns
}

func distinctTemplateIds(rows []dmodel.DynamicFields) []string {
	seen := make(map[string]bool, len(rows))
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		id := readStringField(row, models.ProductVariantFieldProductTemplateId)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		if len(ids) == maxBatchedTemplates {
			break
		}
	}
	return ids
}

func readStringField(row dmodel.DynamicFields, field string) string {
	val, ok := row[field]
	if !ok || val == nil {
		return ""
	}
	// model.Id is a string type, so the string case covers it too.
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func appendMissing(fields []string, additions ...string) []string {
	present := make(map[string]bool, len(fields))
	for _, field := range fields {
		present[strings.TrimSpace(field)] = true
	}
	out := fields
	for _, addition := range additions {
		if !present[addition] {
			out = append(out, addition)
			present[addition] = true
		}
	}
	return out
}

func toAnySlice(values []string) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
