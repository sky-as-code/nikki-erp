package repository

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	"github.com/sky-as-code/nikki-erp/modules/essential/infra/ent"
)

func entToModule(dbMod *ent.Module) *domain.ModuleMetadata {
	mod := &domain.ModuleMetadata{}
	model.MustCopy(dbMod, mod)
	return mod
}

func entToModules(dbModules []*ent.Module) []*domain.ModuleMetadata {
	if dbModules == nil {
		return nil
	}
	return array.Map(dbModules, func(entModule *ent.Module) *domain.ModuleMetadata {
		return entToModule(entModule)
	})
}
