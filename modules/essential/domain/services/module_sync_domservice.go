package services

import (
	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

// SyncModuleMetadata brings essential_module_metadata in step with the modules this binary
// loaded: new modules are created, changed ones updated, vanished ones marked orphaned. An
// advisory lock keeps two instances booting together from racing, and the whole pass is one
// transaction.
func (this *ModuleDomainServiceImpl) SyncModuleMetadata(
	ctx corectx.Context, installedModules []modules.InCodeModule,
) (isSuccess bool, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "sync module metadata"); e != nil {
			err = e
		}
	}()

	acquired, err := this.moduleRepo().AcquireLock(ctx)
	if err != nil {
		return false, err
	}
	if !acquired {
		this.logger.Debugf("Could not acquire lock to sync module metadata. Another instance has acquired it")
		return false, nil
	}
	defer func() {
		this.logger.Debugf("Releasing lock for module metadata sync")
		if releaseErr := this.moduleRepo().ReleaseLock(ctx); releaseErr != nil {
			err = errors.Wrap(releaseErr, "release module metadata lock")
		}
	}()

	err = composable.WithTransaction(ctx, this.Repository(), func(tranxCtx corectx.Context) error {
		return this.syncInTranx(tranxCtx, installedModules)
	})
	return err == nil, err
}

func (this *ModuleDomainServiceImpl) syncInTranx(ctx corectx.Context, installedModules []modules.InCodeModule) error {
	dbModules, err := this.listAllModules(ctx)
	if err != nil {
		return err
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

	levels, err := computeDepLevels(installedModules)
	if err != nil {
		return err
	}

	newCount, modifiedCount, err := this.syncInstalled(ctx, installedMap, dbMap, levels)
	if err != nil {
		return err
	}
	orphanedCount, err := this.syncOrphaned(ctx, installedMap, dbMap)
	if err != nil {
		return err
	}
	this.logger.Info("Sync module status", logging.Attr{
		"installed_modules": len(installedModules),
		"new_modules":       newCount,
		"modified_modules":  modifiedCount,
		"orphaned_modules":  orphanedCount,
	})
	return nil
}

func (this *ModuleDomainServiceImpl) syncInstalled(
	ctx corectx.Context, installedMap map[string]modules.InCodeModule, dbMap map[string]models.ModuleMetadata,
	levels map[string]int32,
) (newCount int, modifiedCount int, err error) {
	for name, installedModule := range installedMap {
		dbModule, exists := dbMap[name]
		if !exists {
			if err := this.syncCreateModule(ctx, installedModule, levels[name]); err != nil {
				return newCount, modifiedCount, err
			}
			newCount++
			continue
		}
		modified := dbModule.ModifiedFields(installedModule, levels[name])
		if modified == nil {
			continue
		}
		if err := this.syncUpdateModule(ctx, dbModule, modified); err != nil {
			return newCount, modifiedCount, err
		}
		modifiedCount++
	}
	return newCount, modifiedCount, nil
}

func (this *ModuleDomainServiceImpl) syncOrphaned(
	ctx corectx.Context, installedMap map[string]modules.InCodeModule, dbMap map[string]models.ModuleMetadata,
) (int, error) {
	orphanedCount := 0
	for name, dbModule := range dbMap {
		if _, exists := installedMap[name]; exists {
			continue
		}
		if dbModule.GetIsOrphaned() != nil && *dbModule.GetIsOrphaned() {
			continue
		}
		changes := models.NewModuleMetadata()
		changes.SetIsOrphaned(util.ToPtr(true))
		if err := this.syncUpdateModule(ctx, dbModule, changes); err != nil {
			return orphanedCount, err
		}
		orphanedCount++
	}
	return orphanedCount, nil
}

func (this *ModuleDomainServiceImpl) listAllModules(ctx corectx.Context) ([]models.ModuleMetadata, error) {
	modulesOut := make([]models.ModuleMetadata, 0)
	for page := 0; ; page++ {
		result, err := this.Repository().Search(ctx, dyn.RepoSearchParam{Page: page, Size: model.MODEL_RULE_PAGE_MAX_SIZE})
		if err != nil {
			return nil, err
		}
		if result.ClientErrors.Count() > 0 {
			return nil, errors.Wrap(result.ClientErrors.ToError(), "search module metadata")
		}
		if !result.HasData || len(result.Data.Items) == 0 {
			break
		}
		for _, row := range result.Data.Items {
			modulesOut = append(modulesOut, *models.NewModuleMetadataFrom(row))
		}
		if len(modulesOut) >= result.Data.Total {
			break
		}
	}
	return modulesOut, nil
}

func (this *ModuleDomainServiceImpl) syncCreateModule(
	ctx corectx.Context, installedModule modules.InCodeModule, depLevel int32,
) error {
	label := make(model.LangJson)
	label[model.LanguageCodeRef] = installedModule.LabelKey()

	module := models.NewModuleMetadata()
	module.SetLabel(&label)
	module.SetName(util.ToPtr(installedModule.Name()))
	module.SetIsInternal(util.ToPtr(installedModule.IsInternal()))
	module.SetVersion(util.ToPtr(installedModule.Version()))
	module.SetHasUi(util.ToPtr(installedModule.HasUi()))
	module.SetModelPrefix(util.ToPtr(installedModule.ModelPrefix()))
	module.SetCategory(util.ToPtr(installedModule.Category()))
	module.SetIsEnabled(util.ToPtr(true))
	module.SetDependencies(installedModule.Deps())
	module.SetDepLevel(&depLevel)

	result, err := this.Create(ctx, it.CreateModuleCommand(module.GetFieldData()))
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrapf(result.ClientErrors.ToError(), "create module %s", installedModule.Name())
	}
	return nil
}

func (this *ModuleDomainServiceImpl) syncUpdateModule(
	ctx corectx.Context, dbModule models.ModuleMetadata, changes *models.ModuleMetadata,
) error {
	update := models.NewModuleMetadataFrom(changes.GetFieldData())
	update.SetId(dbModule.GetId())
	update.SetEtag(dbModule.GetEtag())

	result, err := this.Update(ctx, it.UpdateModuleCommand(update.GetFieldData()))
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return errors.Wrapf(result.ClientErrors.ToError(), "update module %s", util.ValueOrZeroOf(dbModule.GetName()))
	}
	return nil
}
