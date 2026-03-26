package crud

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
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

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFunc[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidation AfterValidationFunc[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to built-in schema validation.
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
			result, clientErrs := schema.Validate(fieldData)
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

type DeleteEqualParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Cmd          dmodel.SchemaGetter
}

func DeleteEqual(ctx corectx.Context, param DeleteEqualParam) (result *crud.OpResult[crud.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	querySchema := param.Cmd.GetSchema()
	queryFields := param.Cmd.GetFieldData()

	sanitizedFields, cErrs := querySchema.Validate(queryFields)
	if cErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: cErrs}, nil
	}

	delResult, err := execDelete(param, ctx, sanitizedFields)
	return delResult, err
}

func execDelete(
	param DeleteEqualParam, ctx corectx.Context, sanitizedFields dmodel.DynamicFields,
) (*crud.OpResult[crud.MutateResultData], error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	delResult, err := baseRepo.Delete(ctx, sanitizedFields)

	if err != nil {
		return nil, err
	}
	if delResult.ClientErrors.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: delResult.ClientErrors}, nil
	}
	if delResult.IsEmpty {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewNotFoundError(baseRepo.GetSchema().Name()))
		return &crud.OpResult[crud.MutateResultData]{
			ClientErrors: cErrs,
		}, nil
	}

	result := &crud.OpResult[crud.MutateResultData]{
		Data: crud.MutateResultData{
			AffectedCount: delResult.Data,
			AffectedAt:    model.NewModelDateTime(),
		},
	}
	return result, nil
}

type GetOneParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Query        GetOneQuery
}

func GetOne[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param GetOneParam,
) (*crud.OpResult[TDomain], error) {
	querySchema := param.Query.GetSchema()
	queryFields := param.Query.GetFieldData()
	sanitizedFields, cErrs := querySchema.Validate(queryFields)
	if cErrs.Count() > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}

	param.Query.DeleteFieldData(&sanitizedFields)

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	dbFound, err := baseRepo.GetOne(ctx, coredyn.GetOneParam{
		Filter:          sanitizedFields,
		Columns:         param.Query.GetColumns(),
		IncludeArchived: param.Query.GetIncludeArchived(),
	})
	if err != nil {
		return nil, err
	}
	if len(dbFound.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: dbFound.ClientErrors}, nil
	}
	if dbFound.IsEmpty {
		cErrs.Append(*ft.NewNotFoundError(baseRepo.GetSchema().Name()))
		return &crud.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}

	model := TDomainPtr(new(TDomain))
	model.SetFieldData(dbFound.Data)
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
	query SearchQuery,
) (*crud.OpResult[crud.PagedResultData[TDomain]], error) {
	querySchema := query.GetSchema()
	sanitizedQuery, cErrs := querySchema.ValidateStruct(query)

	if cErrs.Count() > 0 {
		return &crud.OpResult[crud.PagedResultData[TDomain]]{ClientErrors: cErrs}, nil
	}

	query = *(sanitizedQuery.(*SearchQuery))
	baseRepo := dbRepoGetter.GetBaseRepo()
	dbFound, err := baseRepo.Search(ctx, coredyn.SearchParam{
		Graph:           query.Graph,
		Columns:         query.Columns,
		IncludeArchived: *query.IncludeArchived,
		Page:            *query.Page,
		Size:            *query.Size,
	})
	if err != nil {
		return nil, err
	}
	if len(dbFound.ClientErrors) > 0 {
		return &crud.OpResult[crud.PagedResultData[TDomain]]{ClientErrors: dbFound.ClientErrors}, nil
	}
	paged := dbFound.Data
	items := make([]TDomain, len(paged.Items))
	for i, record := range paged.Items {
		model := TDomainPtr(new(TDomain))
		model.SetFieldData(record)
		items[i] = *model
	}
	out := crud.PagedResultData[TDomain]{
		Items: items,
		Total: paged.Total,
		Page:  paged.Page,
		Size:  paged.Size,
	}
	return &crud.OpResult[crud.PagedResultData[TDomain]]{Data: out, IsEmpty: len(items) == 0}, nil
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
) (*crud.OpResult[crud.MutateResultData], error) {
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(param.Data.GetFieldData())

	clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	updatedRes, err := baseRepo.Update(ctx, model.GetFieldData())
	if err != nil {
		return nil, err
	}
	if len(updatedRes.ClientErrors) > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: updatedRes.ClientErrors}, nil
	}

	updatedAt, etag := tryGetAfterUpdate(updatedRes.Data)
	model.SetFieldData(updatedRes.Data)
	return &crud.OpResult[crud.MutateResultData]{
		Data: crud.MutateResultData{
			AffectedCount: 1,
			AffectedAt:    updatedAt,
			Etag:          etag,
		},
	}, nil
}

func tryGetAfterUpdate(data dmodel.DynamicFields) (updatedAt model.ModelDateTime, etag string) {
	upt, ok := data[basemodel.FieldUpdatedAt]
	if ok {
		updatedAt = upt.(model.ModelDateTime)
	}
	et, ok := data[basemodel.FieldEtag]
	if ok {
		etag = et.(string)
	}
	return
}

func runUpdateValidationFlow[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
	model TDomainPtr,
) (ft.ClientErrors, error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	schema := baseRepo.GetSchema()

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
			result, clientErrs := schema.Validate(model.GetFieldData(), true)
			if clientErrs != nil {
				*vErrs = clientErrs
			} else {
				model.SetFieldData(result)
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			return checkExistenceAndEtag(ctx, schema, baseRepo, model.GetFieldData(), vErrs)
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
	schema *dmodel.ModelSchema,
	baseRepo coredyn.BaseRepository,
	fieldData dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) error {
	primaryKeys := make(dmodel.DynamicFields)
	for _, key := range schema.KeyColumns() {
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
		vErrs.Append(*ft.NewNotFoundError(schema.Name()))
		return nil
	}
	dbRecord := dbRes.Data

	if _, hasEtag := schema.Field("etag"); hasEtag && fieldData["etag"] != dbRecord["etag"] {
		vErrs.Append(*ft.NewEtagMismatchedError(schema.Name()))
	}
	return nil
}
