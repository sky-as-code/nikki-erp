package repository

import (
	"math"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

const moduleMetadataSyncLockKey = math.MaxInt64

// NewModuleRepository is handed the composable default by the module metadata onion.
func NewModuleRepository(base composable.CrudRepository) it.ModuleRepository {
	return &ModuleRepositoryImpl{CrudRepository: base}
}

type ModuleRepositoryImpl struct {
	composable.CrudRepository
}

func (this *ModuleRepositoryImpl) AcquireLock(ctx corectx.Context) (bool, error) {
	var acquired bool
	rows, err := this.GetBaseRepo().QueryFunc(ctx, "pg_try_advisory_lock", moduleMetadataSyncLockKey)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&acquired)
	}
	return acquired, err
}

func (this *ModuleRepositoryImpl) ReleaseLock(ctx corectx.Context) error {
	return this.GetBaseRepo().ExecFunc(ctx, "pg_advisory_unlock", moduleMetadataSyncLockKey)
}
