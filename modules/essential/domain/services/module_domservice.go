package services

import (
	"fmt"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	i18n "github.com/sky-as-code/nikki-erp/modules/core/i18n/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

func NewModuleDomainServiceImpl(
	cqrsBus cqrs.CqrsBus,
	logger logging.LoggerService,
	moduleRepo it.ModuleRepository,
) it.ModuleDomainService {
	return &ModuleDomainServiceImpl{
		cqrsBus:    cqrsBus,
		logger:     logger,
		moduleRepo: moduleRepo,
	}
}

type ModuleDomainServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	logger     logging.LoggerService
	moduleRepo it.ModuleRepository
}

func (this *ModuleDomainServiceImpl) CreateModule(
	ctx corectx.Context, cmd it.CreateModuleCommand,
) (*it.CreateModuleResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.ModuleMetadata, *models.ModuleMetadata]{
		Action:         "create module metadata",
		BaseRepoGetter: this.moduleRepo,
		Data:           cmd,
		ValidateExtra:  this.validateModuleCreate,
	})
}

func (this *ModuleDomainServiceImpl) DeleteModule(
	ctx corectx.Context, cmd it.DeleteModuleCommand,
) (*it.DeleteModuleResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete module metadata",
		DbRepoGetter: this.moduleRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *ModuleDomainServiceImpl) GetModule(
	ctx corectx.Context, query it.GetModuleQuery,
) (*it.GetModuleResult, error) {
	result, err := corecrud.GetOne[models.ModuleMetadata](ctx, corecrud.GetOneParam{
		Action:       "get module metadata",
		DbRepoGetter: this.moduleRepo,
		Query:        dyn.GetOneQuery(query),
		AfterValidationSuccess: func(ctx corectx.Context, query dyn.GetOneQuery) (dyn.GetOneQuery, error) {
			if query.Fields != nil {
				query.Fields = append(query.Fields, models.ModuleMetadataFieldIsInternal)
			}
			// else: get all fields, no need to add is_internal
			return query, nil
		},
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return &it.GetModuleResult{
			ClientErrors: result.ClientErrors,
			HasData:      result.HasData,
		}, nil
	}
	if *result.Data.GetIsInternal() {
		return &it.GetModuleResult{
			Data:    *models.NewModuleMetadata(),
			HasData: false,
		}, nil
	}
	return result, nil
}

func (this *ModuleDomainServiceImpl) SearchModules(
	ctx corectx.Context, query it.SearchModulesQuery,
) (*it.SearchModulesResult, error) {
	return corecrud.Search[models.ModuleMetadata](ctx, corecrud.SearchParam{
		Action:       "search module metadata",
		DbRepoGetter: this.moduleRepo,
		Query:        dyn.SearchQuery(query),
		AfterValidationSuccess: func(ctx corectx.Context, query dyn.SearchQuery) (dyn.SearchQuery, error) {
			if query.Graph != nil {
				excludeInternal := dmodel.NewSearchNode().NewCondition(models.ModuleMetadataFieldIsInternal, dmodel.Equals, false)
				newGraph := dmodel.NewSearchGraph().
					And(
						*excludeInternal,
						*query.Graph.ToSearchNode(),
					).
					Order(query.Graph.GetOrder())
				query.Graph = newGraph
			} else {
				query.Graph = dmodel.NewSearchGraph().NewCondition(models.ModuleMetadataFieldIsInternal, dmodel.Equals, false)
			}
			return query, nil
		},
	})
}

func (this *ModuleDomainServiceImpl) ModuleExists(
	ctx corectx.Context, query it.ModuleExistsQuery,
) (*it.ModuleExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if module metadata exists",
		DbRepoGetter: this.moduleRepo,
		Query:        dyn.ExistsQuery(query),
	})
}

