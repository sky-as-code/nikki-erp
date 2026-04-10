package crud

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
)

type BeforeValidationFn[T any] func(ctx corectx.Context, model T, vErrs *ft.ClientErrors) (T, error)
type AfterValidationSuccessFn[T any] func(ctx corectx.Context, model T) (T, error)
type AfterDeleteValidationSuccessFn func(ctx corectx.Context) error
type CreateValidateExtraFn[T any] func(ctx corectx.Context, inputModel T, vErrs *ft.ClientErrors) error
type UpdateValidateExtraFn[T any] func(ctx corectx.Context, inputModel T, foundModel T, vErrs *ft.ClientErrors) error
type DeleteValidateExtraFn func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action         string
	BaseRepoGetter dyn.DynamicModelRepository

	// Data to create
	Data dmodel.DynamicModelGetter

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFn[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidationSuccess AfterValidationSuccessFn[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to built-in schema validation.
	ValidateExtra CreateValidateExtraFn[TDomainPtr]
}

func Create[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param CreateParam[TDomain, TDomainPtr],
) (_ *dyn.OpResult[TDomain], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	dynamicRepo := param.BaseRepoGetter.GetBaseRepo()
	schema := dynamicRepo.Schema()
	fieldData := param.Data.GetFieldData()
	newModel := TDomainPtr(new(TDomain))
	newModel.SetFieldData(fieldData)

	flow := dyn.StartValidationFlow()
	clientErrs, err := flow.
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, newModel, vErrs)
			if err == nil && vErrs.Count() == 0 && result != nil && result != newModel {
				fieldData = result.GetFieldData()
			}
			return errors.Wrap(err, "Create.BeforeValidation")
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
			err := validateUniques(ctx, fieldData, dynamicRepo, vErrs)
			return errors.Wrap(err, "Create.ValidateUniques")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			newModel.SetFieldData(fieldData)
			err := param.ValidateExtra(ctx, newModel, vErrs)
			return errors.Wrap(err, "Create.ValidateExtra")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidationSuccess != nil {
				result, err := param.AfterValidationSuccess(ctx, newModel)
				if err == nil && result != nil && result != newModel {
					fieldData = result.GetFieldData()
				}
				return errors.Wrap(err, "Create.AfterValidationSuccess")
			}
			return nil
		}).
		End()

	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[TDomain]{
			ClientErrors: clientErrs,
		}, nil
	}

	newModel.SetFieldData(fieldData)
	insRes, err := baserepo.Insert(ctx, dynamicRepo, newModel)
	if err != nil {
		return nil, errors.Wrap(err, "Create.Insert")
	}
	if insRes.ClientErrors.Count() > 0 {
		return &dyn.OpResult[TDomain]{ClientErrors: insRes.ClientErrors}, nil
	}

	return &dyn.OpResult[TDomain]{Data: *newModel, HasData: true}, nil
}

