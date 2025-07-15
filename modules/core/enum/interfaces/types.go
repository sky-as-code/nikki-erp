package interfaces

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
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
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx context.Context, param SearchParam) (*crud.PagedResult[Enum], error)
	Update(ctx context.Context, enum Enum, prevEtag model.Etag) (*Enum, error)
}

type ListParam = ListEnumsQuery
type SearchParam struct {
	Predicate  *orm.Predicate
	Order      []orm.OrderOption
	Page       int
	Size       int
	TypePrefix *string
}

type EnumService interface {
	CreateEnum(ctx context.Context, cmd CreateEnumCommand) (*CreateEnumResult, error)
	DeleteEnum(ctx context.Context, cmd DeleteEnumCommand) (*DeleteEnumResult, error)
	EnumExists(ctx context.Context, cmd EnumExistsQuery) (*EnumExistsResult, error)
	EnumExistsMulti(ctx context.Context, cmd EnumExistsMultiQuery) (*EnumExistsMultiResult, error)
	GetEnum(ctx context.Context, query GetEnumQuery) (result *GetEnumResult, err error)
	ListEnums(ctx context.Context, query ListEnumsQuery) (result *ListEnumsResult, err error)
	SearchEnums(ctx context.Context, query SearchEnumsQuery) (result *SearchEnumsResult, err error)
	UpdateEnum(ctx context.Context, cmd UpdateEnumCommand) (*UpdateEnumResult, error)
}
