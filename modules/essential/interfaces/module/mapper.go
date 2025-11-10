package module

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
)

func (this CreateModuleCommand) ToDomainModel() *domain.ModuleMetadata {
	mod := &domain.ModuleMetadata{}
	model.MustCopy(this, mod)
	return mod
}

func (this CreateBulkModulesCommand) ToDomainModels() []*domain.ModuleMetadata {
	return array.Map(this.Modules, func(cmd CreateModuleCommand) *domain.ModuleMetadata {
		return cmd.ToDomainModel()
	})
}

func (this DeleteModuleCommand) ToDomainModel() *domain.ModuleMetadata {
	mod := &domain.ModuleMetadata{}
	model.MustCopy(this, mod)
	return mod
}

func (this UpdateModuleCommand) ToDomainModel() *domain.ModuleMetadata {
	mod := &domain.ModuleMetadata{}
	model.MustCopy(this, mod)
	return mod
}

func (this UpdateBulkModulesCommand) ToDomainModels() []*domain.ModuleMetadata {
	return array.Map(this.Modules, func(cmd UpdateModuleCommand) *domain.ModuleMetadata {
		return cmd.ToDomainModel()
	})
}