type CreateBulkParam[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
	TDomainGetter dmodel.DynamicModelGetter,
] struct {
	Action         string
	BaseRepoGetter dyn.DynamicModelRepository

	Data []TDomainGetter

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFn[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidationSuccess AfterValidationSuccessFn[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to built-in schema validation.
	ValidateExtra CreateValidateExtraFn[TDomainPtr]
}

func CreateBulk[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param CreateBulkParam[TDomain, TDomainPtr, TDomainPtr],
) (*dyn.OpResult[[]TDomain], error) {

	dynamicRepo := param.BaseRepoGetter.GetBaseRepo()
	schema := dynamicRepo.Schema()

	allClientErrs := ft.NewClientErrors()
	allNewModels := make([]TDomainPtr, 0, len(param.Data))

	for _, dmodel := range param.Data {
		fieldData := dmodel.GetFieldData()

		newModel := TDomainPtr(new(TDomain))
		newModel.SetFieldData(fieldData)

		flow := dyn.StartValidationFlow()
		clientErrs, err := flow.
			Step(func(vErrs *ft.ClientErrors) error {
				if param.BeforeValidation == nil {
					return nil
				}
				result, err := param.BeforeValidation(ctx, newModel, vErrs)
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
				validateUniques(ctx, fieldData, dynamicRepo, vErrs)
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
				if param.AfterValidationSuccess != nil {
					result, err := param.AfterValidationSuccess(ctx, newModel)
					if err == nil && result != nil && result != newModel {
						fieldData = result.GetFieldData()
					}
					return errors.Wrap(err, "Create.AfterValidationSuccess")
				}
				return nil
			}).
			End()

		if err != nil {
			return nil, err
		}

		if clientErrs != nil && clientErrs.Count() > 0 {
			for _, item := range clientErrs {
				allClientErrs.Append(item)
			}
		}

		newModel.SetFieldData(fieldData)
		allNewModels = append(allNewModels, newModel)
	}

	if allClientErrs.Count() > 0 {
		return &dyn.OpResult[[]TDomain]{
			ClientErrors: *allClientErrs,
			HasData:      false,
		}, nil
	}

	insRes, err := baserepo.InsertBulk(ctx, dynamicRepo, allNewModels)
	if err != nil {
		return nil, err
	}

	if insRes.ClientErrors.Count() > 0 {
		return &dyn.OpResult[[]TDomain]{ClientErrors: insRes.ClientErrors}, nil
	}

	data := make([]TDomain, 0, len(allNewModels))
	for _, model := range allNewModels {
		data = append(data, *model)
	}

	return &dyn.OpResult[[]TDomain]{Data: data, HasData: true}, nil
}

type UpdateBulkParam[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
	TDomainGetter dmodel.DynamicModelGetter,
] struct {
	Action         string
	BaseRepoGetter dyn.DynamicModelRepository
	Data           []TDomainGetter

	BeforeValidation       BeforeValidationFn[TDomainPtr]
	AfterValidationSuccess AfterValidationSuccessFn[TDomainPtr]
	ValidateExtra          UpdateValidateExtraFn[TDomainPtr]
}

func UpdateBulk[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param UpdateBulkParam[TDomain, TDomainPtr, TDomainPtr],
) (*dyn.OpResult[dyn.MutateResultData], error) {

	dynamicRepo := param.BaseRepoGetter.GetBaseRepo()
	allClientErrs := ft.NewClientErrors()
	allModels := make([]TDomainPtr, 0, len(param.Data))
	anyNotExisting := false

	for _, dmodelItem := range param.Data {
		model := TDomainPtr(new(TDomain))
		model.SetFieldData(dmodelItem.GetFieldData())

		isExisting, clientErrs, err := runUpdateValidationFlow(ctx, UpdateParam[TDomain, TDomainPtr]{
			Action:                 param.Action,
			DbRepoGetter:           param.BaseRepoGetter,
			BeforeValidation:       param.BeforeValidation,
			AfterValidationSuccess: param.AfterValidationSuccess,
			ValidateExtra:          param.ValidateExtra,
		}, model)
		if err != nil {
			return nil, err
		}

		if clientErrs != nil && clientErrs.Count() > 0 {
			for _, item := range clientErrs {
				allClientErrs.Append(item)
			}
		}

		if !isExisting {
			anyNotExisting = true
			continue
		}

		allModels = append(allModels, model)
	}

	if allClientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{
			ClientErrors: *allClientErrs,
			HasData:      false,
		}, nil
	}

	if anyNotExisting {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	var totalAffected int
	var lastAt model.ModelDateTime
	var lastEtag model.Etag

	for _, m := range allModels {
		updRes, err := baserepo.Update(ctx, dynamicRepo, m.GetFieldData())
		if err != nil {
			return nil, err
		}
		if updRes.ClientErrors.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: updRes.ClientErrors}, nil
		}
		if updRes.HasData {
			totalAffected += updRes.Data.AffectedCount
			lastAt = updRes.Data.AffectedAt
			lastEtag = updRes.Data.Etag
		}
	}

	return &dyn.OpResult[dyn.MutateResultData]{
		Data: dyn.MutateResultData{
			AffectedCount: totalAffected,
			AffectedAt:    lastAt,
			Etag:          lastEtag,
		},
		HasData: true,
	}, nil
}

