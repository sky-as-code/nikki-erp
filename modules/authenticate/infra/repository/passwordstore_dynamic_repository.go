package repository

import (
	"time"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/authenticate/domain"
	it "github.com/sky-as-code/nikki-erp/modules/authenticate/interfaces/password"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"go.uber.org/dig"
)

type PasswordStoreDynamicRepositoryParam struct {
	dig.In

	Client       orm.DbClient
	ConfigSvc    config.ConfigService
	QueryBuilder orm.QueryBuilder
	Logger       logging.LoggerService
}

func NewPasswordStoreDynamicRepository(param PasswordStoreDynamicRepositoryParam) it.PasswordStoreRepository {
	schema := dmodel.MustGetSchema(domain.PasswordStoreSchemaName)
	dynamicRepo := baserepo.NewBaseDynamicRepository(baserepo.NewBaseRepositoryParam{
		Client:       param.Client,
		ConfigSvc:    param.ConfigSvc,
		QueryBuilder: param.QueryBuilder,
		Logger:       param.Logger,
		Schema:       schema,
	})
	return &PasswordStoreDynamicRepository{dynamicRepo: dynamicRepo}
}

type PasswordStoreDynamicRepository struct {
	dynamicRepo dyn.BaseDynamicRepository
}

func (this *PasswordStoreDynamicRepository) GetBaseRepo() dyn.BaseDynamicRepository {
	return this.dynamicRepo
}

func (this *PasswordStoreDynamicRepository) Insert(ctx corectx.Context, pass domain.PasswordStore) (*dyn.OpResult[int], error) {
	data := cloneFields(pass.GetFieldData())
	if pass.GetPassword() != nil {
		data[domain.PasswordStoreFieldPassword] = *pass.GetPassword()
		if pass.GetPasswordUpdatedAt() != nil {
			data[domain.PasswordStoreFieldPasswordUpdatedAt] = domain.TimePtrToModelDateTime(pass.GetPasswordUpdatedAt())
		}
	}
	pass.SetFieldData(data)
	return baserepo.Insert(ctx, this.dynamicRepo, &pass)
}

func (this *PasswordStoreDynamicRepository) Update(
	ctx corectx.Context, pass domain.PasswordStore,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	data := dmodel.DynamicFields{}
	data[basemodel.FieldId] = string(*pass.GetId())
	if pass.GetPassword() != nil {
		data[domain.PasswordStoreFieldPassword] = *pass.GetPassword()
		data[domain.PasswordStoreFieldPasswordUpdatedAt] = domain.TimePtrToModelDateTime(util.ToPtr(time.Now()))
	}
	if pass.GetPasswordExpiredAt() != nil {
		if !model.ZeroTime.Equal(*pass.GetPasswordExpiredAt()) {
			data[domain.PasswordStoreFieldPasswordExpiredAt] = domain.TimePtrToModelDateTime(pass.GetPasswordExpiredAt())
		} else {
			data[domain.PasswordStoreFieldPasswordExpiredAt] = nil
		}
	}
	if pass.GetPasswordotp() != nil {
		if len(*pass.GetPasswordotp()) > 0 {
			data[domain.PasswordStoreFieldPasswordotp] = *pass.GetPasswordotp()
			data[domain.PasswordStoreFieldPasswordotpExpiredAt] = domain.TimePtrToModelDateTime(pass.GetPasswordotpExpiredAt())
		} else {
			data[domain.PasswordStoreFieldPasswordotp] = nil
			data[domain.PasswordStoreFieldPasswordotpExpiredAt] = nil
		}
	}
	if pass.GetPasswordotpExpiredAt() != nil && model.ZeroTime.Equal(*pass.GetPasswordotpExpiredAt()) {
		data[domain.PasswordStoreFieldPasswordotpExpiredAt] = nil
	}
	if pass.GetPasswordotpRecovery() != nil {
		if len(pass.GetPasswordotpRecovery()) > 0 {
			data[domain.PasswordStoreFieldPasswordotpRecovery] = pass.GetPasswordotpRecovery()
		} else {
			data[domain.PasswordStoreFieldPasswordotpRecovery] = nil
		}
	}
	if pass.GetPasswordtmp() != nil {
		if len(*pass.GetPasswordtmp()) > 0 {
			data[domain.PasswordStoreFieldPasswordtmp] = *pass.GetPasswordtmp()
			data[domain.PasswordStoreFieldPasswordtmpExpiredAt] = domain.TimePtrToModelDateTime(pass.GetPasswordtmpExpiredAt())
		} else {
			data[domain.PasswordStoreFieldPasswordtmp] = nil
			data[domain.PasswordStoreFieldPasswordtmpExpiredAt] = nil
		}
	}
	return baserepo.Update(ctx, this.dynamicRepo, data)
}

func (this *PasswordStoreDynamicRepository) GetOne(
	ctx corectx.Context, param dyn.RepoGetOneParam,
) (*dyn.OpResult[domain.PasswordStore], error) {
	return baserepo.GetOne[domain.PasswordStore](ctx, this.dynamicRepo, param)
}

func cloneFields(src dmodel.DynamicFields) dmodel.DynamicFields {
	if src == nil {
		return make(dmodel.DynamicFields)
	}
	out := make(dmodel.DynamicFields, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
