package services

import (
	"sort"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// langJsonRefKey marks a translation-key name ({"$ref": key}); it is not a language.
const langJsonRefKey = "$ref"

// cascadePageSize is how many children of one tree level the canonical-name cascade reads at once.
const cascadePageSize = 500

// BuildCanonicalName appends name to the parent's canonical name, per language. The languages
// are the union of both sides; a side lacking one contributes its fallback segment. Because every
// canonical name already carries the union of its own chain, a parent lacking a language means no
// ancestor has it, so the parent's fallback is exactly the per-node fallback of the whole chain.
// A nil parent yields the name alone, which is the root's canonical name.
func BuildCanonicalName(parentCanonical *model.LangJson, name *model.LangJson) model.LangJson {
	result := model.LangJson{}
	for _, language := range unionLanguages(parentCanonical, name) {
		own := langSegment(name, language)
		if parentCanonical == nil || len(*parentCanonical) == 0 {
			result[language] = own
			continue
		}
		result[language] = langSegment(parentCanonical, language) + models.ProductCategoryCanonicalSeparator + own
	}
	return result
}

func unionLanguages(values ...*model.LangJson) []model.LanguageCode {
	seen := map[model.LanguageCode]bool{}
	languages := make([]string, 0, 2)
	for _, value := range values {
		if value == nil {
			continue
		}
		for language := range *value {
			if language == langJsonRefKey || seen[language] {
				continue
			}
			seen[language] = true
			languages = append(languages, string(language))
		}
	}
	if len(languages) == 0 {
		languages = append(languages, string(model.DefaultLanguageCode))
	}
	sort.Strings(languages)
	return toLanguageCodes(languages)
}

func toLanguageCodes(languages []string) []model.LanguageCode {
	codes := make([]model.LanguageCode, len(languages))
	for i, language := range languages {
		codes[i] = model.LanguageCode(language)
	}
	return codes
}

// langSegment is the text for language, else the default language's, else the first language in
// sorted order (the same rule as nameLabel), else a translation key's raw text.
func langSegment(value *model.LangJson, language model.LanguageCode) string {
	if value == nil {
		return ""
	}
	if text := (*value)[language]; text != "" {
		return text
	}
	if text := (*value)[model.DefaultLanguageCode]; text != "" {
		return text
	}
	languages := make([]string, 0, len(*value))
	for key := range *value {
		if key != langJsonRefKey {
			languages = append(languages, string(key))
		}
	}
	if len(languages) == 0 {
		return (*value)[langJsonRefKey]
	}
	sort.Strings(languages)
	return (*value)[model.LanguageCode(languages[0])]
}

// rejectServerManagedFields refuses a command carrying canonical_name or is_root. Both are the
// server's: the first is derived from the tree, the second is set only on the seeded root.
func rejectServerManagedFields(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if _, sent := cmd[models.ProductCategoryFieldCanonicalName]; sent {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldCanonicalName,
			"product_category.canonical_name_server_managed",
			"canonical_name is maintained by the server and cannot be set"))
	}
	if _, sent := cmd[models.ProductCategoryFieldIsRoot]; sent {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldIsRoot,
			"product_category.is_root_server_managed",
			"is_root is maintained by the server and cannot be set"))
	}
}

// findCategory reads one category by id; nil when it does not exist.
func (this *ProductCategoryDomainServiceImpl) findCategory(
	ctx corectx.Context, filter dmodel.DynamicFields,
) (*models.ProductCategory, error) {
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{
		Filter: filter,
		Fields: []string{
			models.ProductCategoryFieldId, models.ProductCategoryFieldOrgId, models.ProductCategoryFieldName,
			models.ProductCategoryFieldParentCategoryId, models.ProductCategoryFieldIsRoot,
			models.ProductCategoryFieldCanonicalName,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "findCategory")
	}
	if found != nil && found.ClientErrors.Count() > 0 {
		return nil, errors.Wrap(found.ClientErrors.ToError(), "findCategory")
	}
	if found == nil || !found.HasData {
		return nil, nil
	}
	return models.NewProductCategoryFrom(found.Data), nil
}

func (this *ProductCategoryDomainServiceImpl) findCategoryById(
	ctx corectx.Context, categoryId string,
) (*models.ProductCategory, error) {
	if categoryId == "" {
		return nil, nil
	}
	return this.findCategory(ctx, dmodel.DynamicFields{models.ProductCategoryFieldId: categoryId})
}

// fillParentWithRoot puts a parent-less create under the org's root. An org without a root cannot
// hold categories yet: creating the root with the org is a later feature, until then it is seeded.
func (this *ProductCategoryDomainServiceImpl) fillParentWithRoot(
	ctx corectx.Context, entity *composable.DynamicEntity, vErrs *ft.ClientErrors,
) (*composable.DynamicEntity, error) {
	category := models.NewProductCategoryFrom(entity.GetFieldData())
	orgId := derefString(category.GetOrgId())
	if derefString(category.GetParentCategoryId()) != "" || orgId == "" {
		return entity, nil
	}
	rootId, err := this.findRootId(ctx, orgId)
	if err != nil {
		return nil, errors.Wrap(err, "fillParentWithRoot")
	}
	if rootId == "" {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
			"product_category.no_root",
			"this organization has no root category"))
		return entity, nil
	}
	category.SetParentCategoryId(&rootId)
	return entity, nil
}

