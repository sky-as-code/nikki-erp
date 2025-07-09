package role_suite

type RoleSuiteRepository interface {
	// Create(ctx context.Context, resource domain.Resource) (*domain.Resource, error)
	// FindByName(ctx context.Context, param FindByNameParam) (*domain.Resource, error)
	// FindById(ctx context.Context, param FindByIdParam) (*domain.Resource, error)
	// Update(ctx context.Context, resource domain.Resource) (*domain.Resource, error)
	// ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	// Search(ctx context.Context, param SearchParam) (*crud.PagedResult[domain.Resource], error)
}

// type FindByIdParam = GetResourceByIdQuery
// type FindByNameParam = GetResourceByNameCommand

// type SearchParam struct {
// 	Predicate   *orm.Predicate
// 	Order       []orm.OrderOption
// 	Page        int
// 	Size        int
// 	WithActions bool
// }