type DeleteOneParam struct {
	Action                 string
	DbRepoGetter           dyn.DynamicModelRepository
	Cmd                    dyn.DeleteOneCommand
	ValidateExtra          DeleteValidateExtraFn
	AfterValidationSuccess AfterDeleteValidationSuccessFn
}

func DeleteOne(ctx corectx.Context, param DeleteOneParam) (result *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := param.Cmd.GetSchema().ValidateStruct(param.Cmd)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}
	cmd := *(sanitized.(*dyn.DeleteOneCommand))

	if param.ValidateExtra != nil {
		err := param.ValidateExtra(ctx, dmodel.DynamicFields{basemodel.FieldId: cmd.Id}, &cErrs)
		if err != nil {
			return nil, errors.Wrap(err, "DeleteOne.ValidateExtra")
		}
		if cErrs.Count() > 0 {
			return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
		}
	}
	if param.AfterValidationSuccess != nil {
		err := param.AfterValidationSuccess(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "DeleteOne.AfterValidationSuccess")
		}
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	delResult, err := baserepo.DeleteOne(ctx, dynamicRepo, dmodel.DynamicFields{
		basemodel.FieldId: cmd.Id,
	})

	return delResult, errors.Wrap(err, "DeleteOne")
}

type ExistsParam struct {
	Action       string
	DbRepoGetter dyn.DynamicModelRepository
	Query        dyn.ExistsQuery
}

func Exists(ctx corectx.Context, param ExistsParam) (result *dyn.OpResult[dyn.ExistsResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := existsQuerySchema().ValidateStruct(param.Query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: cErrs}, nil
	}
	cmd := *(sanitized.(*dyn.ExistsQuery))
	keys := existsQueryToKeyMaps(cmd)
	repoOut, err := baserepo.Exists(ctx, param.DbRepoGetter.GetBaseRepo(), keys)
	if err != nil {
		return nil, errors.Wrap(err, "Exists")
	}
	if len(repoOut.ClientErrors) > 0 {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: repoOut.ClientErrors}, nil
	}
	data := existsRepoResultToData(repoOut.Data)
	return &dyn.OpResult[dyn.ExistsResultData]{Data: data, HasData: true}, nil
}

func existsQueryToKeyMaps(query dyn.ExistsQuery) []dmodel.DynamicFields {
	keys := array.Map(query.Ids, func(id model.Id) dmodel.DynamicFields {
		return dmodel.DynamicFields{basemodel.FieldId: id}
	})
	return keys
}

func existsRepoResultToData(repo dyn.RepoExistsResult) dyn.ExistsResultData {
	out := dyn.ExistsResultData{
		Existing:    make([]model.Id, 0),
		NotExisting: make([]model.Id, 0),
	}
	out.Existing = array.Map(repo.Existing, func(fields dmodel.DynamicFields) model.Id {
		return *fields.GetModelId(basemodel.FieldId)
	})
	out.NotExisting = array.Map(repo.NotExisting, func(fields dmodel.DynamicFields) model.Id {
		return *fields.GetModelId(basemodel.FieldId)
	})
	return out
}

func existsQuerySchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.exists_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.ExistsQuerySchemaBuilder()
		},
	)
}

