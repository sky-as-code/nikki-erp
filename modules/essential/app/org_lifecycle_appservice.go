package app

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle"
	"github.com/sky-as-code/nikki-erp/modules/core/orglifecycle/defaults"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
	orgDefaults "github.com/sky-as-code/nikki-erp/modules/essential/infra/defaults"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	itExt "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/external"
	itLanguage "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/language"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itUomCat "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uomcat"
)

// OrgLifecycleParam is what the org lifecycle writes through: the UoM catalogue, and the
// activations and settings that decide which languages and currencies a new org uses.
type OrgLifecycleParam struct {
	UomCatSvc    itUomCat.UomCatDomainService
	UomSvc       itUom.UomDomainService
	LanguageSvc  itLanguage.LanguageDomainService
	CurrencySvc  itCurrency.CurrencyDomainService
	Languages    *services.LanguageActivation
	Currencies   *services.CurrencyActivation
	SettingsSeed itExt.OrgSettingsSeedExtService
}

func NewOrgLifecycleAppService(param OrgLifecycleParam) orglifecycle.OrgLifecycleService {
	return &OrgLifecycleAppServiceImpl{param: param, uomCatSvc: param.UomCatSvc, uomSvc: param.UomSvc}
}

// OrgLifecycleAppServiceImpl gives a new org its unit-of-measure catalogue, its active languages
// and currencies, and its system language.
type OrgLifecycleAppServiceImpl struct {
	param     OrgLifecycleParam
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

	return dynamicresource.InTransaction(ctx, this.uomCatSvc.Repository(), func(tranxCtx corectx.Context) error {
		aliases := defaults.NewAliases()
		if err := dynamicresource.CreateDefaults(tranxCtx, this.uomCatSvc, cats, aliases, orgId); err != nil {
			return err
		}
		if err := dynamicresource.CreateDefaults(tranxCtx, this.uomSvc, uoms, aliases, orgId); err != nil {
			return err
		}
		if err := this.linkReferenceUoms(tranxCtx, aliases, referenceOf, orgId); err != nil {
			return err
		}
		return this.initActivations(tranxCtx, orgId)
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
		if err := dynamicresource.UpdateOne(ctx, this.uomCatSvc, catId, fields); err != nil {
			return err
		}
	}
	return nil
}

func (this *OrgLifecycleAppServiceImpl) CleanUpOrg(ctx corectx.Context, orgId model.Id) error {
	return dynamicresource.InTransaction(ctx, this.uomCatSvc.Repository(), func(tranxCtx corectx.Context) error {
		if err := this.param.Languages.DeleteForOrg(tranxCtx, orgId); err != nil {
			return err
		}
		if err := this.param.Currencies.DeleteForOrg(tranxCtx, orgId); err != nil {
			return err
		}
		// Children before parents: issuances name counters, both name definitions (RESTRICT).
		// Tenant-wide definitions carry no org_id and stay. The consumers holding a definition id
		// (commerce, payment invoice) depend on essential and are cleaned up before it.
		if err := dynamicresource.DeleteOrgDataOf(tranxCtx, orgId, models.SequenceIssuanceSchemaName,
			models.SequenceCounterSchemaName, models.SequenceDefinitionSchemaName); err != nil {
			return err
		}
		return dynamicresource.DeleteOrgData(tranxCtx, orgId, this.uomSvc, this.uomCatSvc)
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