func (this *ModuleDomainServiceImpl) SyncModuleMetadata(
	ctx corectx.Context, installedModules []modules.InCodeModule,
) (isSuccess bool, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "sync module metadata"); e != nil {
			err = e
		}
	}()

	acquired, err := this.moduleRepo.AcquireLock(ctx)
	if err != nil {
		return false, err
	}
	if !acquired {
		this.logger.Debugf("Could not acquire lock to sync module metadata. Another instance has acquired it")
		return false, nil
	}

	defer func() {
		this.logger.Debugf("Releasing lock for module metadata sync")
		releaseErr := this.moduleRepo.ReleaseLock(ctx)
		if releaseErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; release module metadata lock: %w", err, releaseErr)
				return
			}
			err = fmt.Errorf("release module metadata lock: %w", releaseErr)
		}
	}()

	result, err := corecrud.ExecInTranx(ctx, this.moduleRepo, func(trxCtx corectx.Context) (*bool, error) {
		dbModules, err := this.listAllModules(trxCtx)
		if err != nil {
			return nil, err
		}

		dbMap := make(map[string]models.ModuleMetadata)
		for _, item := range dbModules {
			if item.GetName() != nil {
				dbMap[*item.GetName()] = item
			}
		}

		installedMap := make(map[string]modules.InCodeModule)
		for _, item := range installedModules {
			installedMap[item.Name()] = item
		}

		newCount := 0
		modifiedCount := 0
		orphanedCount := 0

		for name, installedModule := range installedMap {
			dbModule, exists := dbMap[name]
			if !exists {
				if err := this.syncCreateModule(trxCtx, installedModule); err != nil {
					return nil, err
				}
				newCount++
				continue
			}

			modified := dbModule.ModifiedFields(installedModule)
			if modified == nil {
				continue
			}
			if err := this.syncUpdateModule(trxCtx, dbModule, modified); err != nil {
				return nil, err
			}
			modifiedCount++
		}

		for name, dbModule := range dbMap {
			if _, exists := installedMap[name]; exists {
				continue
			}
			if dbModule.GetIsOrphaned() != nil && *dbModule.GetIsOrphaned() {
				continue
			}

			changes := models.NewModuleMetadata()
			changes.SetIsOrphaned(util.ToPtr(true))
			if err := this.syncUpdateModule(trxCtx, dbModule, changes); err != nil {
				return nil, err
			}
			orphanedCount++
		}

		this.logger.Info("Sync module status", logging.Attr{
			"installed_modules": len(installedModules),
			"new_modules":       newCount,
			"modified_modules":  modifiedCount,
			"orphaned_modules":  orphanedCount,
		})

		return util.ToPtr(true), nil
	})
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, nil
	}
	return *result, nil
}

func (this *ModuleDomainServiceImpl) validateModuleCreate(
	ctx corectx.Context,
	module *models.ModuleMetadata,
	vErrs *ft.ClientErrors,
) error {
	err := this.validateAndNormalizeModuleFields(ctx, module, vErrs)
	if err != nil {
		return err
	}
	return nil
}

func (this *ModuleDomainServiceImpl) validateModuleUpdate(
	ctx corectx.Context,
	module *models.ModuleMetadata,
	_ *models.ModuleMetadata,
	vErrs *ft.ClientErrors,
) error {
	err := this.validateAndNormalizeModuleFields(ctx, module, vErrs)
	if err != nil {
		return err
	}
	return nil
}