func ExecInTranx[TResult any](ctx corectx.Context, repo dyn.DynamicModelRepository, fn func(ctx corectx.Context) (*TResult, error)) (result *TResult, err error) {
	var tranx database.DbTransaction
	tranx, err = setNewDbTranx(ctx, repo)
	if err != nil {
		return nil, err
	}

	defer func() {
		if e := ft.RecoverPanic(recover(), "ExecInTranx"); e != nil {
			err = e
		}
		clearDbTranx(ctx)
		if err != nil {
			rbErr := tranx.Rollback()
			if rbErr != nil {
				err = stdErr.Join(err, rbErr)
			}
		} else {
			err = tranx.Commit()
		}
	}()

	result, err = fn(ctx)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func setNewDbTranx(ctx corectx.Context, repo dyn.DynamicModelRepository) (database.DbTransaction, error) {
	trx, err := repo.BeginTransaction(ctx)
	if err != nil {
		return nil, err
	}

	ctx.SetDbTranx(trx)
	return trx, nil
}

func clearDbTranx(ctx corectx.Context) {
	ctx.SetDbTranx(nil)
}

type GetOneParam struct {
	Action       string
	DbRepoGetter dyn.DynamicModelRepository
	Query        dyn.GetOneQuery
}

func GetOne[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param GetOneParam,
) (_ *dyn.OpResult[TDomain], err error) {
	result, err := getOneWithArchived[TDomain, TDomainPtr](ctx, param, nil)
	return result, errors.Wrap(err, "GetOne")
}

func GetOneEnabled[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param GetOneParam,
) (_ *dyn.OpResult[TDomain], err error) {
	result, err := getOneWithArchived[TDomain, TDomainPtr](ctx, param, util.ToPtr(false))
	return result, errors.Wrap(err, "GetOneEnabled")
}

func getOneWithArchived[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param GetOneParam, isArchived *bool,
) (_ *dyn.OpResult[TDomain], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	querySchema := getOneSchema()
	sanitized, cErrs := querySchema.ValidateStruct(param.Query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}
	sanitizedQuery := sanitized.(*dyn.GetOneQuery)

	filter := dmodel.DynamicFields{
		basemodel.FieldId: sanitizedQuery.Id,
	}
	if isArchived != nil {
		filter[basemodel.FieldIsArchived] = *isArchived
	}
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	result, err := baserepo.GetOne[TDomain, TDomainPtr](ctx, dynamicRepo, dyn.RepoGetOneParam{
		Filter:  filter,
		Columns: sanitizedQuery.Columns,
	})
	return result, err
}

func getOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.get_one_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.GetOneQuerySchemaBuilder()
		},
	)
}

type SearchParam struct {
	Action       string
	DbRepoGetter dyn.DynamicModelRepository
	Query        dyn.SearchQuery
}

func Search[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, param SearchParam,
) (_ *dyn.OpResult[dyn.PagedResultData[TDomain]], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	querySchema := searchSchema()
	sanitized, cErrs := querySchema.ValidateStruct(param.Query)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[TDomain]]{ClientErrors: cErrs}, nil
	}

	sanitizedQuery := *(sanitized.(*dyn.SearchQuery))
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	result, err := baserepo.Search[TDomain, TDomainPtr](ctx, dynamicRepo, dyn.RepoSearchParam{
		Columns:  sanitizedQuery.Columns,
		Page:     sanitizedQuery.Page,
		Size:     sanitizedQuery.Size,
		Graph:    sanitizedQuery.Graph,
		Language: sanitizedQuery.Language,
	})
	return result, errors.Wrap(err, "Search")
}

func searchSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.search_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.SearchQuerySchemaBuilder()
		},
	)
}

func validateUniques(ctx corectx.Context, data dmodel.DynamicFields, dbRepo dyn.BaseDynamicRepository, vErrs *ft.ClientErrors) error {
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
		vErrs.Append(*ft.NewAnonymousBusinessViolation(
			"common.err_unique_constraint_violated",
			"unique constraint violated {{.uniques}}",
			map[string]any{"uniques": collidingKeys},
		))
	}
	return nil
}

