package services

import (
	"fmt"
	"reflect"
	"sort"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/codify"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductCategoryDomainService attaches the category-tree invariants to create and update:
// one root per org that never changes, a parent on every other category (the root when none is
// given), no cycles, and a canonical name kept in step with the tree. It also derives a missing
// code from the name on create. The CRUD processing is the default's; only the rules the schema
// cannot express belong here.
func NewProductCategoryDomainService(base composable.CrudDomainService) itProduct.ProductCategoryDomainService {
	return &ProductCategoryDomainServiceImpl{CrudDomainService: base}
}

type ProductCategoryDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *ProductCategoryDomainServiceImpl) Create(
	ctx corectx.Context, cmd itProduct.CreateProductCategoryCommand, options ...composable.CreateOptions,
) (*itProduct.CreateProductCategoryResult, error) {
	vErrs := ft.NewClientErrors()
	rejectServerManagedFields(cmd, vErrs)
	if vErrs.Count() > 0 {
		return &itProduct.CreateProductCategoryResult{ClientErrors: *vErrs}, nil
	}
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.BeforeValidation = chainBeforeValidation(
		chainBeforeValidation(opts.BeforeValidation, this.fillCodeFromName), this.fillParentWithRoot)
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateCategoryCreate)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

// codeLabelLanguage is the translation a generated code is derived from when the name carries it;
// otherwise the first language in sorted order, so the result does not depend on map iteration.
const codeLabelLanguage = "en-US"

// maxCodeSuffixAttempts bounds the "_2", "_3", … search for a free code within the org.
const maxCodeSuffixAttempts = 20

// fillCodeFromName derives `code` from `name` when the caller gave none, so a category can be
// created from its label alone (the import's "create referenced record" does exactly that). The
// code is unique per org, so a taken one gets a numeric suffix. A name-less create is left to
// schema validation, which reports both missing fields.
func (this *ProductCategoryDomainServiceImpl) fillCodeFromName(
	ctx corectx.Context, entity *composable.DynamicEntity, _ *ft.ClientErrors,
) (*composable.DynamicEntity, error) {
	category := models.NewProductCategoryFrom(entity.GetFieldData())
	if derefString(category.GetCode()) != "" {
		return entity, nil
	}
	base := codify.Codify(nameLabel(category), codify.DefaultMaxLength-4)
	if base == "" {
		return entity, nil
	}
	code, err := this.freeCode(ctx, base, derefString(category.GetOrgId()))
	if err != nil {
		return nil, err
	}
	if code != "" {
		category.SetCode(&code)
	}
	return entity, nil
}

func nameLabel(category *models.ProductCategory) string {
	name := category.GetName()
	if name == nil || len(*name) == 0 {
		return ""
	}
	if text, ok := (*name)[codeLabelLanguage]; ok && text != "" {
		return text
	}
	languages := make([]string, 0, len(*name))
	for language := range *name {
		languages = append(languages, string(language))
	}
	sort.Strings(languages)
	return (*name)[model.LanguageCode(languages[0])]
}

// freeCode answers base, or base with the first free numeric suffix; empty after the last attempt
// so the create falls through to the duplicate check and reports the collision honestly.
func (this *ProductCategoryDomainServiceImpl) freeCode(ctx corectx.Context, base string, orgId string) (string, error) {
	for attempt := 1; attempt <= maxCodeSuffixAttempts; attempt++ {
		candidate := base
		if attempt > 1 {
			candidate = fmt.Sprintf("%s_%d", base, attempt)
		}
		taken, err := this.codeExists(ctx, candidate, orgId)
		if err != nil {
			return "", err
		}
		if !taken {
			return candidate, nil
		}
	}
	return "", nil
}

func (this *ProductCategoryDomainServiceImpl) codeExists(ctx corectx.Context, code string, orgId string) (bool, error) {
	filter := dmodel.DynamicFields{models.ProductCategoryFieldCode: code}
	if orgId != "" {
		filter[models.ProductCategoryFieldOrgId] = orgId
	}
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: filter,
		Fields: []string{models.ProductCategoryFieldId},
	})
	if err != nil {
		return false, errors.Wrap(err, "fillCodeFromName")
	}
	return found != nil && found.HasData, nil
}

// chainBeforeValidation runs the caller's hook first, then ours, over the entity the first returns.
func chainBeforeValidation(first composable.BeforeValidationFn, second composable.BeforeValidationFn) composable.BeforeValidationFn {
	if first == nil {
		return second
	}
	return func(ctx corectx.Context, entity *composable.DynamicEntity, vErrs *ft.ClientErrors) (*composable.DynamicEntity, error) {
		entity, err := first(ctx, entity, vErrs)
		if err != nil || entity == nil {
			return entity, err
		}
		return second(ctx, entity, vErrs)
	}
}

