package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
	itAttribute "github.com/sky-as-code/nikki-erp/modules/inventory/product/interfaces/attribute"
)

func NewAttributeServiceImpl(
	attributeRepo itAttribute.AttributeRepository,
) itAttribute.AttributeService {
	return &AttributeServiceImpl{
		attributeRepo: attributeRepo,
	}
}

type AttributeServiceImpl struct {
	attributeRepo itAttribute.AttributeRepository
}

// Create

func (this *AttributeServiceImpl) CreateAttribute(ctx crud.Context, cmd itAttribute.CreateAttributeCommand) (*itAttribute.CreateAttributeResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*domain.Attribute, itAttribute.CreateAttributeCommand, itAttribute.CreateAttributeResult]{
		Action:              "create attribute",
		Command:             cmd,
		RepoCreate:          this.attributeRepo.Create,
		AssertBusinessRules: this.assertCreateAttribute,
		Sanitize:            this.sanitizeAttribute,
		SetDefault:          this.setAttributeDefaults,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *itAttribute.CreateAttributeResult {
			return &itAttribute.CreateAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *domain.Attribute) *itAttribute.CreateAttributeResult {
			return &itAttribute.CreateAttributeResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
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
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
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
	dbAttribute, err := this.attributeRepo.FindById(ctx, itAttribute.FindByIdParam{
		Id: *attribute.Id,
	})
	if err != nil {
		return err
	}

	if dbAttribute != nil {
		vErrs.Append("id", "attribute already exists")
		return nil
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (s *AttributeServiceImpl) sanitizeAttribute(_ *domain.Attribute) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeServiceImpl) setAttributeDefaults(attribute *domain.Attribute) {
	attribute.SetDefaults()
}

func (s *AttributeServiceImpl) assertAttributeIdExists(ctx crud.Context, attribute *domain.Attribute, vErrs *ft.ValidationErrors) (*domain.Attribute, error) {
	dbAttribute, err := s.attributeRepo.FindById(ctx, itAttribute.FindByIdParam{
		Id: *attribute.Id,
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