func SetIsArchived(
	ctx corectx.Context,
	dbRepoGetter dyn.DynamicModelRepository,
	cmd dyn.SetIsArchivedCommand,
) (_ *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "SetIsArchived"); e != nil {
			err = e
		}
	}()

	cmdSchema := setIsArchivedSchema()
	sanitizedCmd, cErrs := cmdSchema.ValidateStruct(cmd, true)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	cmd = *(sanitizedCmd.(*dyn.SetIsArchivedCommand))
	result, err := UpdateRegardless(ctx, UpdateRegardlessParam{
		Action:       "setIsArchived",
		DbRepoGetter: dbRepoGetter,
		Data: dmodel.DynamicFields{
			basemodel.FieldId:         cmd.Id,
			basemodel.FieldEtag:       cmd.Etag,
			basemodel.FieldIsArchived: cmd.IsArchived,
		},
	})

	return result, errors.Wrap(err, "SetIsArchived")
}

func setIsArchivedSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.set_archived_command",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.SetArchivedCommandSchemaBuilder()
		},
	)
}

type UpdateParam[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
] struct {
	Action                 string
	DbRepoGetter           dyn.DynamicModelRepository
	Data                   dmodel.DynamicModelGetter
	BeforeValidation       BeforeValidationFn[TDomainPtr]
	AfterValidationSuccess AfterValidationSuccessFn[TDomainPtr]
	ValidateExtra          UpdateValidateExtraFn[TDomainPtr]
}

func Update[
	TDomain any,
	TDomainPtr dyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
) (_ *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	model := TDomainPtr(new(TDomain))
	model.SetFieldData(param.Data.GetFieldData())

	isExisting, clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	result, err := baserepo.Update(ctx, dynamicRepo, model.GetFieldData())
	return result, errors.Wrap(err, "Update")
}

type UpdateRegardlessParam struct {
	Action       string
	DbRepoGetter dyn.DynamicModelRepository
	Data         dmodel.DynamicFields
}

// UpdateRegardless updates a record without validation, but it still checks for existence and etag matching.
// Use this function with caution. Must do your own validation before calling this function.
func UpdateRegardless(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (_ *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	isExisting, _, clientErrs, err := runUpdateRegardlessCheckingFlow(ctx, param)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	return baserepo.Update(ctx, dynamicRepo, param.Data)
}

func runUpdateValidationFlow[TDomain any, TDomainPtr dyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
	inputModel TDomainPtr,
) (bool, ft.ClientErrors, error) {
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	schema := dynamicRepo.Schema()

	foundModel := TDomainPtr(new(TDomain))
	isExisting := false
	cErr, err := dyn.StartValidationFlow().
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, inputModel, vErrs)
			if err == nil && vErrs.Count() == 0 {
				inputModel.SetFieldData(result.GetFieldData())
			}
			return errors.Wrap(err, "Update.BeforeValidation")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := schema.Validate(inputModel.GetFieldData(), true)
			if clientErrs != nil {
				*vErrs = clientErrs
			} else {
				inputModel.SetFieldData(result)
			}
			return nil
		}).
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			existing, dbRecord, err := checkExistenceAndEtag(ctx, schema, dynamicRepo, inputModel.GetFieldData(), vErrs)
			if err != nil {
				return errors.Wrap(err, "Update.CheckExistenceAndEtag")
			}
			isExisting = existing
			if !existing {
				stopFlow()
			}
			foundModel.SetFieldData(dbRecord)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			err := validateUniques(ctx, inputModel.GetFieldData(), dynamicRepo, vErrs)
			return errors.Wrap(err, "Update.ValidateUniques")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			err := param.ValidateExtra(ctx, inputModel, foundModel, vErrs)
			return errors.Wrap(err, "Update.ValidateExtra")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidationSuccess == nil {
				return nil
			}
			result, err := param.AfterValidationSuccess(ctx, inputModel)
			if err == nil {
				inputModel.SetFieldData(result.GetFieldData())
			}
			return errors.Wrap(err, "Update.AfterValidationSuccess")
		}).
		End()

	return isExisting, cErr, err
}

