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
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/product"
)

func NewAttributeServiceImpl(
	attributeRepo itAttribute.AttributeRepository,
	productSvc itProduct.ProductService,
) itAttribute.AttributeService {
	return &AttributeServiceImpl{
		attributeRepo: attributeRepo,
		productSvc:    productSvc,
	}
}

type AttributeServiceImpl struct {
	attributeRepo itAttribute.AttributeRepository
	productSvc    itProduct.ProductService
}

// Create

func (this *AttributeServiceImpl) CreateAttribute(ctx crud.Context, cmd itAttribute.CreateAttributeCommand) (result *itAttribute.CreateAttributeResult, err error) {

	attribute := cmd.ToDomainModel()
	this.setAttributeDefaults(ctx, attribute)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = cmd.Validate()
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertCreateAttribute(ctx, attribute, vErrs)
			return nil
		}).
		End()

	if vErrs.Count() > 0 {
		return &itAttribute.CreateAttributeResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbAttribute, err := this.attributeRepo.Create(ctx, attribute)
	if err != nil {
		return nil, err
	}

	return &itAttribute.CreateAttributeResult{
		HasData: true,
		Data:    dbAttribute,
	}, nil
}

// Update

func (s *AttributeServiceImpl) UpdateAttribute(ctx crud.Context, cmd itAttribute.UpdateAttributeCommand) (*itAttribute.UpdateAttributeResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*domain.Attribute, itAttribute.UpdateAttributeCommand, itAttribute.UpdateAttributeResult]{
		Action:       "update attribute",
		Command:      cmd,
		AssertExists: s.assertAttributeIdExists,
		RepoUpdate:   s.attributeRepo.Update,
		Sanitize:     s.sanitizeAttribute,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.UpdateAttributeResult {
			return &itAttribute.UpdateAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Attribute) *itAttribute.UpdateAttributeResult {
			return &itAttribute.UpdateAttributeResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeServiceImpl) DeleteAttribute(ctx crud.Context, cmd itAttribute.DeleteAttributeCommand) (*itAttribute.DeleteAttributeResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*domain.Attribute, itAttribute.DeleteAttributeCommand, itAttribute.DeleteAttributeResult]{
		Action:       "delete attribute",
		Command:      cmd,
		AssertExists: s.assertAttributeIdExists,
		RepoDelete: func(ctx crud.Context, model *domain.Attribute) (int, error) {
			return s.attributeRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.DeleteAttributeResult {
			return &itAttribute.DeleteAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *domain.Attribute, deletedCount int) *itAttribute.DeleteAttributeResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeServiceImpl) GetAttributeById(ctx crud.Context, query itAttribute.GetAttributeByIdQuery) (*itAttribute.GetAttributeByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Attribute, itAttribute.GetAttributeByIdQuery, itAttribute.GetAttributeByIdResult]{
		Action: "get attribute by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itAttribute.GetAttributeByIdQuery, vErrs *ft.ValidationErrors) (*domain.Attribute, error) {
			dbAttribute, err := s.attributeRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttribute == nil {
				vErrs.AppendNotFound("id", "attribute id")
			}
			return dbAttribute, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.GetAttributeByIdResult {
			return &itAttribute.GetAttributeByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Attribute) *itAttribute.GetAttributeByIdResult {
			return &itAttribute.GetAttributeByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

func (s *AttributeServiceImpl) GetAttributeByCodeName(ctx crud.Context, query itAttribute.GetAttributeByCodeName) (*itAttribute.GetAttributeByCodeNameResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*domain.Attribute, itAttribute.GetAttributeByCodeName, itAttribute.GetAttributeByCodeNameResult]{
		Action: "get attribute by code name",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q itAttribute.GetAttributeByCodeName, vErrs *ft.ValidationErrors) (*domain.Attribute, error) {
			dbAttribute, err := s.attributeRepo.FindByCodeName(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttribute == nil {
				vErrs.AppendNotFound("code_name", "attribute code name")
			}
			return dbAttribute, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.GetAttributeByCodeNameResult {
			return &itAttribute.GetAttributeByCodeNameResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Attribute) *itAttribute.GetAttributeByCodeNameResult {
			return &itAttribute.GetAttributeByCodeNameResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *AttributeServiceImpl) SearchAttributes(ctx crud.Context, query itAttribute.SearchAttributesQuery) (*itAttribute.SearchAttributesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[domain.Attribute, itAttribute.SearchAttributesQuery, itAttribute.SearchAttributesResult]{
		Action: "search attributes",
		Query:  query,
		SetQueryDefaults: func(q *itAttribute.SearchAttributesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.attributeRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query itAttribute.SearchAttributesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[domain.Attribute], error) {
			return this.attributeRepo.Search(ctx, itAttribute.SearchParam{
				ProductId:     query.ProductId,
				Predicate:     predicate,
				Order:         order,
				Page:          *query.Page,
				Size:          *query.Size,
				CountValues:   query.CountValues,
				CountVariants: query.CountVariants,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.SearchAttributesResult {
			return &itAttribute.SearchAttributesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[domain.Attribute]) *itAttribute.SearchAttributesResult {
			return &itAttribute.SearchAttributesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// assert methods
// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *AttributeServiceImpl) assertCreateAttribute(ctx crud.Context, attribute *domain.Attribute, vErrs *ft.ValidationErrors) error {
	product, err := this.productSvc.GetProductById(ctx, itProduct.GetProductByIdQuery{
		Id: *attribute.ProductId,
	})
	ft.PanicOnErr(err)

	if product.Data == nil {
		vErrs.Append("id", "product does not exist")
		return nil
	}

	attributeWithSameCodeName, err := this.attributeRepo.FindByCodeName(ctx, itAttribute.GetAttributeByCodeName{
		ProductId: *attribute.ProductId,
		CodeName:  *attribute.CodeName,
	})
	ft.PanicOnErr(err)

	if attributeWithSameCodeName != nil {
		vErrs.Append("codeName", "attribute code name already exists for this product")
	}

	if attribute.IsEnum == nil || *attribute.IsEnum == false {
		if attribute.EnumValue != nil && len(*attribute.EnumValue) > 0 {
			vErrs.Append("enumValue", "enum value should be empty when isEnum is false")
		}
		return nil
	}

	if *attribute.DataType == "string" {
		for _, v := range *attribute.EnumValue {
			var enumValue model.LangJson
			err := json.Unmarshal(v, &enumValue)
			if err != nil {
				vErrs.Append("enumValue", "invalid enum value, should be a valid lang json")
				return nil
			}
		}
	} else if *attribute.DataType == "number" {
		for _, v := range *attribute.EnumValue {
			var enumValue float64
			err := json.Unmarshal(v, &enumValue)
			if err != nil {
				vErrs.Append("enumValue", "invalid enum value, should be a number")
				return nil
			}
		}
	} else {
		vErrs.Append("dataType", "invalid data type, only string and number are allowed for enum attribute")
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (s *AttributeServiceImpl) sanitizeAttribute(_ *domain.Attribute) {
}

func (s *AttributeServiceImpl) setAttributeDefaults(ctx crud.Context, attribute *domain.Attribute) {
	attribute.SetDefaults()

	nextSortIndex, err := s.attributeRepo.GetNextSortIndex(ctx, *attribute.ProductId)
	if err != nil {
		return
	}
	attribute.SortIndex = &nextSortIndex
}

func (s *AttributeServiceImpl) assertAttributeIdExists(ctx crud.Context, attribute *domain.Attribute, vErrs *ft.ValidationErrors) (*domain.Attribute, error) {
	dbAttribute, err := s.attributeRepo.FindById(ctx, itAttribute.FindByIdParam{
		ProductId: *attribute.ProductId,
		Id:        *attribute.Id,
	})
	if err != nil {
		return nil, err
	}

	if dbAttribute == nil {
		vErrs.Append("id", "attribute not found")
		return nil, nil
	}

	return dbAttribute, nil
}
