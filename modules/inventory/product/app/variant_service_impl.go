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
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
	itVariant "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/variant"
)

func NewVariantServiceImpl(
	variantRepo itVariant.VariantRepository,
	attributeValue itAttributeValue.AttributeValueService,
	productSvc itProduct.ProductService,
) itVariant.VariantService {
	return &VariantServiceImpl{
		variantRepo:    variantRepo,
		attribute:      nil,
		attributeValue: attributeValue,
		productSvc:     productSvc,
	}
}

type VariantServiceImpl struct {
	variantRepo    itVariant.VariantRepository
	attribute      itAttribute.AttributeService
	attributeValue itAttributeValue.AttributeValueService
	productSvc     itProduct.ProductService
}

func (this *VariantServiceImpl) SetAttributeService(attribute itAttribute.AttributeService) {
	this.attribute = attribute
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

	var dbVariant *domain.Variant
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = variant.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			dbVariant, err = s.variantRepo.Create(ctx, variant)
			return err
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

func (s *VariantServiceImpl) GetVariantById(ctx crud.Context, query itVariant.GetVariantByIdQuery) (result *itVariant.GetVariantByIdResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get variant by id"); e != nil {
			err = e
		}
	}()

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return s.assertProductIdExists(ctx, query.ProductId, vErrs)
		}).
		End()

	if vErrs.Count() > 0 {
		return &itVariant.GetVariantByIdResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbVariant, err := s.variantRepo.FindById(ctx, itVariant.FindByIdParam{
		Id:        query.Id,
		ProductId: query.ProductId,
	})
	ft.PanicOnErr(err)

	if dbVariant == nil {
		return &itVariant.GetVariantByIdResult{
			ClientError: &ft.ClientError{
				Code:    "not_found",
				Details: "variant not found",
			},
		}, nil
	}

	for _, attrVal := range dbVariant.AttributeValue {
		attribute, err := s.attribute.GetAttributeById(ctx, itAttribute.GetAttributeByIdQuery{
			Id:        *attrVal.AttributeId,
			ProductId: query.ProductId,
		})
		ft.PanicOnErr(err)

		if attribute == nil || attribute.Data == nil {
			continue
		}

		value := attrVal.GetValue()
		if value == nil {
			continue
		}

		(*dbVariant.Attributes)[*attribute.Data.CodeName] = value
	}

	return &itVariant.GetVariantByIdResult{
		HasData: true,
		Data:    dbVariant,
	}, nil
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
				ProductId: query.ProductId,
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
			valueNumber, ok := value.(float64)
			if !ok {
				vErrs.Append("attributes."+codename, "invalid value type")
				return nil
			}

			_, err := s.attributeValue.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
				ProductId:   *variant.ProductId,
				AttributeId: *attribute.Data.Id,
				VariantId:   *variant.Id,
				ValueNumber: &valueNumber,
			})
			ft.PanicOnErr(err)
		} else {
			bytes, err := json.Marshal(value)
			if err != nil {
				vErrs.Append("attributes."+codename, "invalid json value")
				return nil
			}

			var langJson model.LangJson
			err = json.Unmarshal(bytes, &langJson)
			if err != nil {
				vErrs.Append("attributes."+codename, "invalid language structure")
				return nil
			}

			_, err = s.attributeValue.CreateAttributeValue(ctx, itAttributeValue.CreateAttributeValueCommand{
				ProductId:   *variant.ProductId,
				AttributeId: *attribute.Data.Id,
				VariantId:   *variant.Id,
				ValueText:   &langJson,
			})
			ft.PanicOnErr(err)
		}
	}
	return nil
}

func (s *VariantServiceImpl) assertVariantIdExists(ctx crud.Context, variant *domain.Variant, vErrs *ft.ValidationErrors) (*domain.Variant, error) {
	dbVariant, err := s.variantRepo.FindById(ctx, itVariant.FindByIdParam{
		Id:        *variant.Id,
		ProductId: *variant.ProductId,
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

func (s *VariantServiceImpl) assertProductIdExists(ctx crud.Context, productId model.Id, vErrs *ft.ValidationErrors) error {
	product, err := s.productSvc.GetProductById(ctx, itProduct.GetProductByIdQuery{
		Id: productId,
	})
	ft.PanicOnErr(err)

	if product.Data == nil {
		vErrs.Append("productId", "product not found")
		return nil
	}

	return nil
}