func runUpdateRegardlessCheckingFlow(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (isExisting bool, dbRecord dmodel.DynamicFields, clientErrs ft.ClientErrors, err error) {
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	schema := dynamicRepo.Schema()

	isExisting = false
	cErr, err := dyn.StartValidationFlow().
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			isExisting, dbRecord, err = checkExistenceAndEtag(ctx, schema, dynamicRepo, param.Data, vErrs)
			if err != nil {
				return errors.Wrap(err, "UpdateRegardless.CheckExistenceAndEtag")
			}
			if !isExisting {
				stopFlow()
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			err := validateUniques(ctx, param.Data, dynamicRepo, vErrs)
			return errors.Wrap(err, "UpdateRegardless.ValidateUniques")
		}).
		End()

	return isExisting, dbRecord, cErr, err
}

type ManageM2mParam struct {
	Action             string
	DbRepoGetter       dyn.DynamicModelRepository
	DestSchemaName     string
	SrcId              model.Id
	SrcIdFieldForError string
	AssociatedIds      datastructure.Set[model.Id]
	DisassociatedIds   datastructure.Set[model.Id]
	BeforeInsert       func(ctx corectx.Context, dbRecords []dmodel.DynamicFields) error
}

func ManageM2m(ctx corectx.Context, param ManageM2mParam) (
	result *dyn.OpResult[dyn.MutateResultData], err error,
) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()
	_, cErrs := manageAssocsSchema().Validate(dmodel.DynamicFields{
		basemodel.FieldId:           param.SrcId,
		basemodel.FieldAssociations: param.AssociatedIds.ToSlice(),
		basemodel.FieldDesociations: param.DisassociatedIds.ToSlice(),
	})
	if cErrs.Count() > 0 {
		cErrs.RenameField(basemodel.FieldId, param.SrcIdFieldForError)
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	repoOut, err := baserepo.ManageM2m(
		ctx, dynamicRepo, dyn.RepoManageM2mParam{
			DestSchemaName:     param.DestSchemaName,
			SrcId:              param.SrcId,
			SrcIdFieldForError: param.SrcIdFieldForError,
			AssociatedIds:      param.AssociatedIds,
			DisassociatedIds:   param.DisassociatedIds,
			BeforeInsert:       param.BeforeInsert,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "ManageM2m")
	}
	if repoOut.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: repoOut.ClientErrors}, nil
	}
	data := dyn.MutateResultData{
		AffectedCount: repoOut.Data,
		AffectedAt:    model.NewModelDateTime(),
	}
	return &dyn.OpResult[dyn.MutateResultData]{Data: data, HasData: true}, nil
}

func manageAssocsSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.manage_assocs_command",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.ManageAssocsSchemaBuilder()
		},
	)
}

func checkExistenceAndEtag(
	ctx corectx.Context,
	schema *dmodel.ModelSchema,
	dynamicRepo dyn.BaseDynamicRepository,
	fieldData dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) (bool, dmodel.DynamicFields, error) {
	primaryKeys := make(dmodel.DynamicFields)
	for _, key := range schema.KeyColumns() {
		primaryKeys[key] = fieldData[key]
	}

	dbRes, err := dynamicRepo.GetOne(ctx, dyn.RepoGetOneParam{Filter: primaryKeys})
	if err != nil {
		return false, nil, err
	}
	if len(dbRes.ClientErrors) > 0 {
		for _, item := range dbRes.ClientErrors {
			vErrs.Append(item)
		}
		return false, nil, nil
	}
	if !dbRes.HasData {
		return false, nil, nil
	}
	dbRecord := dbRes.Data

	dbEtag, hasEtag := dbRecord[basemodel.FieldEtag]
	etagMatched := dbEtag == fieldData[basemodel.FieldEtag]
	if hasEtag && !etagMatched {
		vErrs.Append(*ft.NewEtagMismatchedError())
	}
	return true, dbRecord, nil
}
