package app

import (
	"encoding/json"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
	itAttributeValue "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attributevalue"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantServiceImpl(
	variantRepo itVariant.VariantRepository,
	attribute itAttribute.AttributeService,
	attributeValue itAttributeValue.AttributeValueService,
) itVariant.VariantService {
	return &VariantServiceImpl{
		variantRepo:    variantRepo,
		attribute:      attribute,
		attributeValue: attributeValue,
	}
}

type VariantServiceImpl struct {
	variantRepo    itVariant.VariantRepository
	attribute      itAttribute.AttributeService
	attributeValue itAttributeValue.AttributeValueService
}

// Create

func (s *VariantServiceImpl) CreateVariant(ctx crud.Context, cmd itVariant.CreateVariantCommand) (result *itVariant.CreateVariantResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to add or remove users"); e != nil {
			err = e
		}
	}()

	variant := cmd.ToDomainModel()
	variant.SetDefaults()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = variant.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return s.assertCreateVariant(ctx, variant, vErrs)
		}).
		End()

	if vErrs.Count() > 0 {
		return &itVariant.CreateVariantResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbVariant, err := s.variantRepo.Create(ctx, variant)
	ft.PanicOnErr(err)

	return &itVariant.CreateVariantResult{
		HasData: true,
		Data:    dbVariant,
	}, nil
}

// Update

func (s *VariantServiceImpl) UpdateVariant(ctx crud.Context, cmd itVariant.UpdateVariantCommand) (*itVariant.UpdateVariantResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Variant, itVariant.UpdateVariantCommand, itVariant.UpdateVariantResult]{
		Action:       "update variant",
		Command:      cmd,
		AssertExists: s.assertVariantIdExists,
		RepoUpdate:   s.variantRepo.Update,
		Sanitize:     s.sanitizeVariant,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itVariant.UpdateVariantResult {
			return &itVariant.UpdateVariantResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Variant) *itVariant.UpdateVariantResult {
			return &itVariant.UpdateVariantResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *VariantServiceImpl) DeleteVariant(ctx crud.Context, cmd itVariant.DeleteVariantCommand) (*itVariant.DeleteVariantResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Variant, itVariant.DeleteVariantCommand, itVariant.DeleteVariantResult]{
		Action:       "delete variant",
		Command:      cmd,
		AssertExists: s.assertVariantIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.Variant) (int, error) {
			return s.variantRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itVariant.DeleteVariantResult {
			return &itVariant.DeleteVariantResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.Variant, deletedCount int) *itVariant.DeleteVariantResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *VariantServiceImpl) GetVariantById(ctx crud.Context, query itVariant.GetVariantByIdQuery) (*itVariant.GetVariantByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Variant, itVariant.GetVariantByIdQuery, itVariant.GetVariantByIdResult]{
		Action: "get variant by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itVariant.GetVariantByIdQuery, vErrs *ft.ValidationErrors) (*domain.Variant, error) {
			dbVariant, err := s.variantRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbVariant == nil {
				vErrs.AppendNotFound("id", "variant id")
			}
			return dbVariant, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itVariant.GetVariantByIdResult {
			return &itVariant.GetVariantByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Variant) *itVariant.GetVariantByIdResult {
			return &itVariant.GetVariantByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *VariantServiceImpl) SearchVariants(ctx crud.Context, query itVariant.SearchVariantsQuery) (*itVariant.SearchVariantsResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Variant, itVariant.SearchVariantsQuery, itVariant.SearchVariantsResult]{
		Action: "search variants",
		Query:  query,
		SetQueryDefaults: func(q *itVariant.SearchVariantsQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.variantRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itVariant.SearchVariantsQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Variant], error) {
			return this.variantRepo.Search(ctx, itVariant.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itVariant.SearchVariantsResult {
			return &itVariant.SearchVariantsResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.Variant]) *itVariant.SearchVariantsResult {
			return &itVariant.SearchVariantsResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// Helpers
//---------------------------------------------------------------------------------------------------------------------------------------------//

func (s *VariantServiceImpl) sanitizeVariant(_ *domain.Variant) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *VariantServiceImpl) assertCreateVariant(ctx crud.Context, variant *domain.Variant, vErrs *ft.ValidationErrors) error {
	for codename, value := range *variant.Attributes {
		attribute, err := s.attribute.GetAttributeByCodeName(ctx, itAttribute.GetAttributeByCodeName{
			ProductId: *variant.ProductId,
			CodeName:  codename,
		})
		ft.PanicOnErr(err)

		if attribute.Data == nil {
			vErrs.Append("attributes", "codename "+codename+" not found")
			return nil
		}

		if *attribute.Data.DataType == "number" {
			var temp float64
			if err := json.Unmarshal(value, &temp); err != nil {
				vErrs.Append("attributes."+codename, "invalid value type")
				return nil
			}

			_, err := s.attributeValue.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
				AttributeId: *attribute.Data.Id,
				VariantId:   *variant.Id,
				ValueNumber: &temp,
			})
			ft.PanicOnErr(err)
		} else {
			var temp model.LangJson
			if err := json.Unmarshal(value, &temp); err != nil {
				vErrs.Append("attributes."+codename, "invalid value type")
				return nil
			}

			_, err := s.attributeValue.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
				AttributeId: *attribute.Data.Id,
				VariantId:   *variant.Id,
				ValueText:   &temp,
			})
			ft.PanicOnErr(err)
		}
	}
	return nil
}

func (s *VariantServiceImpl) assertVariantIdExists(ctx crud.Context, variant *domain.Variant, vErrs *ft.ValidationErrors) (*domain.Variant, error) {
	dbVariant, err := s.variantRepo.FindById(ctx, itVariant.FindByIdParam{
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

// func
