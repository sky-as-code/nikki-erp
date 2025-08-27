package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type EnumRepository interface {
	Create(ctx crud.Context, enum Enum) (*Enum, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	DeleteByType(ctx crud.Context, enumType string) (int, error)
	Exists(ctx crud.Context, id model.Id) (bool, error)
	ExistsMulti(ctx crud.Context, ids []model.Id) (existing []model.Id, notExisting []model.Id, err error)
	FindById(ctx crud.Context, id model.Id) (*Enum, error)
	FindByValue(ctx crud.Context, value string, enumType string) (*Enum, error)
	List(ctx crud.Context, param ListParam) (*crud.PagedResult[Enum], error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[Enum], error)
	Update(ctx crud.Context, enum Enum, prevEtag model.Etag) (*Enum, error)
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
	CreateEnum(ctx crud.Context, cmd CreateEnumCommand) (*CreateEnumResult, error)
	DeleteEnum(ctx crud.Context, cmd DeleteEnumCommand) (*DeleteEnumResult, error)
	EnumExists(ctx crud.Context, cmd EnumExistsQuery) (*EnumExistsResult, error)
	EnumExistsMulti(ctx crud.Context, cmd EnumExistsMultiQuery) (*EnumExistsMultiResult, error)
	GetEnum(ctx crud.Context, query GetEnumQuery) (result *GetEnumResult, err error)
	ListEnums(ctx crud.Context, query ListEnumsQuery) (result *ListEnumsResult, err error)
	SearchEnums(ctx crud.Context, query SearchEnumsQuery) (result *SearchEnumsResult, err error)
	UpdateEnum(ctx crud.Context, cmd UpdateEnumCommand) (*UpdateEnumResult, error)
}
