package services

import (
	"fmt"
	"sort"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/codify"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

// NewProductCategoryDomainService attaches the category-tree invariant to create and update, and
// derives a missing code from the name on create. The CRUD processing is the default's; only the
// rules the schema cannot express belong here.
func NewProductCategoryDomainService(base composable.CrudDomainService) itProduct.ProductCategoryDomainService {
	return &ProductCategoryDomainServiceImpl{CrudDomainService: base}
}

type ProductCategoryDomainServiceImpl struct {
	composable.CrudDomainService
}

func (this *ProductCategoryDomainServiceImpl) Create(
	ctx corectx.Context, cmd itProduct.CreateProductCategoryCommand, options ...composable.CreateOptions,
) (*itProduct.CreateProductCategoryResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.BeforeValidation = chainBeforeValidation(opts.BeforeValidation, this.fillCodeFromName)
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

func (this *ProductCategoryDomainServiceImpl) Update(
	ctx corectx.Context, cmd itProduct.UpdateProductCategoryCommand, options ...composable.UpdateOptions,
) (*itProduct.UpdateProductCategoryResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateCategoryUpdate)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// validateCategoryCreate rejects a category that would be its own ancestor. On create the record
// has no id yet, so only the parent chain itself can be malformed.
func (this *ProductCategoryDomainServiceImpl) validateCategoryCreate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	category := models.NewProductCategoryFrom(inputModel.GetFieldData())
	return this.assertNoCategoryCycle(ctx, derefString(category.GetParentCategoryId()), "", vErrs)
}

// validateCategoryUpdate additionally rejects re-parenting a category under one of its own
// descendants, which is how a cycle is normally introduced.
func (this *ProductCategoryDomainServiceImpl) validateCategoryUpdate(
	ctx corectx.Context, inputModel *composable.DynamicEntity, foundModel *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if foundModel == nil {
		return nil
	}
	submitted := models.NewProductCategoryFrom(inputModel.GetFieldData())
	parentId := submitted.GetParentCategoryId()
	if parentId == nil {
		// The parent is not being changed, so the existing chain still holds.
		return nil
	}

	stored := models.NewProductCategoryFrom(foundModel.GetFieldData())
	selfId := derefString(stored.GetId())

	if derefString(parentId) == selfId {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
			"product_category.self_parent",
			"a category cannot be its own parent"))
		return nil
	}
	return this.assertNoCategoryCycle(ctx, derefString(parentId), selfId, vErrs)
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
		current = derefString(models.NewProductCategoryFrom(found.Data).GetParentCategoryId())
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
	if err := AssertProductCategoryDeletable(ctx, templateRepo, this.Repository(), categoryId, vErrs); err != nil {
		return nil, err
	}
	if vErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}, nil
	}
	return this.CrudDomainService.Delete(ctx, params, options...)
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
