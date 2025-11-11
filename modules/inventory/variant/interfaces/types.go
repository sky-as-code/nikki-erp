package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type VariantRepository interface {
	Create(ctx crud.Context, variant *Variant) (*Variant, error)
	Update(ctx crud.Context, variant *Variant, prevEtag model.Etag) (*Variant, error)
	DeleteById(ctx crud.Context, id model.Id) (int, error)
	FindById(ctx crud.Context, query FindByIdParam) (*Variant, error)
	ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	Search(ctx crud.Context, param SearchParam) (*crud.PagedResult[Variant], error)
}

type VariantService interface {
	CreateVariant(ctx crud.Context, cmd CreateVariantCommand) (*CreateVariantResult, error)
	UpdateVariant(ctx crud.Context, cmd UpdateVariantCommand) (*UpdateVariantResult, error)
	DeleteVariant(ctx crud.Context, cmd DeleteVariantCommand) (*DeleteVariantResult, error)
	GetVariantById(ctx crud.Context, query GetVariantByIdQuery) (*GetVariantByIdResult, error)
	SearchVariants(ctx crud.Context, query SearchVariantsQuery) (*SearchVariantsResult, error)
}

type DeleteParam = DeleteVariantCommand
type FindByIdParam = GetVariantByIdQuery
type SearchParam struct {
	Predicate *orm.Predicate
	Order     []orm.OrderOption
	Page      int
	Size      int
}
