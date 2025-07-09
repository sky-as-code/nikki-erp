package enum

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/model"
)

type EnumRepository interface {
	Create(ctx context.Context, enum Enum) (*Enum, error)
	DeleteById(ctx context.Context, id model.Id) (int, error)
	DeleteByType(ctx context.Context, enumType string) (int, error)
	Exists(ctx context.Context, id model.Id) (bool, error)
	ExistsMulti(ctx context.Context, ids []model.Id) (existing []model.Id, notExisting []model.Id, err error)
	FindById(ctx context.Context, id model.Id) (*Enum, error)
	FindByValue(ctx context.Context, value string, enumType string) (*Enum, error)
	List(ctx context.Context, param ListParam) (*crud.PagedResult[Enum], error)
	Update(ctx context.Context, enum Enum) (*Enum, error)
}

type ListParam = ListEnumsQuery

type EnumService interface {
	CreateEnum(ctx context.Context, cmd CreateEnumCommand) (*CreateEnumResult, error)
	DeleteEnum(ctx context.Context, cmd DeleteEnumCommand) (*DeleteEnumResult, error)
	Exists(ctx context.Context, cmd EnumExistsCommand) (*EnumExistsResult, error)
	ExistsMulti(ctx context.Context, cmd EnumExistsMultiCommand) (*EnumExistsMultiResult, error)
	GetEnum(ctx context.Context, query GetEnumQuery) (result *GetEnumResult, err error)
	ListEnums(ctx context.Context, query ListEnumsQuery) (result *ListEnumsResult, err error)
	UpdateEnum(ctx context.Context, cmd UpdateEnumCommand) (*UpdateEnumResult, error)
}
