package crud

import (
	crud "github.com/sky-as-code/nikki-erp/common/crud"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	corerepo "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
)

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

	newModel.SetFieldData(fieldData)
	insertRes, err := corerepo.Insert[TDomain, TDomainPtr](ctx, baseRepo, newModel)
	if err != nil {
		return nil, err
	}
	if len(insertRes.ClientErrors) > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: insertRes.ClientErrors}, nil
	}

	return insertRes, nil
}

type DeleteOneParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Cmd          DeleteOneQuery
}

func DeleteOne(ctx corectx.Context, param DeleteOneParam) (result *crud.OpResult[crud.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	querySchema := deleteOneSchema()
	sanitizedFields, cErrs := querySchema.ValidateStruct(param.Cmd)

	if cErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: cErrs}, nil
	}

	cmd := *(sanitizedFields.(*DeleteOneQuery))
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	delResult, err := corerepo.DeleteOne(ctx, baseRepo, dmodel.DynamicFields{
		basemodel.FieldId: cmd.Id,
	})

	return delResult, err
}

func deleteOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.crud.delete_one_query",
		func() *dmodel.ModelSchemaBuilder {
			return DeleteOneQuerySchemaBuilder()
		},
	)
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
	querySchema := getOneSchema()
	// querySchema := param.Query.GetSchema()
	// queryFields := param.Query.GetFieldData()
	sanitized, cErrs := querySchema.ValidateStruct(param.Query)
	if cErrs.Count() > 0 {
		return &crud.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}
	sanitizedQuery := sanitized.(*GetOneQuery)

	// param.Query.DeleteFieldData(&sanitizedFields)

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	return corerepo.GetOne[TDomain, TDomainPtr](ctx, baseRepo, coredyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			basemodel.FieldId: sanitizedQuery.Id,
		},
		Columns: sanitizedQuery.Columns,
	})
}

func getOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.crud.get_one_query",
		func() *dmodel.ModelSchemaBuilder {
			return GetOneQuerySchemaBuilder()
		},
	)
}

func Search[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	query SearchQuery,
) (*crud.OpResult[crud.PagedResultData[TDomain]], error) {
	querySchema := searchSchema()
	sanitizedQuery, cErrs := querySchema.ValidateStruct(query)

	if cErrs.Count() > 0 {
		return &crud.OpResult[crud.PagedResultData[TDomain]]{ClientErrors: cErrs}, nil
	}

	query = *(sanitizedQuery.(*SearchQuery))
	baseRepo := dbRepoGetter.GetBaseRepo()
	return corerepo.Search[TDomain, TDomainPtr](ctx, baseRepo, coredyn.RepoSearchParam{
		Graph:   query.Graph,
		Columns: query.Columns,
		Page:    *query.Page,
		Size:    *query.Size,
	})
}

func searchSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.crud.search_query",
		func() *dmodel.ModelSchemaBuilder {
			return SearchQuerySchemaBuilder()
		},
	)
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

func SetIsArchived(
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	cmd SetIsArchivedCommand,
) (*crud.OpResult[crud.MutateResultData], error) {
	cmdSchema := setIsArchivedSchema()
	sanitizedCmd, cErrs := cmdSchema.ValidateStruct(cmd, true)

	if cErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: cErrs}, nil
	}

	cmd = *(sanitizedCmd.(*SetIsArchivedCommand))
	result, err := UpdateRegardless(ctx, UpdateRegardlessParam{
		Action:       "setIsArchived",
		DbRepoGetter: dbRepoGetter,
		Data: dmodel.DynamicFields{
			basemodel.FieldId:         *cmd.Id,
			basemodel.FieldEtag:       *cmd.Etag,
			basemodel.FieldIsArchived: *cmd.IsArchived,
		},
	})

	return result, err
}

func setIsArchivedSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.crud.set_archived_command",
		func() *dmodel.ModelSchemaBuilder {
			return SetArchivedCommandSchemaBuilder()
		},
	)
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

	isExisting, clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &crud.OpResult[crud.MutateResultData]{IsEmpty: true}, nil
	}

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	return corerepo.UpdateMutate(ctx, baseRepo, model.GetFieldData())
}

type UpdateRegardlessParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Data         dmodel.DynamicFields
}

// UpdateRegardless updates a record without validation, but it still checks for existence and etag matching.
func UpdateRegardless(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (*crud.OpResult[crud.MutateResultData], error) {

	isExisting, clientErrs, err := runUpdateRegardlessCheckingFlow(ctx, param)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &crud.OpResult[crud.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &crud.OpResult[crud.MutateResultData]{IsEmpty: true}, nil
	}

	baseRepo := param.DbRepoGetter.GetBaseRepo()
	return corerepo.UpdateMutate(ctx, baseRepo, param.Data)
}

func runUpdateValidationFlow[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
	model TDomainPtr,
) (bool, ft.ClientErrors, error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	schema := baseRepo.GetSchema()

	isExisting := false
	cErr, err := coredyn.StartValidationFlow().
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
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			existing, err := checkExistenceAndEtag(ctx, schema, baseRepo, model.GetFieldData(), vErrs)
			if err != nil {
				return err
			}
			isExisting = existing
			if !existing {
				stopFlow()
			}
			return nil
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

	return isExisting, cErr, err
}

func runUpdateRegardlessCheckingFlow(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (bool, ft.ClientErrors, error) {
	baseRepo := param.DbRepoGetter.GetBaseRepo()
	schema := baseRepo.GetSchema()

	isExisting := false
	cErr, err := coredyn.StartValidationFlow().
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			existing, err := checkExistenceAndEtag(ctx, schema, baseRepo, param.Data, vErrs)
			if err != nil {
				return err
			}
			isExisting = existing
			if !existing {
				stopFlow()
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, param.Data, baseRepo, vErrs)
			return nil
		}).
		End()

	return isExisting, cErr, err
}

func checkExistenceAndEtag(
	ctx corectx.Context,
	schema *dmodel.ModelSchema,
	baseRepo coredyn.BaseRepository,
	fieldData dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) (bool, error) {
	primaryKeys := make(dmodel.DynamicFields)
	for _, key := range schema.KeyColumns() {
		primaryKeys[key] = fieldData[key]
	}

	dbRes, err := baseRepo.GetOne(ctx, coredyn.RepoGetOneParam{Filter: primaryKeys})
	if err != nil {
		return false, err
	}
	if len(dbRes.ClientErrors) > 0 {
		for _, item := range dbRes.ClientErrors {
			vErrs.Append(item)
		}
		return false, nil
	}
	if dbRes.IsEmpty {
		return false, nil
	}
	dbRecord := dbRes.Data

	dbEtag, hasEtag := dbRecord[basemodel.FieldEtag]
	etagMatched := dbEtag == fieldData[basemodel.FieldEtag]
	if hasEtag && !etagMatched {
		vErrs.Append(*ft.NewEtagMismatchedError())
	}
	return true, nil
}
