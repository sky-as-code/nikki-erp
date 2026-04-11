package password

import (
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

type PasswordStoreRepository interface {
	dyn.DynamicModelRepository

	BeginTransaction(ctx corectx.Context) (database.DbTransaction, error)
	DeleteOne(ctx corectx.Context, keys domain.PasswordStore) (*dyn.OpResult[dyn.MutateResultData], error)
	Insert(ctx corectx.Context, store domain.PasswordStore) (*dyn.OpResult[int], error)
	GetOne(ctx corectx.Context, param dyn.RepoGetOneParam) (*dyn.OpResult[domain.PasswordStore], error)
	Update(ctx corectx.Context, store domain.PasswordStore) (*dyn.OpResult[dyn.MutateResultData], error)
}
