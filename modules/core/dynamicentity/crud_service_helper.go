package dynamicentity

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

type CrudServiceHelper struct {
	dyEntService *DynamicEntityService
	dbRepo       DbRepository
}

type DynamicDomainModel interface {
	GetFieldData() schema.DynamicEntity
	SetFieldData(data schema.DynamicEntity)
}

type ToDomainModelFunc[TDomain DynamicDomainModel] func(data schema.DynamicEntity) TDomain
type BeforeValidationFunc[TDomain DynamicDomainModel] func(ctx Context, model TDomain) (TDomain, error)
type AfterValidationFunc[TDomain DynamicDomainModel] func(ctx Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain DynamicDomainModel] func(ctx Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain DynamicDomainModel,
] struct {
	// Action name for logging and error messages
	Action string
	DbRepo DbRepository

	// Data to create
	Data schema.DynamicEntity

	// Function to convert a dynamic entity to a domain model
	ToDomainModel ToDomainModelFunc[TDomain]

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFunc[TDomain]

	// Optional function to do some processing on the domain model after validation.
	AfterValidation AfterValidationFunc[TDomain]

	// Optional function for advanced validation (business rules) in addition to dynamic entity schema validation.
	ValidateExtra ValidateExtraFunc[TDomain]
}

func Create[
	TDomain DynamicDomainModel,
](
	ctx Context,
	param CreateParam[TDomain],
) (*OpResult[TDomain], error) {
	entitySchema := param.DbRepo.GetSchema()

	fieldData := param.Data
	model := param.ToDomainModel(param.Data)
	// if param.BeforeValidation != nil {
	// 	model = param.BeforeValidation(ctx, model)
	// }

	flow := StartValidationFlow()
	clientErrs, err := flow.
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, model)
			if err == nil {
				fieldData = result.GetFieldData()
			}
			return err
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := entitySchema.Validate(fieldData)
			fieldData = result
			*vErrs = *clientErrs
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, fieldData, param.DbRepo, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			model = param.ToDomainModel(fieldData)
			return param.ValidateExtra(ctx, model, vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidation != nil {
				result, err := param.AfterValidation(ctx, model)
				if err == nil {
					fieldData = result.GetFieldData()
				}
				return err
			}
			return nil
		}).
		End()

	ft.PanicOnErr(err)

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &OpResult[TDomain]{
			ClientErrors: clientErrs,
		}, nil
	}

	inserted, err := param.DbRepo.Insert(ctx, fieldData)
	ft.PanicOnErr(err)

	return &OpResult[TDomain]{
		Data:    param.ToDomainModel(inserted),
		IsEmpty: false,
	}, nil
}

func validateUniques(ctx Context, data schema.DynamicEntity, dbRepo DbRepository, vErrs *ft.ClientErrors) error {
	collidingKeys, err := dbRepo.CheckUniqueCollisions(ctx, data)
	if err != nil {
		return err
	}

	if len(collidingKeys) > 0 {
		field := collidingKeys[0][0]
		vErrs.Append(*fault.NewBusinessViolation(
			field,
			"common.err_unique_constraint_violated",
			"unique constraint violated {{.uniques}}",
			map[string]any{"uniques": collidingKeys},
		))
	}
	return nil
}
