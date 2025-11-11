package impl

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	it "github.com/sky-as-code/nikki-erp/modules/inventory/attribute/interfaces"
)

func NewAttributeServiceImpl(
	attributeRepo it.AttributeRepository,
) it.AttributeService {
	return &AttributeServiceImpl{
		attributeRepo: attributeRepo,
	}
}

type AttributeServiceImpl struct {
	attributeRepo it.AttributeRepository
}

// Create

func (this *AttributeServiceImpl) CreateAttribute(ctx crud.Context, cmd it.CreateAttributeCommand) (*it.CreateAttributeResult, error) {
	result, err := crud.Create(ctx, crud.CreateParam[*it.Attribute, it.CreateAttributeCommand, it.CreateAttributeResult]{
		Action:              "create attribute",
		Command:             cmd,
		RepoCreate:          this.attributeRepo.Create,
		AssertBusinessRules: this.assertCreateAttribute,
		Sanitize:            this.sanitizeAttribute,
		SetDefault:          this.setAttributeDefaults,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.CreateAttributeResult {
			return &it.CreateAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Attribute) *it.CreateAttributeResult {
			return &it.CreateAttributeResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Update

func (s *AttributeServiceImpl) UpdateAttribute(ctx crud.Context, cmd it.UpdateAttributeCommand) (*it.UpdateAttributeResult, error) {
	result, err := crud.Update(ctx, crud.UpdateParam[*it.Attribute, it.UpdateAttributeCommand, it.UpdateAttributeResult]{
		Action:       "update attribute",
		Command:      cmd,
		AssertExists: s.assertAttributeIdExists,
		RepoUpdate:   s.attributeRepo.Update,
		Sanitize:     s.sanitizeAttribute,
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.UpdateAttributeResult {
			return &it.UpdateAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Attribute) *it.UpdateAttributeResult {
			return &it.UpdateAttributeResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Delete

func (s *AttributeServiceImpl) DeleteAttribute(ctx crud.Context, cmd it.DeleteAttributeCommand) (*it.DeleteAttributeResult, error) {
	result, err := crud.DeleteHard(ctx, crud.DeleteHardParam[*it.Attribute, it.DeleteAttributeCommand, it.DeleteAttributeResult]{
		Action:       "delete attribute",
		Command:      cmd,
		AssertExists: s.assertAttributeIdExists,
		RepoDelete: func(ctx crud.Context, model *it.Attribute) (int, error) {
			return s.attributeRepo.DeleteById(ctx, *model.Id)
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.DeleteAttributeResult {
			return &it.DeleteAttributeResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(_ *it.Attribute, deletedCount int) *it.DeleteAttributeResult {
			return crud.NewSuccessDeletionResult(cmd.Id, &deletedCount)
		},
	})
	return result, err
}

// Get by ID

func (s *AttributeServiceImpl) GetAttributeById(ctx crud.Context, query it.GetAttributeByIdQuery) (*it.GetAttributeByIdResult, error) {
	result, err := crud.GetOne(ctx, crud.GetOneParam[*it.Attribute, it.GetAttributeByIdQuery, it.GetAttributeByIdResult]{
		Action: "get attribute by id",
		Query:  query,
		RepoFindOne: func(ctx crud.Context, q it.GetAttributeByIdQuery, vErrs *ft.ValidationErrors) (*it.Attribute, error) {
			dbAttribute, err := s.attributeRepo.FindById(ctx, q)
			if err != nil {
				return nil, err
			}
			if dbAttribute == nil {
				vErrs.AppendNotFound("id", "attribute id")
			}
			return dbAttribute, nil
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.GetAttributeByIdResult {
			return &it.GetAttributeByIdResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(model *it.Attribute) *it.GetAttributeByIdResult {
			return &it.GetAttributeByIdResult{
				HasData: true,
				Data:    model,
			}
		},
	})
	return result, err
}

// Search

func (this *AttributeServiceImpl) SearchAttributes(ctx crud.Context, query it.SearchAttributesQuery) (*it.SearchAttributesResult, error) {
	result, err := crud.Search(ctx, crud.SearchParam[it.Attribute, it.SearchAttributesQuery, it.SearchAttributesResult]{
		Action: "search attributes",
		Query:  query,
		SetQueryDefaults: func(q *it.SearchAttributesQuery) {
			q.SetDefaults()
		},
		ParseSearchGraph: func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
			return this.attributeRepo.ParseSearchGraph(criteria)
		},
		RepoSearch: func(ctx crud.Context, query it.SearchAttributesQuery, predicate *orm.Predicate, order []orm.OrderOption) (*crud.PagedResult[it.Attribute], error) {
			return this.attributeRepo.Search(ctx, it.SearchParam{
				Predicate: predicate,
				Order:     order,
				Page:      *query.Page,
				Size:      *query.Size,
			})
		},
		ToFailureResult: func(vErrs *ft.ValidationErrors) *it.SearchAttributesResult {
			return &it.SearchAttributesResult{
				ClientError: vErrs.ToClientError(),
			}
		},
		ToSuccessResult: func(paged *crud.PagedResult[it.Attribute]) *it.SearchAttributesResult {
			return &it.SearchAttributesResult{
				Data:    paged,
				HasData: paged.Items != nil,
			}
		},
	})
	return result, err
}

// assert methods
// ---------------------------------------------------------------------------------------------------------------------------------------------//
func (this *AttributeServiceImpl) assertCreateAttribute(ctx crud.Context, attribute *it.Attribute, vErrs *ft.ValidationErrors) error {
	dbAttribute, err := this.attributeRepo.FindById(ctx, it.FindByIdParam{
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
func (s *AttributeServiceImpl) sanitizeAttribute(_ *it.Attribute) {
	// Keep for future: trim/sanitize plain-text fields if any.
}

func (s *AttributeServiceImpl) setAttributeDefaults(attribute *it.Attribute) {
	attribute.SetDefaults()
}

func (s *AttributeServiceImpl) assertAttributeIdExists(ctx crud.Context, attribute *it.Attribute, vErrs *ft.ValidationErrors) (*it.Attribute, error) {
	dbAttribute, err := s.attributeRepo.FindById(ctx, it.FindByIdParam{
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
