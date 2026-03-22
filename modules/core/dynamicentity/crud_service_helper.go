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

type ToDomainModelFunc[TDomain any] func(data schema.DynamicFields) TDomain
type BeforeValidationFunc[TDomain any] func(ctx Context, model TDomain) (TDomain, error)
type AfterValidationFunc[TDomain any] func(ctx Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain any] func(ctx Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action       string
	DbRepoGetter DbRepoGetter

	// Data to create
	Data schema.DynamicModelGetter

	// Function to convert a dynamic entity to a domain model
	// ToDomainModel ToDomainModelFunc[TDomain]

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFunc[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidation AfterValidationFunc[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to dynamic entity schema validation.
	ValidateExtra ValidateExtraFunc[TDomainPtr]
}

func Create[
	TDomain any,
	TDomainPtr DynamicModelPtr[TDomain],
](
	ctx Context,
	param CreateParam[TDomain, TDomainPtr],
) (*OpResult[TDomain], error) {
	dbRepo := param.DbRepoGetter.GetDbRepo()
	entitySchema := dbRepo.GetSchema()

	fieldData := param.Data.GetFieldData()
	newModel := TDomainPtr(new(TDomain))
	newModel.SetFieldData(fieldData)
	// model := param.ToDomainModel(fieldData)

	flow := StartValidationFlow()
	clientErrs, err := flow.
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, newModel)
			if err == nil {
				fieldData = result.GetFieldData()
			}
			return err
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := entitySchema.Validate(fieldData)
			if clientErrs != nil {
				*vErrs = *clientErrs
			} else {
				fieldData = result
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, fieldData, dbRepo, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			newModel.SetFieldData(fieldData)
			return param.ValidateExtra(ctx, newModel, vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidation != nil {
				result, err := param.AfterValidation(ctx, newModel)
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

	inserted, err := dbRepo.Insert(ctx, fieldData)
	ft.PanicOnErr(err)

	insertedModel := TDomainPtr(new(TDomain))
	insertedModel.SetFieldData(inserted)
	return &OpResult[TDomain]{
		Data:    *insertedModel,
		IsEmpty: false,
	}, nil
}

func validateUniques(ctx Context, data schema.DynamicFields, dbRepo DbRepository, vErrs *ft.ClientErrors) error {
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
