package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
)

func NewAttributeValueServiceImpl(
	attributeValueRepo itAttributeValue.AttributeValueRepository,
	attributeService itAttribute.AttributeService,
) itAttributeValue.AttributeValueService {
	return &AttributeValueServiceImpl{
		attributeValueRepo: attributeValueRepo,
		attributeService:   attributeService,
	}
}

type AttributeValueServiceImpl struct {
	attributeValueRepo itAttributeValue.AttributeValueRepository
	attributeService   itAttribute.AttributeService
}

// Create

func (s *AttributeValueServiceImpl) CreateAttributeValue(ctx crud.Context, cmd itAttributeValue.CreateAttributeValueCommand) (result *itAttributeValue.CreateAttributeValueResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to add or remove users"); e != nil {
			err = e
		}
	}()

	value := cmd.ToDomainModel()
	value.SetDefaults()

	var attributeResult *itAttribute.GetAttributeByIdResult

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			attributeResult, err = s.assertAttributeExists(ctx, value, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbAttributeValue, err := s.assertCreateAttributeValue(ctx, value, *attributeResult.Data.DataType, vErrs)
			if err != nil {
				return err
			}

			if dbAttributeValue == nil {
				dbAttributeValue, err = s.attributeValueRepo.CreateAndLinkVariant(ctx, value, cmd.VariantId)
				return err
			}

			_, _, err = s.attributeValueRepo.LinkVariantToExisting(ctx, *dbAttributeValue.Id, cmd.VariantId, *dbAttributeValue.Etag)
			return nil
		}).
		End()

	if vErrs.Count() > 0 {
		return &itAttributeValue.CreateAttributeValueResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbValue, err := s.attributeValueRepo.Create(ctx, value)
	ft.PanicOnErr(err)

	return &itAttributeValue.CreateAttributeValueResult{
		HasData: true,
		Data:    dbValue,
	}, nil
	// result, err := crud.Create(ctx, crud.CreateParam[*domain.AttributeValue, itAttributeValue.CreateAttributeValueCommand, itAttributeValue.CreateAttributeValueResult]{
	// 	Action:              "create attribute value",
	// 	Command:             cmd,
	// 	RepoCreate:          s.attributeValueRepo.Create,
	// 	AssertBusinessRules: s.assertCreateAttributeValue,
	// 	Sanitize:            s.sanitizeAttributeValue,
	// 	ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeValue.CreateAttributeValueResult {
	// 		return &itAttributeValue.CreateAttributeValueResult{
	// 			ClientError: vErrs.ToClientError(),
	// 		}
	// 	},
	// 	ToSuccessResult: func(model *domain.AttributeValue) *itAttributeValue.CreateAttributeValueResult {
	// 		return &itAttributeValue.CreateAttributeValueResult{
	// 			HasData: true,
	// 			Data:    model,
	// 		}
	// 	},
	// })
	// return result, err
}

// Update

func (s *AttributeValueServiceImpl) UpdateAttributeValue(ctx crud.Context, cmd itAttributeValue.UpdateAttributeValueCommand) (*itAttributeValue.UpdateAttributeValueResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.AttributeValue, itAttributeValue.UpdateAttributeValueCommand, itAttributeValue.UpdateAttributeValueResult]{
		Action:       "update attribute value",
		Command:      cmd,
		AssertExists: s.assertAttributeValueIdExists,
		RepoUpdate:   s.attributeValueRepo.Update,
		Sanitize:     s.sanitizeAttributeValue,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeValue.UpdateAttributeValueResult {
			return &itAttributeValue.UpdateAttributeValueResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.AttributeValue) *itAttributeValue.UpdateAttributeValueResult {
			return &itAttributeValue.UpdateAttributeValueResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeValueServiceImpl) DeleteAttributeValue(ctx crud.Context, cmd itAttributeValue.DeleteAttributeValueCommand) (*itAttributeValue.DeleteAttributeValueResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.AttributeValue, itAttributeValue.DeleteAttributeValueCommand, itAttributeValue.DeleteAttributeValueResult]{
		Action:       "delete attribute value",
		Command:      cmd,
		AssertExists: s.assertAttributeValueIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.AttributeValue) (int, error) {
			return s.attributeValueRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeValue.DeleteAttributeValueResult {
			return &itAttributeValue.DeleteAttributeValueResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.AttributeValue, deletedCount int) *itAttributeValue.DeleteAttributeValueResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeValueServiceImpl) GetAttributeValueById(ctx crud.Context, query itAttributeValue.GetAttributeValueByIdQuery) (*itAttributeValue.GetAttributeValueByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.AttributeValue, itAttributeValue.GetAttributeValueByIdQuery, itAttributeValue.GetAttributeValueByIdResult]{
		Action: "get attribute value by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itAttributeValue.GetAttributeValueByIdQuery, vErrs *ft.ValidationErrors) (*domain.AttributeValue, error) {
			dbAttributeValue, err := s.attributeValueRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttributeValue == nil {
				vErrs.AppendNotFound("id", "attribute value id")
			}
			return dbAttributeValue, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeValue.GetAttributeValueByIdResult {
			return &itAttributeValue.GetAttributeValueByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.AttributeValue) *itAttributeValue.GetAttributeValueByIdResult {
			return &itAttributeValue.GetAttributeValueByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *AttributeValueServiceImpl) SearchAttributeValues(ctx crud.Context, query itAttributeValue.SearchAttributeValuesQuery) (*itAttributeValue.SearchAttributeValuesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.AttributeValue, itAttributeValue.SearchAttributeValuesQuery, itAttributeValue.SearchAttributeValuesResult]{
		Action: "search attribute values",
		Query:  query,
		SetQueryDefaults: func(q *itAttributeValue.SearchAttributeValuesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.attributeValueRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itAttributeValue.SearchAttributeValuesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.AttributeValue], error) {
			return this.attributeValueRepo.Search(ctx, itAttributeValue.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttributeValue.SearchAttributeValuesResult {
			return &itAttributeValue.SearchAttributeValuesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.AttributeValue]) *itAttributeValue.SearchAttributeValuesResult {
			return &itAttributeValue.SearchAttributeValuesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *AttributeValueServiceImpl) sanitizeAttributeValue(_ *domain.AttributeValue) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeValueServiceImpl) assertAttributeValueIdExists(ctx crud.Context, attributeValue *domain.AttributeValue, vErrs *ft.ValidationErrors) (*domain.AttributeValue, error) {
	dbAttributeValue, err := s.attributeValueRepo.FindById(ctx, itAttributeValue.FindByIdParam{
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

func (s *AttributeValueServiceImpl) assertCreateAttributeValue(ctx crud.Context, attributeValue *domain.AttributeValue, dataType string, vErrs *ft.ValidationErrors) (*domain.AttributeValue, error) {
	value, err := s.attributeValueRepo.FindByValueRef(ctx, attributeValue, dataType)
	ft.PanicOnErr(err)

	if value != nil {
		vErrs.Append("value", "attribute value already exists")
		return nil, nil
	}
	return value, nil
}

func (s *AttributeValueServiceImpl) assertAttributeExists(ctx crud.Context, attributeValue *domain.AttributeValue, vErrs *ft.ValidationErrors) (result *itAttribute.GetAttributeByIdResult, err error) {
	attribute, err := s.attributeService.GetAttributeById(ctx, itAttribute.FindByIdParam{
		Id: *attributeValue.AttributeId,
	})
	ft.PanicOnErr(err)
	if attribute.Data == nil {
		vErrs.Append("attributeId", "attribute not found")
		return nil, nil
	}
	return attribute, nil
}