func (this *ModuleDomainServiceImpl) validateAndNormalizeModuleFields(
	ctx corectx.Context,
	module *models.ModuleMetadata,
	vErrs *ft.ClientErrors,
) error {
	if module.GetName() != nil {
		if err := val.ApiBased.ValidateRaw(*module.GetName(), model.ModelRuleModuleLogicalName); err != nil {
			vErrs.Append(*ft.NewValidationError(
				models.ModuleMetadataFieldName,
				"module.invalid_name",
				"name must contain only letters, numbers, underscores, and dots",
			))
		}
	}

	rawVersion := module.GetFieldData().GetString(models.ModuleMetadataFieldVersion)
	if rawVersion != nil {
		if _, err := semver.ParseSemVer(*rawVersion); err != nil {
			vErrs.Append(*ft.NewValidationError(
				models.ModuleMetadataFieldVersion,
				"module.invalid_version",
				"version must use valid semantic version format",
			))
		}
	}

	if module.GetLabel() != nil {
		whitelistLangs, err := this.getEnabledLanguages(ctx)
		if err != nil {
			return err
		}

		newLabel, fieldCount, err := module.GetLabel().SanitizeClone(whitelistLangs, false)
		if err != nil {
			return err
		}
		if fieldCount == 0 {
			vErrs.Append(*ft.NewValidationError(
				models.ModuleMetadataFieldLabel,
				"module.label_requires_enabled_language",
				"label must contain at least one enabled language",
			))
		} else {
			module.SetLabel(newLabel)
		}
	}
	return nil
}

func (this *ModuleDomainServiceImpl) getEnabledLanguages(ctx corectx.Context) ([]model.LanguageCode, error) {
	query := i18n.ListEnabledLangCodesQuery{}
	result := i18n.ListEnabledLangCodesResult{}
	err := this.cqrsBus.Request(ctx, query, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (this *ModuleDomainServiceImpl) listAllModules(ctx corectx.Context) ([]models.ModuleMetadata, error) {
	modulesOut := make([]models.ModuleMetadata, 0)
	page := 0

	for {
		result, err := this.moduleRepo.Search(ctx, dyn.RepoSearchParam{
			Page: page,
			Size: model.MODEL_RULE_PAGE_MAX_SIZE,
		})
		if err != nil {
			return nil, err
		}
		if result.ClientErrors.Count() > 0 {
			return nil, fmt.Errorf("search module metadata: %v", result.ClientErrors)
		}
		if !result.HasData || len(result.Data.Items) == 0 {
			break
		}

		modulesOut = append(modulesOut, result.Data.Items...)
		if len(modulesOut) >= result.Data.Total {
			break
		}
		page++
	}

	return modulesOut, nil
}

func (this *ModuleDomainServiceImpl) syncCreateModule(
	ctx corectx.Context,
	installedModule modules.InCodeModule,
) error {
	label := make(model.LangJson)
	label[model.LabelRefLanguageCode] = installedModule.LabelKey()

	cmd := it.CreateModuleCommand{*models.NewModuleMetadata()}
	cmd.SetLabel(&label)
	cmd.SetName(util.ToPtr(installedModule.Name()))
	cmd.SetIsInternal(util.ToPtr(installedModule.IsInternal()))
	cmd.SetVersion(util.ToPtr(installedModule.Version()))

	result, err := this.CreateModule(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return fmt.Errorf("create module %s: %v", installedModule.Name(), result.ClientErrors)
	}
	return nil
}

func (this *ModuleDomainServiceImpl) syncUpdateModule(
	ctx corectx.Context,
	dbModule models.ModuleMetadata,
	changes *models.ModuleMetadata,
) error {
	cmd := it.UpdateModuleCommand{}
	cmd.SetFieldData(changes.GetFieldData())
	cmd.SetId(dbModule.GetId())
	cmd.SetEtag(dbModule.GetEtag())

	result, err := this.UpdateModule(ctx, cmd)
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		moduleName := "unknown"
		if dbModule.GetName() != nil {
			moduleName = *dbModule.GetName()
		}
		return fmt.Errorf("update module %s: %v", moduleName, result.ClientErrors.ToError())
	}
	return nil
}

func (this *ModuleDomainServiceImpl) UpdateModule(
	ctx corectx.Context, cmd it.UpdateModuleCommand,
) (*it.UpdateModuleResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.ModuleMetadata, *models.ModuleMetadata]{
		Action:        "update module metadata",
		DbRepoGetter:  this.moduleRepo,
		Data:          cmd,
		ValidateExtra: this.validateModuleUpdate,
	})
}
