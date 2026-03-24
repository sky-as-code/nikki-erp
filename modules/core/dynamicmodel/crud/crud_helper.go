package crud

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type ToDomainModelFunc[TDomain any] func(data dmodel.DynamicFields) TDomain
type BeforeValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain) (TDomain, error)
type AfterValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain any] func(ctx corectx.Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action         string
	BaseRepoGetter coredyn.BaseRepoGetter

	// Data to create
	Data dmodel.DynamicModelGetter

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
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param CreateParam[TDomain, TDomainPtr],
) (*crud.OpResult[TDomain], error) {

	baseRepo := param.BaseRepoGetter.GetBaseRepo()
	schema := baseRepo.GetSchema()
	fieldData := param.Data.GetFieldData()
	newModel := TDomainPtr(new(TDomain))
	newModel.SetFieldData(fieldData)

	flow := coredyn.StartValidationFlow()
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
			result, clientErrs := schema.Validate(fieldData, dmodel.ModelSchemaValidateOpts{AutoGenerateValues: true, StripReadOnly: true})
			if clientErrs != nil {
				*vErrs = clientErrs
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
		return &crud.OpResult[TDomain]{
			ClientErrors: clientErrs,
		}, nil
	}

	insertRes, err := baseRepo.Insert(ctx, fieldData)
	if err != nil {
		return nil, err
	}
	if len(insertRes.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: insertRes.ClientErrors}, nil
	}

	insertedModel := TDomainPtr(new(TDomain))
	insertedModel.SetFieldData(insertRes.Data)
	return &crud.OpResult[TDomain]{
		Data:    *insertedModel,
		IsEmpty: false,
	}, nil
}

type UpdateParam[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
] struct {
	Action           string
	DbRepoGetter     coredyn.BaseRepoGetter
	Data             dmodel.DynamicModelGetter
	BeforeValidation BeforeValidationFunc[TDomainPtr]
	AfterValidation  AfterValidationFunc[TDomainPtr]
	ValidateExtra    ValidateExtraFunc[TDomainPtr]
}

func Update[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
) (*crud.OpResult[TDomain], error) {
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(param.Data.GetFieldData())

	clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: clientErrs}, nil
	}

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	prevEtag, _ := model.GetFieldData()[basemodel.FieldEtag].(string)
	updatedRes, err := baseRepo.Update(ctx, model.GetFieldData(), prevEtag)
	if err != nil {
		return nil, err
	}
	if len(updatedRes.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: updatedRes.ClientErrors}, nil
	}

	model.SetFieldData(updatedRes.Data)
	return &crud.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func runUpdateValidationFlow[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
	model TDomainPtr,
) (ft.ClientErrors, error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	entitySchema := baseRepo.GetSchema()

	return coredyn.StartValidationFlow().
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
			result, clientErrs := entitySchema.Validate(model.GetFieldData(), dmodel.ModelSchemaValidateOpts{ForEdit: true, StripReadOnly: true})
			if clientErrs != nil {
				*vErrs = clientErrs
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
	ctx corectx.Context,
	entitySchema *dmodel.ModelSchema,
	baseRepo coredyn.BaseRepository,
	fieldData dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	primaryKeys := make(dmodel.DynamicFields)
	for _, key := range entitySchema.KeyColumns() {
		primaryKeys[key] = fieldData[key]
	}

	dbRes, err := baseRepo.GetOne(ctx, coredyn.GetOneParam{Filter: primaryKeys})
	if err != nil {
		return err
	}
	if len(dbRes.ClientErrors) > 0 {
		for _, item := range dbRes.ClientErrors {
			vErrs.Append(item)
		}
		return nil
	}
	if dbRes.IsEmpty {
		vErrs.Append(*ft.NewNotFoundError(entitySchema.Name()))
		return nil
	}
	dbRecord := dbRes.Data

	if _, hasEtag := entitySchema.Field("etag"); hasEtag && fieldData["etag"] != dbRecord["etag"] {
		vErrs.Append(*ft.NewEtagMismatchedError(entitySchema.Name()))
	}
	return nil
}

func GetOne[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	query GetOneQuery,
) (*crud.OpResult[TDomain], error) {
	querySchema := query.GetSchema()
	queryFields := query.GetFieldData()
	sanitizedFields, cErrs := querySchema.Validate(queryFields, dmodel.ModelSchemaValidateOpts{StripReadOnly: false})
	if cErrs.Count() > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}

	delete(sanitizedFields, basemodel.FieldIncludeArchived)
	delete(sanitizedFields, basemodel.FieldColumns)

	baseRepo := dbRepoGetter.GetBaseRepo()
	dbRes, err := baseRepo.GetOne(ctx, coredyn.GetOneParam{
		Filter:          sanitizedFields,
		Columns:         query.GetColumns(),
		IncludeArchived: query.GetIncludeArchived(),
	})
	if err != nil {
		return nil, err
	}
	if len(dbRes.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: dbRes.ClientErrors}, nil
	}
	if dbRes.IsEmpty {
		cErrs.Append(*ft.NewNotFoundError(baseRepo.GetSchema().Name()))
		return &crud.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}

	model := TDomainPtr(new(TDomain))
	model.SetFieldData(dbRes.Data)
	return &crud.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func Archive[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	keys dmodel.DynamicFields,
) (*crud.OpResult[TDomain], error) {
	baseRepo := dbRepoGetter.GetBaseRepo()
	archRes, err := baseRepo.Archive(ctx, keys)
	if err != nil {
		return nil, err
	}
	if len(archRes.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: archRes.ClientErrors}, nil
	}
	if archRes.IsEmpty {
		return &crud.OpResult[TDomain]{IsEmpty: true}, nil
	}
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(archRes.Data)
	return &crud.OpResult[TDomain]{Data: *model, IsEmpty: false}, nil
}

func Search[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	param coredyn.SearchParam,
) (*crud.OpResult[crud.PagedResult[TDomain]], error) {
	baseRepo := dbRepoGetter.GetBaseRepo()
	searchRes, err := baseRepo.Search(ctx, param)
	if err != nil {
		return nil, err
	}
	if len(searchRes.ClientErrors) > 0 {
		return &crud.OpResult[crud.PagedResult[TDomain]]{ClientErrors: searchRes.ClientErrors}, nil
	}
	paged := searchRes.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		model := TDomainPtr(new(TDomain))
		model.SetFieldData(record)
		items[i] = *model
	}
	out := crud.PagedResult[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}
	return &crud.OpResult[crud.PagedResult[TDomain]]{Data: out, IsEmpty: len(items) == 0}, nil
}

func validateUniques(ctx corectx.Context, data dmodel.DynamicFields, dbRepo coredyn.BaseRepository, vErrs *ft.ClientErrors) error {
	collRes, err := dbRepo.CheckUniqueCollisions(ctx, data)
	if err != nil {
		return err
	}
	if len(collRes.ClientErrors) > 0 {
		for _, item := range collRes.ClientErrors {
			vErrs.Append(item)
		}
		return nil
	}
	collidingKeys := collRes.Data

	if len(collidingKeys) > 0 {
		field := collidingKeys[0][0]
		vErrs.Append(*ft.NewBusinessViolation(
			field,
			"common.err_unique_constraint_violated",
			"unique constraint violated {{.uniques}}",
			map[string]any{"uniques": collidingKeys},
		))
	}
	return nil
}
