package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/variant/interfaces"
)

func NewVariantServiceImpl(
	variantRepo it.VariantRepository,
) it.VariantService {
	return &VariantServiceImpl{
		variantRepo: variantRepo,
	}
}

type VariantServiceImpl struct {
	variantRepo it.VariantRepository
}

// Create

func (s *VariantServiceImpl) CreateVariant(ctx crud.Context, cmd it.CreateVariantCommand) (*it.CreateVariantResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.Variant, it.CreateVariantCommand, it.CreateVariantResult]{
		Action:     "create variant",
		Command:    cmd,
		RepoCreate: s.variantRepo.Create,
		Sanitize:   s.sanitizeVariant,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateVariantResult {
			return &it.CreateVariantResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Variant) *it.CreateVariantResult {
			return &it.CreateVariantResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *VariantServiceImpl) UpdateVariant(ctx crud.Context, cmd it.UpdateVariantCommand) (*it.UpdateVariantResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.Variant, it.UpdateVariantCommand, it.UpdateVariantResult]{
		Action:       "update variant",
		Command:      cmd,
		AssertExists: s.assertVariantIdExists,
		RepoUpdate:   s.variantRepo.Update,
		Sanitize:     s.sanitizeVariant,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateVariantResult {
			return &it.UpdateVariantResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Variant) *it.UpdateVariantResult {
			return &it.UpdateVariantResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *VariantServiceImpl) DeleteVariant(ctx crud.Context, cmd it.DeleteVariantCommand) (*it.DeleteVariantResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.Variant, it.DeleteVariantCommand, it.DeleteVariantResult]{
		Action:       "delete variant",
		Command:      cmd,
		AssertExists: s.assertVariantIdExists,
		RepoDelete: func(ctx crud.Context, model *it.Variant) (int, error) {
			return s.variantRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteVariantResult {
			return &it.DeleteVariantResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *it.Variant, deletedCount int) *it.DeleteVariantResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *VariantServiceImpl) GetVariantById(ctx crud.Context, query it.GetVariantByIdQuery) (*it.GetVariantByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.Variant, it.GetVariantByIdQuery, it.GetVariantByIdResult]{
		Action: "get variant by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetVariantByIdQuery, vErrs *ft.ValidationErrors) (*it.Variant, error) {
			dbVariant, err := s.variantRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbVariant == nil {
				vErrs.AppendNotFound("id", "variant id")
			}
			return dbVariant, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetVariantByIdResult {
			return &it.GetVariantByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Variant) *it.GetVariantByIdResult {
			return &it.GetVariantByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *VariantServiceImpl) SearchVariants(ctx crud.Context, query it.SearchVariantsQuery) (*it.SearchVariantsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.Variant, it.SearchVariantsQuery, it.SearchVariantsResult]{
		Action: "search variants",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchVariantsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.variantRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query it.SearchVariantsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.Variant], error) {
			return this.variantRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchVariantsResult {
			return &it.SearchVariantsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.Variant]) *it.SearchVariantsResult {
			return &it.SearchVariantsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *VariantServiceImpl) sanitizeVariant(_ *it.Variant) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *VariantServiceImpl) assertVariantIdExists(ctx crud.Context, variant *it.Variant, vErrs *ft.ValidationErrors) (*it.Variant, error) {
	dbVariant, err := s.variantRepo.FindById(ctx, it.FindByIdParam{
		Id: *variant.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbVariant == nil {
		vErrs.Append("id", "variant not found")
		return nil, nil
	}

	return dbVariant, nil
}