// Update rewrites the descendants' canonical names in the same transaction when the name or the
// parent changes, so the tree never shows a path that no longer exists.
func (this *ProductCategoryDomainServiceImpl) Update(
	ctx corectx.Context, cmd itProduct.UpdateProductCategoryCommand, options ...composable.UpdateOptions,
) (*itProduct.UpdateProductCategoryResult, error) {
	vErrs := ft.NewClientErrors()
	rejectServerManagedFields(cmd, vErrs)
	rejectClearedParent(cmd, vErrs)
	if vErrs.Count() > 0 {
		return &itProduct.UpdateProductCategoryResult{ClientErrors: *vErrs}, nil
	}
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateCategoryUpdate)
	if !changesTree(cmd) {
		return this.CrudDomainService.Update(ctx, cmd, opts)
	}

	var result *itProduct.UpdateProductCategoryResult
	err := runInTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		res, err := this.CrudDomainService.Update(tranxCtx, cmd, opts)
		if err != nil {
			return err
		}
		result = res
		if res.ClientErrors.Count() > 0 {
			return composable.ClientErrorSignal{Errors: res.ClientErrors}
		}
		if !res.HasData {
			return nil
		}
		return this.cascadeCanonicalNames(tranxCtx, readStringParam(cmd, models.ProductCategoryFieldId))
	})
	if signal, ok := composable.AsClientErrorSignal(err); ok {
		return &itProduct.UpdateProductCategoryResult{ClientErrors: signal.Errors}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "ProductCategoryDomainService.Update")
	}
	return result, nil
}

func changesTree(cmd dmodel.DynamicFields) bool {
	_, renamed := cmd[models.ProductCategoryFieldName]
	_, moved := cmd[models.ProductCategoryFieldParentCategoryId]
	return renamed || moved
}

// rejectClearedParent refuses an explicit null parent, which the NOT NULL column would otherwise
// turn into a server error.
func rejectClearedParent(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if value, sent := cmd[models.ProductCategoryFieldParentCategoryId]; sent && value == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
			"product_category.parent_required",
			"a category must have a parent"))
	}
}

// validateCategoryCreate rejects a category that would be its own ancestor or whose parent sits in
// another org, then derives its canonical name. On create the record has no id yet, so only the
// parent chain itself can be malformed.
func (this *ProductCategoryDomainServiceImpl) validateCategoryCreate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	category := models.NewProductCategoryFrom(inputModel.GetFieldData())
	parentId := derefString(category.GetParentCategoryId())
	if err := this.assertNoCategoryCycle(ctx, parentId, "", vErrs); err != nil || vErrs.Count() > 0 {
		return err
	}
	parent, err := this.findCategoryById(ctx, parentId)
	if err != nil {
		return errors.Wrap(err, "validateCategoryCreate")
	}
	// The entity is the one inserted next, so the derived name is persisted with it.
	assignCanonicalName(category, derefString(category.GetOrgId()), parent, category.GetName(), vErrs)
	return nil
}

// validateCategoryUpdate keeps the root fixed, rejects re-parenting a category under itself or one
// of its descendants (which is how a cycle is normally introduced), and re-derives the canonical
// name when the name or the parent changes.
func (this *ProductCategoryDomainServiceImpl) validateCategoryUpdate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if foundModel == nil {
		return nil
	}
	submitted := models.NewProductCategoryFrom(inputModel.GetFieldData())
	stored := models.NewProductCategoryFrom(foundModel.GetFieldData())
	if stored.IsRoot() {
		assertRootUnchanged(submitted, stored, vErrs)
		return nil
	}
	parentId := submitted.GetParentCategoryId()
	if parentId == nil && submitted.GetName() == nil {
		// Neither the parent nor the name changes, so the chain and the canonical name still hold.
		return nil
	}

	selfId := derefString(stored.GetId())
	effectiveParentId := derefString(stored.GetParentCategoryId())
	if parentId != nil {
		effectiveParentId = derefString(parentId)
		if effectiveParentId == selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
				"product_category.self_parent",
				"a category cannot be its own parent"))
			return nil
		}
		if err := this.assertNoCategoryCycle(ctx, effectiveParentId, selfId, vErrs); err != nil || vErrs.Count() > 0 {
			return err
		}
	}
	return this.reassignCanonicalName(ctx, submitted, stored, effectiveParentId, vErrs)
}

func (this *ProductCategoryDomainServiceImpl) reassignCanonicalName(
	ctx corectx.Context, submitted *models.ProductCategory, stored *models.ProductCategory,
	parentId string, vErrs *ft.ClientErrors,
) error {
	parent, err := this.findCategoryById(ctx, parentId)
	if err != nil {
		return errors.Wrap(err, "validateCategoryUpdate")
	}
	name := submitted.GetName()
	if name == nil {
		name = stored.GetName()
	}
	// The input entity is the one written next, so the derived name is persisted with it.
	assignCanonicalName(submitted, derefString(stored.GetOrgId()), parent, name, vErrs)
	return nil
}

