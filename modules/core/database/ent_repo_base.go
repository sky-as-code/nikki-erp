package database

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/orm"
)

type MutationBuilder[TDb any] interface {
	Save(context.Context) (*TDb, error)
}

type QueryOneBuilder[TDb any] interface {
	Only(context.Context) (*TDb, error)
}

type EntToDomainFn[TDb any, TDomain any] func(*TDb) *TDomain
type EntToDomainArrFn[TDb any, TDomain any] func([]*TDb) []*TDomain

type EntRepositoryBase struct {
}

func Mutate[TDb any, TDomain any](
	ctx context.Context,
	mutationBuilder MutationBuilder[TDb],
	isNotFoundFn func(err error) bool,
	convertFn EntToDomainFn[TDb, TDomain],
) (*TDomain, error) {
	entEntity, err := mutationBuilder.Save(ctx)
	if err != nil {
		if isNotFoundFn(err) {
			return nil, nil
		}
		return nil, err
	}

	domainEntity := convertFn(entEntity)
	return domainEntity, nil
}

func FindOne[TDb any, TDomain any](
	ctx context.Context,
	queryBuilder QueryOneBuilder[TDb],
	isNotFoundFn func(err error) bool,
	convertFn EntToDomainFn[TDb, TDomain],
) (*TDomain, error) {
	entEntity, err := queryBuilder.Only(ctx)
	if err != nil {
		if isNotFoundFn(err) {
			return nil, nil
		}
		return nil, err
	}
	return convertFn(entEntity), nil
}

func List[TDb any, TDomain any, TQuery interface {
	All(context.Context) ([]*TDb, error)
}](
	ctx context.Context,
	query TQuery,
	convertFn func([]*TDb) []TDomain,
) ([]TDomain, error) {
	entEntities, err := query.All(ctx)
	if err != nil {
		return nil, err
	}
	return convertFn(entEntities), nil
}

func ParseSearchGraphStr[TDb any, TDomain any](criteria *string, entityName string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	if criteria == nil {
		return nil, nil, nil
	}

	var graph orm.SearchGraph
	err := json.Unmarshal([]byte(*criteria), &graph)
	if err != nil {
		vErr := ft.NewValidationErrors()
		vErr.Appendf("graph", "invalid search graph: %s", err.Error())
		return nil, nil, vErr
	}

	return ParseSearchGraph[TDb, TDomain](&graph, entityName)
}

func ParseSearchGraph[TDb any, TDomain any](criteria *orm.SearchGraph, entityName string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	if criteria == nil {
		return nil, nil, nil
	}

	predicate, vErrsPre := criteria.ToPredicate(entityName)
	order, vErrsOrd := criteria.Order.ToOrderOptions(entityName)
	vErrsPre.Merge(vErrsOrd)

	return &predicate, order, vErrsPre
}

func Search[TDb any, TDomain any, TQuery interface {
	Where(...orm.Predicate) TQuery
	Clone() TQuery
	Count(context.Context) (int, error)
	Offset(int) TQuery
	Limit(int) TQuery
	Order(...orm.OrderOption) TQuery
	All(context.Context) ([]*TDb, error)
}](
	ctx context.Context,
	predicate *orm.Predicate,
	order []orm.OrderOption,
	opts crud.PagingOptions,
	query TQuery,
	convertFn func([]*TDb) []TDomain,
) (*crud.PagedResult[TDomain], error) {
	wholeQuery := query
	if predicate != nil {
		wholeQuery = wholeQuery.Where(*predicate)
	}

	if len(order) > 0 {
		wholeQuery = wholeQuery.Order(order...)
	}

	pagedQuery := wholeQuery.Clone().
		Offset(opts.Page * opts.Size).
		Limit(opts.Size)

	total, err := wholeQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	dbEntities, err := pagedQuery.All(ctx)
	if err != nil {
		return nil, err
	}

	return &crud.PagedResult[TDomain]{
		Items: convertFn(dbEntities),
		Total: total,
		Page:  opts.Page,
		Size:  opts.Size,
	}, nil
}
