package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attributevalue/interfaces"
)

func NewAttributeValueServiceImpl(
	attributeValueRepo it.AttributeValueRepository,
) it.AttributeValueService {
	return &AttributeValueServiceImpl{
		attributeValueRepo: attributeValueRepo,
	}
}

type AttributeValueServiceImpl struct {
	attributeValueRepo it.AttributeValueRepository
}

// Create

func (s *AttributeValueServiceImpl) CreateAttributeValue(ctx crud.Context, cmd it.CreateAttributeValueCommand) (*it.CreateAttributeValueResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.AttributeValue, it.CreateAttributeValueCommand, it.CreateAttributeValueResult]{
		Action:     "create attribute value",
		Command:    cmd,
		RepoCreate: s.attributeValueRepo.Create,
		Sanitize:   s.sanitizeAttributeValue,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateAttributeValueResult {
			return &it.CreateAttributeValueResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeValue) *it.CreateAttributeValueResult {
			return &it.CreateAttributeValueResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *AttributeValueServiceImpl) UpdateAttributeValue(ctx crud.Context, cmd it.UpdateAttributeValueCommand) (*it.UpdateAttributeValueResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.AttributeValue, it.UpdateAttributeValueCommand, it.UpdateAttributeValueResult]{
		Action:       "update attribute value",
		Command:      cmd,
		AssertExists: s.assertAttributeValueIdExists,
		RepoUpdate:   s.attributeValueRepo.Update,
		Sanitize:     s.sanitizeAttributeValue,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateAttributeValueResult {
			return &it.UpdateAttributeValueResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeValue) *it.UpdateAttributeValueResult {
			return &it.UpdateAttributeValueResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeValueServiceImpl) DeleteAttributeValue(ctx crud.Context, cmd it.DeleteAttributeValueCommand) (*it.DeleteAttributeValueResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.AttributeValue, it.DeleteAttributeValueCommand, it.DeleteAttributeValueResult]{
		Action:       "delete attribute value",
		Command:      cmd,
		AssertExists: s.assertAttributeValueIdExists,
		RepoDelete: func(ctx crud.Context, model *it.AttributeValue) (int, error) {
			return s.attributeValueRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteAttributeValueResult {
			return &it.DeleteAttributeValueResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *it.AttributeValue, deletedCount int) *it.DeleteAttributeValueResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeValueServiceImpl) GetAttributeValueById(ctx crud.Context, query it.GetAttributeValueByIdQuery) (*it.GetAttributeValueByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.AttributeValue, it.GetAttributeValueByIdQuery, it.GetAttributeValueByIdResult]{
		Action: "get attribute value by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetAttributeValueByIdQuery, vErrs *ft.ValidationErrors) (*it.AttributeValue, error) {
			dbAttributeValue, err := s.attributeValueRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttributeValue == nil {
				vErrs.AppendNotFound("id", "attribute value id")
			}
			return dbAttributeValue, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetAttributeValueByIdResult {
			return &it.GetAttributeValueByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.AttributeValue) *it.GetAttributeValueByIdResult {
			return &it.GetAttributeValueByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *AttributeValueServiceImpl) SearchAttributeValues(ctx crud.Context, query it.SearchAttributeValuesQuery) (*it.SearchAttributeValuesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.AttributeValue, it.SearchAttributeValuesQuery, it.SearchAttributeValuesResult]{
		Action: "search attribute values",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchAttributeValuesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.attributeValueRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query it.SearchAttributeValuesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.AttributeValue], error) {
			return this.attributeValueRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchAttributeValuesResult {
			return &it.SearchAttributeValuesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.AttributeValue]) *it.SearchAttributeValuesResult {
			return &it.SearchAttributeValuesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *AttributeValueServiceImpl) sanitizeAttributeValue(_ *it.AttributeValue) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeValueServiceImpl) assertAttributeValueIdExists(ctx crud.Context, attributeValue *it.AttributeValue, vErrs *ft.ValidationErrors) (*it.AttributeValue, error) {
	dbAttributeValue, err := s.attributeValueRepo.FindById(ctx, it.FindByIdParam{
		Id: *attributeValue.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbAttributeValue == nil {
		vErrs.Append("id", "attribute value not found")
		return nil, nil
	}

	return dbAttributeValue, nil
}