// assertRootUnchanged keeps the root named "All" and parented by itself.
func assertRootUnchanged(submitted *models.ProductCategory, stored *models.ProductCategory, vErrs *ft.ClientErrors) {
	parentId := submitted.GetParentCategoryId()
	moved := parentId != nil && derefString(parentId) != derefString(stored.GetId())
	name := submitted.GetName()
	renamed := name != nil && (stored.GetName() == nil || !reflect.DeepEqual(*name, *stored.GetName()))
	if moved || renamed {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldId,
			"product_category.root_immutable",
			"the root category cannot be renamed, moved, archived or deleted"))
	}
}

// assertNoCategoryCycle walks upwards from parentId. Reaching selfId means the proposed parent
// is a descendant of the category being edited, so the edit would close a loop.
func (this *ProductCategoryDomainServiceImpl) assertNoCategoryCycle(
	ctx corectx.Context, parentId string, selfId string, vErrs *ft.ClientErrors,
) error {
	if parentId == "" {
		return nil
	}

	seen := map[string]bool{}
	current := parentId
	for depth := 0; current != "" && depth < maxCategoryDepth; depth++ {
		if current == selfId {
			vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
				"product_category.cycle",
				"this parent is a descendant of the category, which would create a cycle"))
			return nil
		}
		if seen[current] {
			// A pre-existing cycle in stored data. Stop rather than loop, and let the edit
			// through: it is not the change under validation that introduced it.
			return nil
		}
		seen[current] = true

		found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
			Filter: dmodel.DynamicFields{models.ProductCategoryFieldId: current},
			Fields: []string{models.ProductCategoryFieldId, models.ProductCategoryFieldParentCategoryId},
		})
		if err != nil {
			return errors.Wrap(err, "assertNoCategoryCycle")
		}
		if !found.HasData {
			// A missing parent is the reference check's business, not ours.
			return nil
		}
		next := derefString(models.NewProductCategoryFrom(found.Data).GetParentCategoryId())
		if next == current {
			// The root is its own parent: the chain ends here.
			return nil
		}
		current = next
	}
	return nil
}

// Delete refuses to remove a category that still holds products or child categories. Without the
// guard either foreign key refuses the statement and the failure reaches the client as a 500.
func (this *ProductCategoryDomainServiceImpl) Delete(
	ctx corectx.Context, params dmodel.DynamicFields, options ...composable.DeleteOptions,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	templateRepo, err := repoFor(models.ProductTemplateSchemaName)
	if err != nil {
		return nil, err
	}

	vErrs := ft.NewClientErrors()
	categoryId := readStringParam(params, models.ProductCategoryFieldId)
	if err := this.assertNotRoot(ctx, categoryId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	if err := AssertProductCategoryDeletable(ctx, templateRepo, this.Repository(), categoryId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
}

// SetArchived refuses to archive the root: every category hangs under it, and creates default to it.
func (this *ProductCategoryDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd itProduct.SetProductCategoryArchivedCommand,
) (*itProduct.SetProductCategoryArchivedResult, error) {
	if archived, ok := cmd[basemodel.FieldIsArchived].(bool); ok && archived {
		vErrs := ft.NewClientErrors()
		if err := this.assertNotRoot(ctx, readStringParam(cmd, models.ProductCategoryFieldId), vErrs); err != nil {
			return nil, err
		}
		if vErrs.Count() > 0 {
			return &itProduct.SetProductCategoryArchivedResult{ClientErrors: *vErrs}, nil
		}
	}
	return this.CrudDomainService.SetArchived(ctx, cmd)
}

func (this *ProductCategoryDomainServiceImpl) assertNotRoot(
	ctx corectx.Context, categoryId string, vErrs *ft.ClientErrors,
) error {
	category, err := this.findCategoryById(ctx, categoryId)
	if err != nil {
		return errors.Wrap(err, "assertNotRoot")
	}
	if category != nil && category.IsRoot() {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldId,
			"product_category.root_immutable",
			"the root category cannot be renamed, moved, archived or deleted"))
	}
	return nil
}

// AssertProductCategoryDeletable reports every kind of child blocking the delete rather than
// stopping at the first, so a user clearing them sees the whole list in one response.
func AssertProductCategoryDeletable(
	ctx corectx.Context, templateRepo models.ProductSearcher, categoryRepo models.ProductSearcher,
	categoryId string, vErrs *ft.ClientErrors,
) error {
	if categoryId == "" {
		return nil
	}

	templates, err := models.FindTemplatesByCategory(ctx, templateRepo, categoryId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertProductCategoryDeletable")
	}
	if len(templates) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldId,
			"product_category.has_templates",
			"this category still has products"))
	}

	children, err := models.FindChildCategories(ctx, categoryRepo, categoryId, 1)
	if err != nil {
		return errors.Wrap(err, "AssertProductCategoryDeletable")
	}
	if len(children) > 0 {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldId,
			"product_category.has_children",
			"this category still has child categories"))
	}
	return nil
}