// findRootId searches rather than fetching one: GetOne accepts only a whole key as its filter,
// and the root's key (tenant_id, org_id, is_root) is completed by the repository's tenant scope.
func (this *ProductCategoryDomainServiceImpl) findRootId(ctx corectx.Context, orgId string) (string, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(models.ProductCategoryFieldIsRoot, dmodel.Equals, true),
		*dmodel.NewSearchNode().NewCondition(models.ProductCategoryFieldOrgId, dmodel.Equals, orgId),
	)
	found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{
		Graph: graph, Page: 0, Size: 1, Fields: []string{models.ProductCategoryFieldId},
	})
	if err != nil {
		return "", errors.Wrap(err, "findRootId")
	}
	if found.ClientErrors.Count() > 0 {
		return "", errors.Wrap(found.ClientErrors.ToError(), "findRootId")
	}
	if !found.HasData || len(found.Data.Items) == 0 {
		return "", nil
	}
	return derefString(models.NewProductCategoryFrom(found.Data.Items[0]).GetId()), nil
}

// assignCanonicalName checks the parent belongs to the same org, then derives the category's
// canonical name from it. A missing parent is left to the reference check.
func assignCanonicalName(
	category *models.ProductCategory, orgId string, parent *models.ProductCategory, name *model.LangJson,
	vErrs *ft.ClientErrors,
) {
	if parent == nil {
		return
	}
	if derefString(parent.GetOrgId()) != orgId {
		vErrs.Append(*ft.NewBusinessViolation(models.ProductCategoryFieldParentCategoryId,
			"product_category.parent_other_org",
			"the parent category belongs to another organization"))
		return
	}
	parentCanonical := parent.GetCanonicalName()
	if parentCanonical == nil {
		parentCanonical = parent.GetName()
	}
	canonical := BuildCanonicalName(parentCanonical, name)
	category.SetCanonicalName(&canonical)
}

// runInTransaction joins the transaction ctx already carries (an import or bulk run), since the
// database client refuses a nested one, and otherwise opens its own.
func runInTransaction(
	ctx corectx.Context, repo composable.CrudRepository, body func(tranxCtx corectx.Context) error,
) error {
	if ctx.GetDbTranx() != nil {
		return body(ctx)
	}
	return composable.WithTransaction(ctx, repo, body)
}

type canonicalNode struct {
	id        string
	canonical model.LangJson
}

// cascadeCanonicalNames rewrites the canonical name of every descendant of categoryId, level by
// level. The visited set also stops at the root, whose parent is itself.
func (this *ProductCategoryDomainServiceImpl) cascadeCanonicalNames(ctx corectx.Context, categoryId string) error {
	start, err := this.findCategoryById(ctx, categoryId)
	if err != nil || start == nil || start.GetCanonicalName() == nil {
		return err
	}
	visited := map[string]bool{categoryId: true}
	frontier := []canonicalNode{{id: categoryId, canonical: *start.GetCanonicalName()}}

	for depth := 0; depth < maxCategoryDepth && len(frontier) > 0; depth++ {
		next, err := this.rewriteChildren(ctx, frontier, visited)
		if err != nil {
			return err
		}
		frontier = next
	}
	return nil
}

func (this *ProductCategoryDomainServiceImpl) rewriteChildren(
	ctx corectx.Context, parents []canonicalNode, visited map[string]bool,
) ([]canonicalNode, error) {
	parentCanonical := make(map[string]model.LangJson, len(parents))
	parentIds := make([]any, 0, len(parents))
	for _, parent := range parents {
		parentCanonical[parent.id] = parent.canonical
		parentIds = append(parentIds, parent.id)
	}

	children, err := this.readChildren(ctx, parentIds)
	if err != nil {
		return nil, err
	}
	next := make([]canonicalNode, 0, len(children))
	for _, child := range children {
		childId := derefString(child.GetId())
		if visited[childId] {
			continue
		}
		visited[childId] = true
		canonical := parentCanonical[derefString(child.GetParentCategoryId())]
		childCanonical := BuildCanonicalName(&canonical, child.GetName())
		if err := this.writeCanonicalName(ctx, childId, childCanonical); err != nil {
			return nil, err
		}
		next = append(next, canonicalNode{id: childId, canonical: childCanonical})
	}
	return next, nil
}

func (this *ProductCategoryDomainServiceImpl) readChildren(
	ctx corectx.Context, parentIds []any,
) ([]*models.ProductCategory, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(*dmodel.NewSearchNode().NewCondition(models.ProductCategoryFieldParentCategoryId, dmodel.In, parentIds...))

	children := make([]*models.ProductCategory, 0)
	for page := 0; ; page++ {
		found, err := this.Repository().Search(ctx, dyn.RepoSearchParam{
			Graph: graph,
			Page:  page,
			Size:  cascadePageSize,
			Fields: []string{
				models.ProductCategoryFieldId, models.ProductCategoryFieldParentCategoryId, models.ProductCategoryFieldName,
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "readChildren")
		}
		if found.ClientErrors.Count() > 0 {
			return nil, errors.Wrap(found.ClientErrors.ToError(), "readChildren")
		}
		if !found.HasData {
			return children, nil
		}
		for _, item := range found.Data.Items {
			children = append(children, models.NewProductCategoryFrom(item))
		}
		if len(found.Data.Items) < cascadePageSize {
			return children, nil
		}
	}
}

// writeCanonicalName goes straight to the repository: the value is derived, not client input, so
// it skips validation and the etag check, while the repository still bumps etag and updated_at.
func (this *ProductCategoryDomainServiceImpl) writeCanonicalName(
	ctx corectx.Context, categoryId string, canonical model.LangJson,
) error {
	data := dmodel.DynamicFields{models.ProductCategoryFieldId: categoryId}
	data.SetLangJson(models.ProductCategoryFieldCanonicalName, &canonical)
	result, err := this.Repository().Update(ctx, data)
	if err != nil {
		return errors.Wrap(err, "writeCanonicalName")
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrap(result.ClientErrors.ToError(), "writeCanonicalName")
	}
	return nil
}
