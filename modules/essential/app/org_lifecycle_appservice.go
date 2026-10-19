package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/essential/infra/defaults"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

func NewOrgLifecycleAppService(
	uomCatSvc itUomCat.UomCatDomainService, uomSvc itUom.UomDomainService,
) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{uomCatSvc: uomCatSvc, uomSvc: uomSvc}
}

// OrgLifecycleAppServiceImpl gives a new org its unit-of-measure catalogue.
type OrgLifecycleAppServiceImpl struct {
	uomCatSvc itUomCat.UomCatDomainService
	uomSvc    itUom.UomDomainService
}

// InitOrg writes the categories, then their units, then points each category at its reference
// unit: a category and its reference unit name each other, so one must exist without the edge.
func (this *OrgLifecycleAppServiceImpl) InitOrg(ctx corectx.Context, orgId model.Id) error {
	cats, err := defaults.Load(orgDefaults.Files, models.UomCatSchemaName)
	if err != nil {
		return err
	}
	uoms, err := defaults.Load(orgDefaults.Files, models.UomSchemaName)
	if err != nil {
		return err
	}
	referenceOf := takeField(cats, models.UomCatFieldReferenceUomId)

	return composable.InTransaction(ctx, this.uomCatSvc.Repository(), func(tranxCtx corectx.Context) error {
		aliases := defaults.NewAliases()
		if err := composable.CreateDefaults(tranxCtx, this.uomCatSvc, cats, aliases, orgId); err != nil {
			return err
		}
		if err := composable.CreateDefaults(tranxCtx, this.uomSvc, uoms, aliases, orgId); err != nil {
			return err
		}
		return this.linkReferenceUoms(tranxCtx, aliases, referenceOf, orgId)
	})
}

func (this *OrgLifecycleAppServiceImpl) linkReferenceUoms(
	ctx corectx.Context, aliases *defaults.Aliases, referenceOf map[string]any, orgId model.Id,
) error {
	for catAlias, ref := range referenceOf {
		resolved, _, err := aliases.Resolve(dmodel.DynamicFields{models.UomCatFieldReferenceUomId: ref}, orgId)
		if err != nil {
			return err
		}
		catId, err := aliases.MustId(catAlias)
		if err != nil {
			return err
		}
		fields := dmodel.DynamicFields{models.UomCatFieldReferenceUomId: resolved[models.UomCatFieldReferenceUomId]}
		if err := composable.UpdateOne(ctx, this.uomCatSvc, catId, fields); err != nil {
			return err
		}
	}
	return nil
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	return composable.InTransaction(ctx, this.uomCatSvc.Repository(), func(tranxCtx corectx.Context) error {
		return composable.DeleteOrgData(tranxCtx, orgId, this.uomSvc, this.uomCatSvc)
	})
}

// takeField removes field from every row and returns it keyed by the row's alias.
func takeField(rows []dmodel.DynamicFields, field string) map[string]any {
	taken := make(map[string]any)
	for _, row := range rows {
		if value, found := row[field]; found {
			alias, _ := row[defaults.AliasKey].(string)
			taken[alias] = value
			delete(row, field)
		}
	}
	return taken
}
