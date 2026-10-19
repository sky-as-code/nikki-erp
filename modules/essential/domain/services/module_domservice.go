package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/semver"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	itLang "github.com/sky-as-code/nikki-erp/modules/core/language/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

// NewModuleDomainService is handed the composable default by the module metadata onion. The CRUD
// is the default's, with the field rules chained into create and update and internal modules
// hidden from every read; the boot-time sync is this type's own.
func NewModuleDomainService(
	base composable.CrudDomainService, cqrsBus cqrs.CqrsBus, logger logging.LoggerService,
) it.ModuleDomainService {
	return &ModuleDomainServiceImpl{CrudDomainService: base, cqrsBus: cqrsBus, logger: logger}
}

type ModuleDomainServiceImpl struct {
	composable.CrudDomainService
	cqrsBus cqrs.CqrsBus
	logger  logging.LoggerService
}

func (this *ModuleDomainServiceImpl) moduleRepo() it.ModuleRepository {
	return this.Repository().(it.ModuleRepository)
}

func (this *ModuleDomainServiceImpl) Create(
	ctx corectx.Context, cmd it.CreateModuleCommand, options ...composable.CreateOptions,
) (*it.CreateModuleResult, error) {
	opts := safe.GetOptional(options, composable.CreateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateModuleFields)
	return this.CrudDomainService.Create(ctx, cmd, opts)
}

func (this *ModuleDomainServiceImpl) Update(
	ctx corectx.Context, cmd it.UpdateModuleCommand, options ...composable.UpdateOptions,
) (*it.UpdateModuleResult, error) {
	opts := safe.GetOptional(options, composable.UpdateOptions{})
	opts.ValidateExtra = composable.ChainValidateExtra(opts.ValidateExtra, this.validateModuleFields)
	return this.CrudDomainService.Update(ctx, cmd, opts)
}

// GetById answers "not found" for an internal module: internal modules are platform plumbing
// that no screen lists, and naming one by id must not reveal it either.
func (this *ModuleDomainServiceImpl) GetById(
	ctx corectx.Context, query it.GetModuleByIdQuery,
) (*it.GetModuleByIdResult, error) {
	result, err := this.CrudDomainService.GetById(ctx, query)
	if err != nil || result.ClientErrors.Count() > 0 || !result.HasData {
		return result, err
	}
	internal, err := this.isInternal(ctx, result.Data.Item)
	if err != nil {
		return nil, err
	}
	if internal {
		return &it.GetModuleByIdResult{HasData: false}, nil
	}
	return result, nil
}

// isInternal reads the flag off the row, or from the table when the caller's field list left it
// out of the row.
func (this *ModuleDomainServiceImpl) isInternal(ctx corectx.Context, row dmodel.DynamicFields) (bool, error) {
	if flag, ok := row[models.ModuleMetadataFieldIsInternal].(bool); ok {
		return flag, nil
	}
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{Filter: dmodel.DynamicFields{
		models.ModuleMetadataFieldId:         row[models.ModuleMetadataFieldId],
		models.ModuleMetadataFieldIsInternal: true,
	}})
	if err != nil {
		return false, errors.Wrap(err, "module isInternal")
	}
	return found.HasData, nil
}

// Search never returns an internal module, whatever graph the caller sent.
func (this *ModuleDomainServiceImpl) Search(
	ctx corectx.Context, query it.SearchModulesQuery, options ...corecrud.ServiceSearchOptions,
) (*it.SearchModulesResult, error) {
	opts := safe.GetOptional(options, corecrud.ServiceSearchOptions{})
	callerHook := opts.AfterValidationSuccess
	opts.AfterValidationSuccess = func(ctx corectx.Context, query dyn.SearchQuery) (dyn.SearchQuery, error) {
		if callerHook != nil {
			var err error
			if query, err = callerHook(ctx, query); err != nil {
				return query, err
			}
		}
		return excludeInternalModules(query), nil
	}
	return this.CrudDomainService.Search(ctx, query, opts)
}

func excludeInternalModules(query dyn.SearchQuery) dyn.SearchQuery {
	excludeInternal := dmodel.NewSearchNode().NewCondition(models.ModuleMetadataFieldIsInternal, dmodel.Equals, false)
	if query.Graph == nil {
		query.Graph = dmodel.NewSearchGraph().NewCondition(models.ModuleMetadataFieldIsInternal, dmodel.Equals, false)
		return query
	}
	query.Graph = dmodel.NewSearchGraph().
		And(*excludeInternal, *query.Graph.ToSearchNode()).
		Order(query.Graph.GetOrder())
	return query
}

// validateModuleFields checks the name and version formats and narrows the label to the enabled
// languages, writing the narrowed label back onto the input.
func (this *ModuleDomainServiceImpl) validateModuleFields(
	ctx corectx.Context, inputModel *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	module := models.NewModuleMetadataFrom(inputModel.GetFieldData())
	if module.GetName() != nil {
		if err := val.ApiBased.ValidateRaw(*module.GetName(), model.ModelRuleModuleLogicalName); err != nil {
			vErrs.Append(*ft.NewValidationError(models.ModuleMetadataFieldName,
				ft.ErrorKey("module.invalid_name", "essential"),
				"name must contain only letters, numbers, underscores, and dots"))
		}
	}
	if rawVersion := module.GetFieldData().GetString(models.ModuleMetadataFieldVersion); rawVersion != nil {
		if _, err := semver.ParseSemVer(*rawVersion); err != nil {
			vErrs.Append(*ft.NewValidationError(models.ModuleMetadataFieldVersion,
				ft.ErrorKey("module.invalid_version", "essential"),
				"version must use valid semantic version format"))
		}
	}
	if module.GetLabel() == nil {
		return nil
	}
	return this.normalizeLabel(ctx, module, vErrs)
}

func (this *ModuleDomainServiceImpl) normalizeLabel(
	ctx corectx.Context, module *models.ModuleMetadata, vErrs *ft.ClientErrors,
) error {
	whitelistLangs, err := this.getEnabledLanguages(ctx)
	if err != nil {
		return err
	}
	newLabel, fieldCount, err := module.GetLabel().SanitizeClone(whitelistLangs, false)
	if err != nil {
		return err
	}
	if fieldCount == 0 {
		vErrs.Append(*ft.NewValidationError(models.ModuleMetadataFieldLabel,
			ft.ErrorKey("module.label_requires_enabled_language", "essential"),
			"label must contain at least one enabled language"))
		return nil
	}
	module.SetLabel(newLabel)
	return nil
}

func (this *ModuleDomainServiceImpl) getEnabledLanguages(ctx corectx.Context) ([]model.LanguageCode, error) {
	query := itLang.ListEnabledLangCodesQuery{}
	result := itLang.ListEnabledLangCodesResult{}
	if err := this.cqrsBus.Request(ctx, query, &result); err != nil {
		return nil, err
	}
	return result.Data, nil
}
