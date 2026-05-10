package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/enum"
)

func NewEnumServiceImpl(enumRepo it.EnumRepository) it.EnumService {
	return &EnumServiceImpl{
		enumRepo: enumRepo,
	}
}

type EnumServiceImpl struct {
	enumRepo it.EnumRepository
}

func (this *EnumServiceImpl) CreateEnum(ctx corectx.Context, cmd it.CreateEnumCommand) (*it.CreateEnumResult, error) {
	return corecrud.Create(ctx, corecrud.CreateParam[models.Enum, *models.Enum]{
		Action:         "create enum",
		BaseRepoGetter: this.enumRepo,
		Data:           cmd,
	})
}

func (this *EnumServiceImpl) UpdateEnum(ctx corectx.Context, cmd it.UpdateEnumCommand) (*it.UpdateEnumResult, error) {
	return corecrud.Update(ctx, corecrud.UpdateParam[models.Enum, *models.Enum]{
		Action:       "update enum",
		DbRepoGetter: this.enumRepo,
		Data:         cmd,
	})
}

func (this *EnumServiceImpl) DeleteEnum(ctx corectx.Context, cmd it.DeleteEnumCommand) (*it.DeleteEnumResult, error) {
	return corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:       "delete enum",
		DbRepoGetter: this.enumRepo,
		Cmd:          dyn.DeleteOneCommand(cmd),
	})
}

func (this *EnumServiceImpl) GetEnum(ctx corectx.Context, query it.GetEnumQuery) (*it.GetEnumResult, error) {
	return corecrud.GetOne[models.Enum](ctx, corecrud.GetOneParam{
		Action:       "get enum",
		DbRepoGetter: this.enumRepo,
		Query:        dyn.GetOneQuery(query),
	})
}

func (this *EnumServiceImpl) SearchEnums(ctx corectx.Context, query it.SearchEnumsQuery) (*it.SearchEnumsResult, error) {
	return corecrud.Search[models.Enum](ctx, corecrud.SearchParam{
		Action:       "search enums",
		DbRepoGetter: this.enumRepo,
		Query:        dyn.SearchQuery(query),
	})
}

func (this *EnumServiceImpl) EnumExists(ctx corectx.Context, query it.EnumExistsQuery) (*it.EnumExistsResult, error) {
	return corecrud.Exists(ctx, corecrud.ExistsParam{
		Action:       "check if enum exists",
		DbRepoGetter: this.enumRepo,
		Query:        dyn.ExistsQuery(query),
	})
}
