package database

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/json"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/predicate"
	entUser "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/user"
)

type MutationBuilder[TDb any] interface {
	Save(context.Context) (*TDb, error)
}

type QueryOneBuilder[TDb any] interface {
	Only(context.Context) (*TDb, error)
}

type SearchBuilder[TDb any, TQuery any] interface {
	All(context.Context) ([]*TDb, error)
	Count(context.Context) (int, error)
	Offset(int) *TQuery
	Limit(int) *TQuery
	Order(...orm.OrderOption) *TQuery
	Only(context.Context) (*TDb, error)
	Where(...predicate.User) *TQuery
}

type EntToDomainFn[TDb any, TDomain any] func(*TDb) *TDomain
type EntToDomainArrFn[TDb any, TDomain any] func([]*TDb) []*TDomain

type EntRepositoryBase struct {
	client *ent.Client
}

func (this *EntRepositoryBase) Client() *ent.Client {
	return this.client
}

func Mutate[TDb any, TDomain any](
	ctx context.Context,
	mutationBuilder MutationBuilder[TDb],
	convertFn EntToDomainFn[TDb, TDomain],
) (*TDomain, error) {
	entEntity, err := mutationBuilder.Save(ctx)
	if err != nil {
		return nil, err
	}

	domainEntity := convertFn(entEntity)
	return domainEntity, nil
}

func Delete[TDb any](
	ctx context.Context,
	deleteBuilder interface{ Exec(context.Context) error },
) error {
	return deleteBuilder.Exec(ctx)
}

func FindOne[TDb any, TDomain any](
	ctx context.Context,
	queryBuilder QueryOneBuilder[TDb],
	convertFn EntToDomainFn[TDb, TDomain],
) (*TDomain, error) {
	entEntity, err := queryBuilder.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return convertFn(entEntity), nil
}

// func Search[TDb any, TDomain any, TQuery any](
// 	ctx context.Context,
// 	criteria *orm.SearchGraph,
// 	opts *crud.PagingOptions,
// 	entityName string,
// 	searchBuilder SearchBuilder[TDb, TQuery],
// 	convertFn EntToDomainArrFn[TDb, TDomain],
// ) (*crud.PagedResult[*TDomain], error) {
// 	var errs error
// 	predicate, err := criteria.ToPredicate(entityName)
// 	errs = stdErr.Join(errs, err)

// 	order, err := orm.ToOrder[orm.OrderOption](entityName, criteria)
// 	errs = stdErr.Join(errs, err)

// 	if errs != nil {
// 		return nil, errs
// 	}

// 	wholeQuery := searchBuilder.Where(predicate)
// 	pagedQuery := searchBuilder.
// 		Offset(opts.Page * opts.Size).
// 		Limit(opts.Size).
// 		Order(order...)

// 	total, err := wholeQuery.Count(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	dbUsers, err := pagedQuery.All(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &crud.PagedResult[*TDomain]{
// 		Items: convertFn(dbUsers),
// 		Total: total,
// 	}, nil
// }

func ParseSearchGraphStr[TDb any, TDomain any](criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
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

	return ParseSearchGraph[TDb, TDomain](&graph)
}

func ParseSearchGraph[TDb any, TDomain any](criteria *orm.SearchGraph) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	if criteria == nil {
		return nil, nil, nil
	}

	predicate, vErrsPre := criteria.ToPredicate(entUser.Label)
	// order, vErrsOrd := orm.ToOrder(entUser.Label, *criteria)
	order, vErrsOrd := criteria.Order.ToOrderOptions(entUser.Label)
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
