package app

import (
	"context"
	"fmt"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	i18n "github.com/sky-as-code/nikki-erp/modules/core/i18n/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

func NewModuleServiceImpl(
	cqrsBus cqrs.CqrsBus,
	eventBus event.EventBus,
	logger logging.LoggerService,
	moduleRepo it.ModuleRepository,
) it.ModuleService {
	return &ModuleServiceImpl{
		cqrsBus:    cqrsBus,
		eventBus:   eventBus,
		logger:     logger,
		moduleRepo: moduleRepo,
	}
}

type ModuleServiceImpl struct {
	cqrsBus    cqrs.CqrsBus
	eventBus   event.EventBus
	logger     logging.LoggerService
	moduleRepo it.ModuleRepository
}

func (this *ModuleServiceImpl) CreateModule(ctx crud.Context, cmd it.CreateModuleCommand) (*it.CreateModuleResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.ModuleMetadata, it.CreateModuleCommand, it.CreateModuleResult]{
		Action:              "create module",
		Command:             cmd,
		AssertBusinessRules: this.assertModuleUnique,
		RepoCreate:          this.moduleRepo.Create,
		SetDefault:          this.setModuleDefaults,
		Sanitize:            this.sanitizeModule,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateModuleResult {
			return &it.CreateModuleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ModuleMetadata) *it.CreateModuleResult {
			return &it.CreateModuleResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) CreateBulkModules(ctx crud.Context, cmd it.CreateBulkModulesCommand) (*it.CreateBulkModulesResult, error) {
	result, err := crud.CreateBulk(ctx, crud.CreateBulkParam[*domain.ModuleMetadata, it.CreateBulkModulesCommand, it.CreateBulkModulesResult]{
		Action:              "create bulk modules",
		Command:             cmd,
		AssertBusinessRules: this.assertModuleUnique,
		RepoCreateBulk:      this.moduleRepo.CreateBulk,
		SetDefault:          this.setModuleDefaults,
		Sanitize:            this.sanitizeModule,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateBulkModulesResult {
			return &it.CreateBulkModulesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model []*domain.ModuleMetadata) *it.CreateBulkModulesResult {
			return &it.CreateBulkModulesResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) UpdateModule(ctx crud.Context, cmd it.UpdateModuleCommand) (*it.UpdateModuleResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.ModuleMetadata, it.UpdateModuleCommand, it.UpdateModuleResult]{
		Action:              "update module",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertModuleIdExists,
		RepoUpdate:          this.moduleRepo.Update,
		Sanitize:            this.sanitizeModule,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateModuleResult {
			return &it.UpdateModuleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ModuleMetadata) *it.UpdateModuleResult {
			return &it.UpdateModuleResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) UpdateBulkModules(ctx crud.Context, cmd it.UpdateBulkModulesCommand) (*it.UpdateBulkModulesResult, error) {
	result, err := crud.UpdateBulk(ctx, crud.UpdateBulkParam[*domain.ModuleMetadata, it.UpdateBulkModulesCommand, it.UpdateBulkModulesResult]{
		Action:              "update bulk modules",
		Command:             cmd,
		AssertBusinessRules: this.assertUpdateRules,
		AssertExists:        this.assertModuleIdExists,
		RepoUpdate:          this.moduleRepo.Update,
		Sanitize:            this.sanitizeModule,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateBulkModulesResult {
			return &it.UpdateBulkModulesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model []*domain.ModuleMetadata) *it.UpdateBulkModulesResult {
			return &it.UpdateBulkModulesResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

// func (this *ModuleServiceImpl) sanitizeModule(module *it.Module, vErrs *ft.ValidationErrors) {
func (this *ModuleServiceImpl) sanitizeModule(module *domain.ModuleMetadata) {
	// TODO: Should pass ctx from outside
	ctx := crud.NewRequestContext(context.Background())
	whitelistLangs, err := this.getEnabledLanguages(ctx)
	ft.PanicOnErr(err)

	newLabel, _, err := module.Label.SanitizeClone(whitelistLangs, false)
	ft.PanicOnErr(err)

	// TODO: Should pass vErrs from outside
	// if fieldCount == 0 {
	// 	vErrs.Append("label", "no enabled language")
	// }
	module.Label = newLabel

	if module.Name != nil {
		module.Name = defense.SanitizePlainTextPtr(module.Name, true)
	}
}

func (this *ModuleServiceImpl) assertUpdateRules(ctx crud.Context, updatedMod *domain.ModuleMetadata, _ *domain.ModuleMetadata, vErrs *ft.ValidationErrors) error {
	return this.assertModuleUnique(ctx, updatedMod, vErrs)
}

func (this *ModuleServiceImpl) setModuleDefaults(module *domain.ModuleMetadata) {
	module.SetDefaults()
	module.IsOrphaned = util.ToPtr(false)
}

func (this *ModuleServiceImpl) assertModuleUnique(ctx crud.Context, module *domain.ModuleMetadata, vErrs *ft.ValidationErrors) error {
	dbMod, err := this.moduleRepo.FindByName(ctx, it.FindByNameParam{Name: *module.Name})
	if err != nil {
		return err
	}

	if dbMod != nil {
		vErrs.AppendAlreadyExists("name", "module name")
	}
	return nil
}

func (this *ModuleServiceImpl) assertModuleIdExists(ctx crud.Context, module *domain.ModuleMetadata, vErrs *ft.ValidationErrors) (dbModule *domain.ModuleMetadata, err error) {
	dbModule, err = this.moduleRepo.FindById(ctx, it.FindByIdParam{Id: *module.Id})
	if dbModule == nil {
		vErrs.AppendNotFound("id", "module ID")
	}
	return
}

func (this *ModuleServiceImpl) getEnabledLanguages(ctx crud.Context) ([]model.LanguageCode, error) {
	query := i18n.ListEnabledLangCodesQuery{}
	result := i18n.ListEnabledLangCodesResult{}
	err := this.cqrsBus.Request(ctx, query, &result)
	ft.PanicOnErr(err)

	return result.Data, nil
}

func (this *ModuleServiceImpl) DeleteModule(ctx crud.Context, cmd it.DeleteModuleCommand) (*it.DeleteModuleResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.ModuleMetadata, it.DeleteModuleCommand, it.DeleteModuleResult]{
		Action:       "delete module",
		Command:      cmd,
		AssertExists: this.assertModuleIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.ModuleMetadata) (int, error) {
			return this.moduleRepo.DeleteById(ctx, it.DeleteByIdParam{Id: *model.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteModuleResult {
			return &it.DeleteModuleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ModuleMetadata, deletedCount int) *it.DeleteModuleResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) ModuleExists(ctx crud.Context, query it.ModuleExistsQuery) (*it.ModuleExistsResult, error) {
	result, err := crud.ExistsOne(ctx, crud.ExistsOneParam[*domain.ModuleMetadata, it.ModuleExistsQuery, it.ModuleExistsResult]{
		Action: "check if module exists",
		Query:  query,
		RepoExistsOne: func(ctx crud.Context, query it.ModuleExistsQuery, vErrs *ft.ValidationErrors) (bool, error) {
			return this.moduleRepo.Exists(ctx, it.ExistsParam{Id: query.Id})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.ModuleExistsResult {
			return &it.ModuleExistsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: crud.NewSuccessExistsResult,
	})

	return result, err
}

func (this *ModuleServiceImpl) GetModule(ctx crud.Context, query it.GetModuleByIdQuery) (*it.GetModuleResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.ModuleMetadata, it.GetModuleByIdQuery, it.GetModuleResult]{
		Action:      "get module by ID",
		Query:       query,
		RepoFindOne: this.getModuleByIdFull,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetModuleResult {
			return &it.GetModuleResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.ModuleMetadata) *it.GetModuleResult {
			return &it.GetModuleResult{
				Data:    model,
				HasData: model != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) getModuleByIdFull(ctx crud.Context, query it.GetModuleByIdQuery, vErrs *ft.ValidationErrors) (dbModule *domain.ModuleMetadata, err error) {
	dbModule, err = this.moduleRepo.FindById(ctx, query)
	if dbModule == nil {
		vErrs.AppendNotFound("id", "module ID")
	}
	return
}

func (this *ModuleServiceImpl) ListModules(ctx crud.Context, query it.ListModulesQuery) (*it.ListModulesResult, error) {
	result, err := crud.ListAll(ctx, crud.ListAllParam[domain.ModuleMetadata, it.ListModulesQuery, it.ListModulesResult]{
		Action: "get module by ID",
		Query:  query,
		RepoListAll: func(ctx crud.Context, query it.ListModulesQuery, vErrs *ft.ValidationErrors) ([]domain.ModuleMetadata, error) {
			return this.moduleRepo.List(ctx, it.ListParam{})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.ListModulesResult {
			return &it.ListModulesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(models []domain.ModuleMetadata) *it.ListModulesResult {
			return &it.ListModulesResult{
				Data:    models,
				HasData: models != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) SearchModules(ctx crud.Context, query it.SearchModulesQuery) (*it.SearchModulesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.ModuleMetadata, it.SearchModulesQuery, it.SearchModulesResult]{
		Action: "search modules",
		Query:  query,
		SetQueryDefaults: func(query *it.SearchModulesQuery) {
			query.SetDefaults()
		},
		ParseSearchGraph: this.moduleRepo.ParseSearchGraph,
		RepoSearch: func(ctx crud.Context, query it.SearchModulesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.ModuleMetadata], error) {
			return this.moduleRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchModulesResult {
			return &it.SearchModulesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(pagedResult *crud.PagedResult[domain.ModuleMetadata]) *it.SearchModulesResult {
			return &it.SearchModulesResult{
				Data:    pagedResult,
				HasData: pagedResult.Items != nil,
			}
		},
	})

	return result, err
}

func (this *ModuleServiceImpl) SyncModuleMetadata(ctx crud.Context, installedModules []modules.InCodeModule) (isSuccess bool, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "sync module metadata"); e != nil {
			err = e
		}
	}()

	// Prevent other replicas to run this method.
	accquired, err := this.moduleRepo.AcquireLock(ctx)
	ft.PanicOnErr(err)
	if !accquired {
		this.logger.Infof("Could not acquired lock to sync module metadata. Another instance has acquired it")
		return false, nil
	}
	this.logger.Infof("Acquired lock successfully to sync module metadata")

	defer this.moduleRepo.ReleaseLock(ctx)

	trxCtx, err := this.moduleRepo.IncludeTransaction(ctx)
	ft.PanicOnErr(err)
	defer trxCtx.GetDbTranx().Commit()

	dbModules, err := this.moduleRepo.List(trxCtx, it.ListParam{})
	ft.PanicOnErr(err)

	// INSERT_YOUR_CODE

	// Build maps for quick lookup
	dbMap := make(map[string]domain.ModuleMetadata)
	for _, m := range dbModules {
		dbMap[*m.Name] = m
	}

	installedMap := make(map[string]modules.InCodeModule)
	for _, m := range installedModules {
		installedMap[m.Name()] = m
	}

	var orphanedMods []domain.ModuleMetadata
	var modifiedMods []domain.ModuleMetadata
	var newMods []modules.InCodeModule

	for name, dbMod := range dbMap {
		if installedMod, ok := installedMap[name]; ok {
			modified := dbMod.ModifiedFields(installedMod)
			if modified != nil {
				modified.Id = dbMod.Id
				modifiedMods = append(modifiedMods, *modified)
			}
			// if installedMod.Version() != *dbMod.Version {
			// 	modifiedMods = append(modifiedMods, dbMod)
			// }
		} else {
			orphanedMods = append(orphanedMods, dbMod)
		}
	}

	for name, im := range installedMap {
		if _, ok := dbMap[name]; !ok {
			newMods = append(newMods, im)
		}
	}

	this.logger.Info("Sync module status", logging.Attr{
		"installed_modules": len(installedModules),
		"new_modules":       len(newMods),
		"modified_modules":  len(modifiedMods),
		"orphaned_modules":  len(orphanedMods),
	})

	createCmds := array.Map(newMods, func(mod modules.InCodeModule) it.CreateModuleCommand {
		label := make(model.LangJson)
		label[model.LabelRefLanguageCode] = mod.LabelKey()
		return it.CreateModuleCommand{
			Label:   label,
			Name:    mod.Name(),
			Version: util.ToPtr(mod.Version()).String(),
		}
	})
	result, err := this.CreateBulkModules(trxCtx, it.CreateBulkModulesCommand{
		Modules: createCmds,
	})
	ft.PanicOnErr(err)

	if result.ClientError != nil {
		return false, fmt.Errorf("failed to create new modules: %v", result.ClientError)
	}

	updatedCmds := array.Map(modifiedMods, func(mod domain.ModuleMetadata) it.UpdateModuleCommand {
		cmd := it.UpdateModuleCommand{
			Id: *mod.Id,
		}
		if mod.Label != nil {
			cmd.Label = mod.Label
		}
		if mod.Version != nil {
			cmd.Version = util.ToPtr(mod.Version.String())
		}
		return cmd
	})
	orphanedCmds := array.Map(orphanedMods, func(mod domain.ModuleMetadata) it.UpdateModuleCommand {
		return it.UpdateModuleCommand{
			Id:         *mod.Id,
			IsOrphaned: util.ToPtr(true),
		}
	})

	updatedCmds = append(updatedCmds, orphanedCmds...)
	// result, err = this.UpdateBulkModules(trxCtx, it.UpdateBulkModulesCommand{
	// 	Modules: updatedCmds,
	// })
	// ft.PanicOnErr(err)

	// if result.ClientError != nil {
	// 	return false, fmt.Errorf("failed to update modified modules: %v", result.ClientError)
	// }

	return true, nil
}

// func (this *ModuleServiceImpl) sprintModuleMetadata(mods []domain.ModuleMetadata) string {
// 	modStr := []string{}
// 	for _, m := range mods {
// 		modStr = append(modStr, fmt.Sprintf("%s %s", *m.Name, *m.Version))
// 	}
// 	return strings.Join(modStr, ", ")
// }

// func (this *ModuleServiceImpl) sprintInCodeModule(mods []modules.InCodeModule) string {
// 	modStr := []string{}
// 	for _, m := range mods {
// 		modStr = append(modStr, fmt.Sprintf("%s %s", m.Name(), m.Version()))
// 	}
// 	return strings.Join(modStr, ", ")
// }
