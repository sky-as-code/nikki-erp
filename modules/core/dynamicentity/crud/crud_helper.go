package crud

import (
	"github.com/sky-as-code/nikki-erp/common/dynamicentity/schema"
	"github.com/sky-as-code/nikki-erp/common/fault"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	dEnt "github.com/sky-as-code/nikki-erp/modules/core/dynamicentity"
)

type ToDomainModelFunc[TDomain any] func(data schema.DynamicFields) TDomain
type BeforeValidationFunc[TDomain any] func(ctx dEnt.Context, model TDomain) (TDomain, error)
type AfterValidationFunc[TDomain any] func(ctx dEnt.Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain any] func(ctx dEnt.Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action       string
	DbRepoGetter dEnt.BaseRepoGetter

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
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
](
	ctx dEnt.Context,
	param CreateParam[TDomain, TDomainPtr],
) (*dEnt.OpResult[TDomain], error) {

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	entitySchema := baseRepo.GetSchema()
	fieldData := param.Data.GetFieldData()
	newModel := TDomainPtr(new(TDomain))
	newModel.SetFieldData(fieldData)

	flow := dEnt.StartValidationFlow()
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
			validateUniques(ctx, fieldData, baseRepo, vErrs)
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

	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dEnt.OpResult[TDomain]{
			ClientErrors: clientErrs,
		}, nil
	}

	inserted, err := baseRepo.Insert(ctx, fieldData)
	if err != nil {
		return nil, err
	}

	insertedModel := TDomainPtr(new(TDomain))
	insertedModel.SetFieldData(inserted)
	return &dEnt.OpResult[TDomain]{
		Data:    *insertedModel,
		IsEmpty: false,
	}, nil
}

type UpdateParam[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
] struct {
	Action           string
	DbRepoGetter     dEnt.BaseRepoGetter
	Data             schema.DynamicModelGetter
	BeforeValidation BeforeValidationFunc[TDomainPtr]
	AfterValidation  AfterValidationFunc[TDomainPtr]
	ValidateExtra    ValidateExtraFunc[TDomainPtr]
}

func Update[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
](
	ctx dEnt.Context,
	param UpdateParam[TDomain, TDomainPtr],
) (*dEnt.OpResult[TDomain], error) {
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(param.Data.GetFieldData())

	clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dEnt.OpResult[TDomain]{ClientErrors: clientErrs}, nil
	}

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	updated, err := baseRepo.Update(ctx, model.GetFieldData())
	if err != nil {
		return nil, err
	}

	model.SetFieldData(updated)
	return &dEnt.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func runUpdateValidationFlow[TDomain any, TDomainPtr dEnt.DynamicModelPtr[TDomain]](
	ctx dEnt.Context,
	param UpdateParam[TDomain, TDomainPtr],
	model TDomainPtr,
) (ft.ClientErrors, error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	entitySchema := baseRepo.GetSchema()

	return dEnt.StartValidationFlow().
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, model)
			if err == nil {
				model.SetFieldData(result.GetFieldData())
			}
			return err
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := entitySchema.Validate(model.GetFieldData(), true)
			if clientErrs != nil {
				*vErrs = *clientErrs
			} else {
				model.SetFieldData(result)
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			return checkExistenceAndEtag(ctx, entitySchema, baseRepo, model.GetFieldData(), vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, model.GetFieldData(), baseRepo, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			return param.ValidateExtra(ctx, model, vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidation == nil {
				return nil
			}
			result, err := param.AfterValidation(ctx, model)
			if err == nil {
				model.SetFieldData(result.GetFieldData())
			}
			return err
		}).
		End()
}

func checkExistenceAndEtag(
	ctx dEnt.Context,
	entitySchema *schema.EntitySchema,
	baseRepo dEnt.BaseRepository,
	fieldData schema.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	pkData := make(schema.DynamicFields)
	for _, key := range entitySchema.KeyColumns() {
		pkData[key] = fieldData[key]
	}

	dbRecord, err := baseRepo.FindByPk(ctx, pkData)
	if err != nil {
		return err
	}

	if dbRecord == nil {
		vErrs.Append(*fault.NewNotFoundError(entitySchema.Name()))
		return nil
	}

	if _, hasEtag := entitySchema.Field("etag"); hasEtag && fieldData["etag"] != dbRecord["etag"] {
		vErrs.Append(*fault.NewEtagMismatchedError(entitySchema.Name()))
	}
	return nil
}

func GetByPk[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
](
	ctx dEnt.Context,
	dbRepoGetter dEnt.BaseRepoGetter,
	keys schema.DynamicFields,
) (*dEnt.OpResult[TDomain], error) {
	baseRepo := dbRepoGetter.GetBaseRepo()
	dbRecord, err := baseRepo.FindByPk(ctx, keys)
	if err != nil {
		return nil, err
	}

	if dbRecord == nil {
		clientErrs := fault.NewClientErrors()
		clientErrs.Append(*fault.NewNotFoundError(baseRepo.GetSchema().Name()))
		return &dEnt.OpResult[TDomain]{ClientErrors: *clientErrs}, nil
	}

	model := TDomainPtr(new(TDomain))
	model.SetFieldData(dbRecord)
	return &dEnt.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func Archive[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
](
	ctx dEnt.Context,
	dbRepoGetter dEnt.BaseRepoGetter,
	keys schema.DynamicFields,
) (*dEnt.OpResult[TDomain], error) {
	baseRepo := dbRepoGetter.GetBaseRepo()
	record, err := baseRepo.Archive(ctx, keys)
	if err != nil {
		return nil, err
	}

	if record == nil {
		return &dEnt.OpResult[TDomain]{IsEmpty: true}, nil
	}
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(record)
	return &dEnt.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func Search[
	TDomain any,
	TDomainPtr dEnt.DynamicModelPtr[TDomain],
](
	ctx dEnt.Context,
	dbRepoGetter dEnt.BaseRepoGetter,
	graph schema.SearchGraph,
	columns []string,
) (*dEnt.OpResult[[]TDomain], error) {
	baseRepo := dbRepoGetter.GetBaseRepo()
	records, err := baseRepo.Search(ctx, graph, columns)
	if err != nil {
		return nil, err
	}

	items := make([]TDomain, len(records))
	for i, record := range records {
		model := TDomainPtr(new(TDomain))
		model.SetFieldData(record)
		items[i] = *model
	}
	return &dEnt.OpResult[[]TDomain]{Data: items, IsEmpty: len(items) == 0}, nil
}

func validateUniques(ctx dEnt.Context, data schema.DynamicFields, dbRepo dEnt.BaseRepository, vErrs *ft.ClientErrors) error {
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
